package app

import (
	"context"
	"fmt"

	"github.com/grafana/grafana-app-sdk/app"
	"github.com/grafana/grafana-app-sdk/k8s"
	"github.com/grafana/grafana-app-sdk/plugin/router"
	"github.com/grafana/grafana-app-sdk/resource"
	"github.com/grafana/grafana-app-sdk/simple"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"

	investigationsv0alpha1 "github.com/grafana/grafana/apps/investigations/pkg/apis/investigations/v0alpha1"
)

// InvestigationApp represents the investigations app with both REST and GraphQL support
type InvestigationApp struct {
	app.App
	graphqlProvider *router.AppGraphQLProvider
}

func New(cfg app.Config) (app.App, error) {
	var err error
	simpleConfig := simple.AppConfig{
		Name:       "investigation",
		KubeConfig: cfg.KubeConfig,
		InformerConfig: simple.AppInformerConfig{
			ErrorHandler: func(_ context.Context, err error) {
				klog.ErrorS(err, "Informer processing error")
			},
		},
		ManagedKinds: []simple.AppManagedKind{
			{
				Kind: investigationsv0alpha1.InvestigationKind(),
			},
			{
				Kind: investigationsv0alpha1.InvestigationIndexKind(),
			},
		},
	}

	baseApp, err := simple.NewApp(simpleConfig)
	if err != nil {
		return nil, err
	}

	err = baseApp.ValidateManifest(cfg.ManifestData)
	if err != nil {
		return nil, err
	}

	// Create resource collection for GraphQL
	resourceCollection := CreateInvestigationCollection()

	// Create client generator from KubeConfig
	clientGenerator := k8s.NewClientRegistry(cfg.KubeConfig, k8s.ClientConfig{})

	// Create store using the client generator
	store := resource.NewStore(clientGenerator, resourceCollection)

	// Create GraphQL provider
	graphqlProvider, err := router.NewAppGraphQLProvider("investigations", resourceCollection, store)
	if err != nil {
		return nil, fmt.Errorf("failed to create GraphQL provider: %w", err)
	}

	return &InvestigationApp{
		App:             baseApp,
		graphqlProvider: graphqlProvider,
	}, nil
}

// GetGraphQLProvider returns the GraphQL provider for this app
func (a *InvestigationApp) GetGraphQLProvider() *router.AppGraphQLProvider {
	return a.graphqlProvider
}

func GetKinds() map[schema.GroupVersion][]resource.Kind {
	gv := schema.GroupVersion{
		Group:   investigationsv0alpha1.InvestigationKind().Group(),
		Version: investigationsv0alpha1.InvestigationKind().Version(),
	}
	return map[schema.GroupVersion][]resource.Kind{
		gv: {
			investigationsv0alpha1.InvestigationKind(),
			investigationsv0alpha1.InvestigationIndexKind(),
		},
	}
}

// CreateInvestigationCollection creates a resource collection for investigations
func CreateInvestigationCollection() resource.KindCollection {
	kinds := []resource.Kind{
		{
			Schema: investigationsv0alpha1.InvestigationKind(),
			Codecs: map[resource.KindEncoding]resource.Codec{
				resource.KindEncodingJSON: resource.NewJSONCodec(),
			},
		},
		{
			Schema: investigationsv0alpha1.InvestigationIndexKind(),
			Codecs: map[resource.KindEncoding]resource.Codec{
				resource.KindEncodingJSON: resource.NewJSONCodec(),
			},
		},
	}

	return resource.NewKindCollection(kinds...)
}

/*
Example GraphQL queries that would be supported:

query GetInvestigation {
  investigation(name: "my-investigation", namespace: "default") {
    metadata {
      name
      namespace
      uid
      creationTimestamp
    }
    spec {
      title
      createdByProfile {
        uid
        name
        gravatarUrl
      }
      hasCustomName
      isFavorite
      overviewNote
      viewMode {
        mode
        showComments
        showTooltips
      }
    }
  }
}

query ListInvestigations {
  investigations(namespace: "default", limit: 10) {
    metadata {
      name
      namespace
    }
    spec {
      title
      createdByProfile {
        name
      }
      isFavorite
    }
  }
}

mutation CreateInvestigation {
  createInvestigation(input: {
    metadata: {
      name: "new-investigation"
      namespace: "default"
    }
    spec: {
      title: "My New Investigation"
      createdByProfile: {
        uid: "user123"
        name: "John Doe"
        gravatarUrl: "https://gravatar.com/avatar/123"
      }
      hasCustomName: true
      isFavorite: false
      overviewNote: "This is a new investigation"
      viewMode: {
        mode: "compact"
        showComments: true
        showTooltips: true
      }
    }
  }) {
    metadata {
      name
      uid
    }
    spec {
      title
    }
  }
}

# Note: With smart conflict resolution, field names are clean:
# - "investigation" instead of "investigations_investigation"  
# - "investigations" instead of "investigations_investigations"
# - Only conflicts get prefixed (e.g., if multiple apps had "customQuery")
*/
