package collection

import (
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	restclient "k8s.io/client-go/rest"

	"github.com/grafana/grafana-app-sdk/app"
	appsdkapiserver "github.com/grafana/grafana-app-sdk/k8s/apiserver"
	"github.com/grafana/grafana-app-sdk/simple"
	"github.com/grafana/grafana/apps/collection/pkg/apis"
	collectionv0alpha1 "github.com/grafana/grafana/apps/collection/pkg/apis/collection/v0alpha1"
	collectionapp "github.com/grafana/grafana/apps/collection/pkg/app"
	"github.com/grafana/grafana/pkg/apimachinery/utils"
	grafanarest "github.com/grafana/grafana/pkg/apiserver/rest"
	"github.com/grafana/grafana/pkg/services/apiserver/appinstaller"
	"github.com/grafana/grafana/pkg/services/apiserver/endpoints/request"
	"github.com/grafana/grafana/pkg/services/featuremgmt"
	"github.com/grafana/grafana/pkg/setting"
)

var (
	_ appsdkapiserver.AppInstaller       = (*CollectionAppInstaller)(nil)
	_ appinstaller.LegacyStorageProvider = (*CollectionAppInstaller)(nil)
)

type CollectionAppInstaller struct {
	appsdkapiserver.AppInstaller
	cfg *setting.Cfg
}

func RegisterApp(
	cfg *setting.Cfg,
	features featuremgmt.FeatureToggles,
) (*CollectionAppInstaller, error) {
	installer := &CollectionAppInstaller{
		cfg: cfg,
	}
	specificConfig := any(&collectionapp.CollectionConfig{})
	provider := simple.NewAppProvider(apis.LocalManifest(), specificConfig, collectionapp.New)

	appConfig := app.Config{
		KubeConfig:     restclient.Config{}, // this will be overridden by the installer's InitializeApp method
		ManifestData:   *apis.LocalManifest().ManifestData,
		SpecificConfig: specificConfig,
	}
	i, err := appsdkapiserver.NewDefaultAppInstaller(provider, appConfig, apis.ManifestGoTypeAssociator, apis.ManifestCustomRouteResponsesAssociator)
	if err != nil {
		return nil, err
	}
	installer.AppInstaller = i

	return installer, nil
}

// GetLegacyStorage returns the legacy storage for the collection app.
func (p *CollectionAppInstaller) GetLegacyStorage(requested schema.GroupVersionResource) grafanarest.Storage {
	gvr := schema.GroupVersionResource{
		Group:    collectionv0alpha1.CollectionKind().Group(),
		Version:  collectionv0alpha1.CollectionKind().Version(),
		Resource: collectionv0alpha1.CollectionKind().Plural(),
	}
	if requested.String() != gvr.String() {
		return nil
	}
	legacyStore := &legacyStorage{
		namespacer: request.GetNamespaceMapper(p.cfg),
	}
	legacyStore.tableConverter = utils.NewTableConverter(
		gvr.GroupResource(),
		utils.TableColumns{
			Definition: []metav1.TableColumnDefinition{
				{Name: "Name", Type: "string", Format: "name"},
				{Name: "Title", Type: "string", Format: "string", Description: "The collection name"},
				{Name: "Created At", Type: "date"},
			},
			Reader: func(obj any) ([]interface{}, error) {
				m, ok := obj.(*collectionv0alpha1.Collection)
				if !ok {
					return nil, fmt.Errorf("expected collection")
				}
				return []interface{}{
					m.Name,
					m.Spec.Title,
					m.CreationTimestamp.UTC().Format(time.RFC3339),
				}, nil
			},
		},
	)
	return legacyStore
}
