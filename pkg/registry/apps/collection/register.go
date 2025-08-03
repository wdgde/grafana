package collection

import (
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	restclient "k8s.io/client-go/rest"

	"k8s.io/apiserver/pkg/registry/rest"

	"github.com/grafana/grafana-app-sdk/app"
	appsdkapiserver "github.com/grafana/grafana-app-sdk/k8s/apiserver"
	"github.com/grafana/grafana-app-sdk/simple"
	"github.com/grafana/grafana/apps/collection/pkg/apis"
	collectionv0alpha1 "github.com/grafana/grafana/apps/collection/pkg/apis/collection/v0alpha1"
	collectionapp "github.com/grafana/grafana/apps/collection/pkg/app"
	"github.com/grafana/grafana/pkg/apimachinery/utils"
	"github.com/grafana/grafana/pkg/infra/db"
	"github.com/grafana/grafana/pkg/registry/apps/collection/legacy"
	"github.com/grafana/grafana/pkg/services/apiserver/appinstaller"
	"github.com/grafana/grafana/pkg/services/apiserver/endpoints/request"
	"github.com/grafana/grafana/pkg/services/featuremgmt"
	"github.com/grafana/grafana/pkg/setting"
	"github.com/grafana/grafana/pkg/storage/legacysql"
)

var (
	_ appsdkapiserver.AppInstaller       = (*CollectionAppInstaller)(nil)
	_ appinstaller.LegacyStorageProvider = (*CollectionAppInstaller)(nil)
)

type CollectionAppInstaller struct {
	appsdkapiserver.AppInstaller

	namespacer request.NamespaceMapper
	db         legacysql.LegacyDatabaseProvider
}

func RegisterApp(
	features featuremgmt.FeatureToggles,
	cfg *setting.Cfg,
	db db.DB,
) (*CollectionAppInstaller, error) {
	installer := &CollectionAppInstaller{
		namespacer: request.GetNamespaceMapper(cfg),
		db:         legacysql.NewDatabaseProvider(db),
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
func (p *CollectionAppInstaller) GetLegacyStorage(requested schema.GroupVersionResource) rest.Storage {
	gvr := schema.GroupVersionResource{
		Group:    collectionv0alpha1.StarsKind().Group(),
		Version:  collectionv0alpha1.StarsKind().Version(),
		Resource: collectionv0alpha1.StarsKind().Plural(),
	}
	if requested.String() != gvr.String() {
		return nil
	}
	legacyStore := &legacyStorage{
		namespacer: p.namespacer,
		sql:        &legacy.LegacyStarSQL{DB: p.db},
	}
	legacyStore.tableConverter = utils.NewTableConverter(
		gvr.GroupResource(),
		utils.TableColumns{
			Definition: []metav1.TableColumnDefinition{
				{Name: "Name", Type: "string", Format: "name"},
				{Name: "Title", Type: "string", Format: "string", Description: "The collection name"},
				{Name: "Created At", Type: "date"},
			},
			Reader: func(obj any) ([]any, error) {
				m, ok := obj.(*collectionv0alpha1.Stars)
				if !ok {
					return nil, fmt.Errorf("expected collection")
				}
				return []any{
					m.Name,
					"???",
					m.CreationTimestamp.UTC().Format(time.RFC3339),
				}, nil
			},
		},
	)
	return legacyStore
}
