package investigations

import (
	"github.com/grafana/grafana-app-sdk/app"
	"github.com/grafana/grafana-app-sdk/plugin/router"
	"github.com/grafana/grafana-app-sdk/simple"
	"github.com/grafana/grafana/apps/investigations/pkg/apis"
	investigationv0alpha1 "github.com/grafana/grafana/apps/investigations/pkg/apis/investigations/v0alpha1"
	investigationapp "github.com/grafana/grafana/apps/investigations/pkg/app"
	"github.com/grafana/grafana/pkg/services/apiserver/builder"
	"github.com/grafana/grafana/pkg/services/apiserver/builder/runner"
	"github.com/grafana/grafana/pkg/setting"
)

type InvestigationsAppProvider struct {
	app.Provider
	cfg *setting.Cfg
}

func RegisterApp(
	cfg *setting.Cfg,
) *InvestigationsAppProvider {
	provider := &InvestigationsAppProvider{
		cfg: cfg,
	}
	appCfg := &runner.AppBuilderConfig{
		OpenAPIDefGetter:         investigationv0alpha1.GetOpenAPIDefinitions,
		ManagedKinds:             investigationapp.GetKinds(),
		Authorizer:               investigationapp.GetAuthorizer(),
		AllowedV0Alpha1Resources: []string{builder.AllResourcesAllowed},
	}
	provider.Provider = simple.NewAppProvider(apis.LocalManifest(), appCfg, investigationapp.New)
	return provider
}

// GetAppName returns the name of the app
func (p *InvestigationsAppProvider) GetAppName() string {
	return "investigations"
}

// GetGraphQLProvider returns the GraphQL provider for the investigations app
func (p *InvestigationsAppProvider) GetGraphQLProvider() *router.AppGraphQLProvider {
	// Create GraphQL provider on-demand from the simple.App
	if cachedApp := p.Provider.(*simple.AppProvider).GetCachedApp(); cachedApp != nil {
		// Create GraphQL provider directly from the simple.App
		if graphqlProvider, err := router.NewAppGraphQLProviderFromApp("investigations", cachedApp); err == nil {
			return graphqlProvider
		}
	}
	return nil
}
