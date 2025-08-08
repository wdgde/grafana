package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/grafana/authlib/authn"
	"github.com/grafana/grafana-app-sdk/logging"
	"github.com/urfave/cli/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilnet "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	k8srest "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/transport"
	"k8s.io/client-go/util/workqueue"

	provisioning "github.com/grafana/grafana/apps/provisioning/pkg/apis/provisioning/v0alpha1"
	client "github.com/grafana/grafana/apps/provisioning/pkg/generated/clientset/versioned"
	typedclient "github.com/grafana/grafana/apps/provisioning/pkg/generated/clientset/versioned/typed/provisioning/v0alpha1"
	informer "github.com/grafana/grafana/apps/provisioning/pkg/generated/informers/externalversions"
	informerv0alpha1 "github.com/grafana/grafana/apps/provisioning/pkg/generated/informers/externalversions/provisioning/v0alpha1"
	listers "github.com/grafana/grafana/apps/provisioning/pkg/generated/listers/provisioning/v0alpha1"
)

var (
	kubeconfig            = flag.String("kubeconfig", "", "Path to kubeconfig file")
	token                 = flag.String("token", "", "Token to use for authentication")
	tokenExchangeURL      = flag.String("token-exchange-url", "", "Token exchange URL")
	provisioningServerURL = flag.String("provisioning-server-url", "", "Provisioning server URL")
)

func main() {
	app := &cli.App{
		Name:  "provisioning-controller",
		Usage: "Watch repositories and manage provisioning resources",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "kubeconfig",
				Usage:       "Path to kubeconfig file",
				Value:       "",
				Destination: kubeconfig,
			},
			&cli.StringFlag{
				Name:        "token",
				Usage:       "Token to use for authentication",
				Value:       "",
				Destination: token,
			},
			&cli.StringFlag{
				Name:        "token-exchange-url",
				Usage:       "Token exchange URL",
				Value:       "",
				Destination: tokenExchangeURL,
			},
			&cli.StringFlag{
				Name:        "provisioning-server-url",
				Usage:       "Provisioning server URL",
				Value:       "",
				Destination: provisioningServerURL,
			},
		},
		Action: runSimpleController,
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// SimpleRepositoryController is a simplified version of the repository controller
type SimpleRepositoryController struct {
	client     typedclient.ProvisioningV0alpha1Interface
	repoLister listers.RepositoryLister
	repoSynced cache.InformerSynced
	logger     logging.Logger
	queue      workqueue.RateLimitingInterface
}

func NewSimpleRepositoryController(
	provisioningClient typedclient.ProvisioningV0alpha1Interface,
	repoInformer informerv0alpha1.RepositoryInformer,
) *SimpleRepositoryController {
	controller := &SimpleRepositoryController{
		client:     provisioningClient,
		repoLister: repoInformer.Lister(),
		repoSynced: repoInformer.Informer().HasSynced,
		logger:     logging.DefaultLogger.With("logger", "simple-provisioning-controller"),
		queue:      workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
	}

	// Add event handlers
	repoInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueue,
		UpdateFunc: func(oldObj, newObj interface{}) {
			controller.enqueue(newObj)
		},
		DeleteFunc: controller.enqueue,
	})

	return controller
}

func (c *SimpleRepositoryController) enqueue(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		c.logger.Error("Couldn't get key for object", "error", err)
		return
	}

	// Log the event type and key
	eventType := "unknown"
	switch obj.(type) {
	case *provisioning.Repository:
		repo := obj.(*provisioning.Repository)
		if repo.DeletionTimestamp != nil {
			eventType = "delete"
		} else {
			eventType = "add/update"
		}
		fmt.Println("Received repository event",
			"event_type", eventType,
			"key", key,
			"namespace", repo.Namespace,
			"name", repo.Name,
			"generation", repo.Generation)
	}

	c.queue.Add(key)
}

func (c *SimpleRepositoryController) Run(ctx context.Context) {
	defer c.queue.ShutDown()

	if !cache.WaitForCacheSync(ctx.Done(), c.repoSynced) {
		c.logger.Error("Failed to sync informer cache")
		return
	}

	go func() {
		wait.UntilWithContext(ctx, c.runWorker, time.Second)
		fmt.Println("Worker stopped")
	}()

	<-ctx.Done()
}

func (c *SimpleRepositoryController) runWorker(ctx context.Context) {
	for c.processNextWorkItem(ctx) {
	}
}

func (c *SimpleRepositoryController) processNextWorkItem(ctx context.Context) bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(key)

	logger := c.logger.With("key", key)
	fmt.Println("Processing work item from queue")

	err := c.processRepository(ctx, key.(string))
	if err == nil {
		c.queue.Forget(key)
		fmt.Println("Successfully processed work item")
		return true
	}

	logger.Error("Failed to process repository", "error", err)
	c.queue.AddRateLimited(key)
	return true
}

func (c *SimpleRepositoryController) processRepository(ctx context.Context, key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	repo, err := c.repoLister.Repositories(namespace).Get(name)
	if err != nil {
		return err
	}

	fmt.Println("Processing repository",
		"namespace", repo.Namespace,
		"name", repo.Name,
		"type", repo.Spec.Type,
		"generation", repo.Generation,
		"observedGeneration", repo.Status.ObservedGeneration)

	// Update observed generation
	if repo.Generation != repo.Status.ObservedGeneration {
		repo.Status.ObservedGeneration = repo.Generation

		_, err = c.client.Repositories(repo.Namespace).UpdateStatus(ctx, repo, metav1.UpdateOptions{})
		if err != nil {
			return err
		}

		fmt.Println("Updated repository status",
			"namespace", repo.Namespace,
			"name", repo.Name,
			"observedGeneration", repo.Status.ObservedGeneration)
	}

	return nil
}

func runSimpleController(c *cli.Context) error {
	tokenExchangeClient, err := authn.NewTokenExchangeClient(authn.TokenExchangeConfig{
		TokenExchangeURL: *tokenExchangeURL,
		Token:            *token,
	})
	if err != nil {
		return fmt.Errorf("failed to create token exchange client: %w", err)
	}

	config := &k8srest.Config{
		APIPath: "/apis",
		Host:    *provisioningServerURL,
		WrapTransport: transport.WrapperFunc(func(rt http.RoundTripper) http.RoundTripper {
			return &authRoundTripper{
				tokenExchangeClient: tokenExchangeClient,
				transport:           rt,
			}
		}),
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: true,
		},
	}

	provisioningClient, err := client.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create provisioning client: %w", err)
	}

	// TODO: make this configurable
	informerFactory := informer.NewSharedInformerFactoryWithOptions(
		provisioningClient,
		30*time.Second, // resync period
	)

	repoInformer := informerFactory.Provisioning().V0alpha1().Repositories()
	controller := NewSimpleRepositoryController(
		provisioningClient.ProvisioningV0alpha1(),
		repoInformer,
	)
	informerFactory.Start(context.Background().Done())
	if !cache.WaitForCacheSync(context.Background().Done(), repoInformer.Informer().HasSynced) {
		return fmt.Errorf("failed to sync informer cache")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("Received shutdown signal, stopping controller")
		cancel()
	}()

	controller.Run(ctx)
	return nil
}

type authRoundTripper struct {
	tokenExchangeClient *authn.TokenExchangeClient
	transport           http.RoundTripper
}

func (t *authRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	tokenResponse, err := t.tokenExchangeClient.Exchange(req.Context(), authn.TokenExchangeRequest{
		Audiences: []string{"provisioning.grafana.app"},
		Namespace: "*",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to exchange token: %w", err)
	}

	// clone the request as RTs are not expected to mutate the passed request
	req = utilnet.CloneRequest(req)

	req.Header.Set("X-Access-Token", "Bearer "+tokenResponse.Token)
	return t.transport.RoundTrip(req)
}
