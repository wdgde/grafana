package app

import (
	"context"

	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"

	"github.com/grafana/grafana-app-sdk/app"
	"github.com/grafana/grafana-app-sdk/k8s"
	"github.com/grafana/grafana-app-sdk/operator"
	"github.com/grafana/grafana-app-sdk/resource"
	"github.com/grafana/grafana-app-sdk/simple"
	collectionv0alpha1 "github.com/grafana/grafana/apps/collection/pkg/apis/collection/v0alpha1"
)

type CollectionConfig struct {
	EnableReconcilers bool
}

func getPatchClient(restConfig rest.Config, collectionKind resource.Kind) (operator.PatchClient, error) {
	clientGenerator := k8s.NewClientRegistry(restConfig, k8s.ClientConfig{})
	return clientGenerator.ClientFor(collectionKind)
}

func New(cfg app.Config) (app.App, error) {
	simpleConfig := simple.AppConfig{
		Name:       "collection",
		KubeConfig: cfg.KubeConfig,
		InformerConfig: simple.AppInformerConfig{
			ErrorHandler: func(ctx context.Context, err error) {
				klog.ErrorS(err, "Informer processing error")
			},
		},
		ManagedKinds: []simple.AppManagedKind{
			{
				Kind: collectionv0alpha1.StarsKind(),
				Validator: &simple.Validator{
					ValidateFunc: func(ctx context.Context, req *app.AdmissionRequest) error {
						// do something here if needed

						return nil
					},
				},
			},
		},
	}

	a, err := simple.NewApp(simpleConfig)
	if err != nil {
		return nil, err
	}

	err = a.ValidateManifest(cfg.ManifestData)
	if err != nil {
		return nil, err
	}

	return a, nil
}
