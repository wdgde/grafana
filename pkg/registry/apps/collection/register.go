package collection

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	genericrest "k8s.io/apiserver/pkg/registry/rest"
	restclient "k8s.io/client-go/rest"

	"github.com/grafana/grafana-app-sdk/app"
	appsdkapiserver "github.com/grafana/grafana-app-sdk/k8s/apiserver"
	"github.com/grafana/grafana-app-sdk/simple"
	"github.com/grafana/grafana/apps/collection/pkg/apis"
	collectionv0alpha1 "github.com/grafana/grafana/apps/collection/pkg/apis/collection/v0alpha1"
	collectionapp "github.com/grafana/grafana/apps/collection/pkg/app"
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
func (p *CollectionAppInstaller) GetLegacyStorage(requested schema.GroupVersionResource) genericrest.Storage {
	gvr := schema.GroupVersionResource{
		Group:    collectionv0alpha1.StarsKind().Group(),
		Version:  collectionv0alpha1.StarsKind().Version(),
		Resource: collectionv0alpha1.StarsKind().Plural(),
	}
	if requested.String() != gvr.String() {
		return nil
	}
	return legacy.NewLegacyStorage(p.namespacer, p.db)
}
