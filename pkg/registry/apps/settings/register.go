package settings

import (
	"github.com/grafana/grafana-app-sdk/app"
	"github.com/grafana/grafana-app-sdk/simple"
	"github.com/grafana/grafana/apps/settings/pkg/apis"
	settingsv0alpha1 "github.com/grafana/grafana/apps/settings/pkg/apis/settings/v0alpha1"
	settingsapp "github.com/grafana/grafana/apps/settings/pkg/app"
	"github.com/grafana/grafana/pkg/services/apiserver/builder/runner"
	"github.com/grafana/grafana/pkg/setting"
)

type SettingsAppProvider struct {
	app.Provider
	cfg *setting.Cfg
}

func RegisterApp(
	cfg *setting.Cfg,
) *SettingsAppProvider {
	provider := &SettingsAppProvider{
		cfg: cfg,
	}
	appCfg := &runner.AppBuilderConfig{
		OpenAPIDefGetter: settingsv0alpha1.GetOpenAPIDefinitions,
		ManagedKinds:     settingsapp.GetKinds(),
		Authorizer:       settingsapp.GetAuthorizer(),
	}
	provider.Provider = simple.NewAppProvider(apis.LocalManifest(), appCfg, settingsapp.NewFactory(cfg))
	return provider
}
