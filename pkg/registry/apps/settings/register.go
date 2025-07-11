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
		Authorizer:       settingsapp.GetAuthorizer(cfg),
		CustomConfig:     cfg,
		// LegacyStorageGetter: provider.legacyStorageGetter,
	}
	provider.Provider = simple.NewAppProvider(apis.LocalManifest(), appCfg, settingsapp.New)
	return provider
}

// func (c *SettingsAppProvider) legacyStorageGetter(requested schema.GroupVersionResource) grafanarest.Storage {
// 	gvr := schema.GroupVersionResource{
// 		Group:    settingsv0alpha1.SettingKind().Group(),
// 		Version:  settingsv0alpha1.SettingKind().Version(),
// 		Resource: settingsv0alpha1.SettingKind().Plural(),
// 	}
// 	if requested.String() != gvr.String() {
// 		return nil
// 	}
// 	legacyStore := &legacyStorage{
// 		setting:    c.cfg,
// 		namespacer: request.GetNamespaceMapper(c.cfg),
// 	}
// 	legacyStore.tableConverter = utils.NewTableConverter(
// 		gvr.GroupResource(),
// 		utils.TableColumns{
// 			Definition: []metav1.TableColumnDefinition{
// 				{Name: "Name", Type: "string", Format: "name"},
// 				{Name: "Group", Type: "string", Format: "string"},
// 				{Name: "Value", Type: "string", Format: "string"},
// 			},
// 			Reader: func(obj any) ([]interface{}, error) {
// 				m, ok := obj.(*settingsv0alpha1.Setting)
// 				if !ok {
// 					return nil, fmt.Errorf("expected setting")
// 				}
// 				return []interface{}{
// 					m.Name,
// 					m.Spec.Group,
// 					m.Spec.Value,
// 				}, nil
// 			},
// 		},
// 	)
// 	return legacyStore
// }
