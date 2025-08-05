package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/grafana/grafana-app-sdk/logging"
	"github.com/urfave/cli/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	k8srest "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"

	provisioning "github.com/grafana/grafana/apps/provisioning/pkg/apis/provisioning/v0alpha1"
	client "github.com/grafana/grafana/apps/provisioning/pkg/generated/clientset/versioned"
	typedclient "github.com/grafana/grafana/apps/provisioning/pkg/generated/clientset/versioned/typed/provisioning/v0alpha1"
	informer "github.com/grafana/grafana/apps/provisioning/pkg/generated/informers/externalversions"
	informerv0alpha1 "github.com/grafana/grafana/apps/provisioning/pkg/generated/informers/externalversions/provisioning/v0alpha1"
	listers "github.com/grafana/grafana/apps/provisioning/pkg/generated/listers/provisioning/v0alpha1"
)

var (
	kubeconfig = flag.String("kubeconfig", "", "Path to kubeconfig file")
	namespace  = flag.String("namespace", "default", "Namespace to watch")
	workers    = flag.Int("workers", 2, "Number of worker goroutines")
)

func main() {
	fmt.Println("Starting provisioning controller...")

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
				Name:        "namespace",
				Usage:       "Namespace to watch",
				Value:       "default",
				Destination: namespace,
			},
			&cli.IntFlag{
				Name:        "workers",
				Usage:       "Number of worker goroutines",
				Value:       2,
				Destination: workers,
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

func (c *SimpleRepositoryController) Run(ctx context.Context, workerCount int) {
	defer c.queue.ShutDown()

	fmt.Println("Starting SimpleRepositoryController")
	defer fmt.Println("Shutting down SimpleRepositoryController")

	if !cache.WaitForCacheSync(ctx.Done(), c.repoSynced) {
		c.logger.Error("Failed to sync informer cache")
		return
	}

	fmt.Println("Cache synced successfully, starting workers", "count", workerCount)
	for i := 0; i < workerCount; i++ {
		workerID := i
		fmt.Println("Starting worker", "worker_id", workerID)
		go func() {
			wait.UntilWithContext(ctx, c.runWorker, time.Second)
			fmt.Println("Worker stopped", "worker_id", workerID)
		}()
	}

	fmt.Println("All workers started, waiting for events")
	<-ctx.Done()
	fmt.Println("Shutting down workers")
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
	fmt.Println("About to load kubeconfig...")
	var config *k8srest.Config
	var err error

	if *kubeconfig != "" {
		fmt.Println("Loading kubeconfig from file: %s\n", *kubeconfig)
		config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
		fmt.Println("BuildConfigFromFlags completed")
	} else {
		fmt.Println("Loading in-cluster config")
		config, err = k8srest.InClusterConfig()
		fmt.Println("InClusterConfig completed")
	}
	if err != nil {
		fmt.Println("Error loading kubeconfig: %v\n", err)
		return fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	fmt.Println("Successfully loaded kubeconfig, host: %s\n", config.Host)

	// Create clients
	fmt.Println("Creating Kubernetes client")
	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}
	fmt.Println("Kubernetes client created successfully", "client_type", fmt.Sprintf("%T", k8sClient))

	fmt.Println("Creating provisioning client")
	provisioningClient, err := client.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create provisioning client: %w", err)
	}
	fmt.Println("Provisioning client created successfully")

	fmt.Println("Successfully created all clients")

	// Create informer factory
	fmt.Println("Creating informer factory", "namespace", *namespace, "resync_period", "10m")
	informerFactory := informer.NewSharedInformerFactoryWithOptions(
		provisioningClient,
		10*time.Minute, // resync period
		informer.WithNamespace(*namespace),
	)
	fmt.Println("Informer factory created successfully")

	// Create repository informer
	fmt.Println("Creating repository informer")
	repoInformer := informerFactory.Provisioning().V0alpha1().Repositories()
	fmt.Println("Repository informer created successfully", "informer_type", fmt.Sprintf("%T", repoInformer))

	// Create simple controller
	fmt.Println("Creating SimpleRepositoryController")
	controller := NewSimpleRepositoryController(
		provisioningClient.ProvisioningV0alpha1(),
		repoInformer,
	)
	fmt.Println("SimpleRepositoryController created successfully")

	// Start informer factory
	fmt.Println("Starting informer factory")
	informerFactory.Start(context.Background().Done())
	fmt.Println("Informer factory started")

	// Wait for cache sync
	fmt.Println("Waiting for cache sync")
	if !cache.WaitForCacheSync(context.Background().Done(), repoInformer.Informer().HasSynced) {
		return fmt.Errorf("failed to sync informer cache")
	}
	fmt.Println("Cache synced successfully")

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("Received shutdown signal, stopping controller")
		cancel()
	}()

	// Run controller
	fmt.Println("Starting simple repository controller", "workers", *workers)
	controller.Run(ctx, *workers)

	fmt.Println("Controller stopped")
	return nil
}
