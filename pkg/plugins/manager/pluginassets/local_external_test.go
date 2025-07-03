package pluginassets

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/grafana/grafana/pkg/plugins"
	"github.com/grafana/grafana/pkg/plugins/config"
	"github.com/grafana/grafana/pkg/plugins/manager/fakes"
	"github.com/grafana/grafana/pkg/plugins/pluginscdn"
)

func TestLocalExternal_PluginAssets(t *testing.T) {
	tests := []struct {
		name                string
		cdnTemplate         string
		pluginSettings      map[string]map[string]string
		setupPlugin         func() plugins.PluginAssetPlugin
		expectedBase        string
		expectedModule      string
		expectedRelativeURL string
	}{
		{
			name:        "standalone external plugin (non-CDN)",
			cdnTemplate: "https://cdn.example.com",
			pluginSettings: map[string]map[string]string{
				"external-datasource": {}, // No CDN setting
			},
			setupPlugin: func() plugins.PluginAssetPlugin {
				jsonData := plugins.JSONData{
					ID:   "external-datasource",
					Type: plugins.TypeDataSource,
					Info: plugins.Info{Version: "1.0.0"},
				}
				pluginFS := fakes.NewFakePluginFS("/plugins/external-datasource")
				return plugins.NewPluginWithAssets(jsonData, pluginFS, nil)
			},
			expectedBase:        "public/plugins/external-datasource",
			expectedModule:      "public/plugins/external-datasource/module.js",
			expectedRelativeURL: "public/plugins/external-datasource/styles.css",
		},
		{
			name:        "standalone external plugin (CDN-enabled)",
			cdnTemplate: "https://cdn.example.com",
			pluginSettings: map[string]map[string]string{
				"cdn-plugin": {"cdn": "true"},
			},
			setupPlugin: func() plugins.PluginAssetPlugin {
				jsonData := plugins.JSONData{
					ID:   "cdn-plugin",
					Type: plugins.TypePanel,
					Info: plugins.Info{Version: "2.0.0"},
				}
				pluginFS := fakes.NewFakePluginFS("/plugins/cdn-plugin")
				return plugins.NewPluginWithAssets(jsonData, pluginFS, nil)
			},
			expectedBase:        "https://cdn.example.com/cdn-plugin/2.0.0/public/plugins/cdn-plugin",
			expectedModule:      "https://cdn.example.com/cdn-plugin/2.0.0/public/plugins/cdn-plugin/module.js",
			expectedRelativeURL: "https://cdn.example.com/cdn-plugin/cdn-plugin/public/plugins/cdn-plugin/styles.css",
		},
		{
			name:        "child plugin with non-CDN parent",
			cdnTemplate: "https://cdn.example.com",
			pluginSettings: map[string]map[string]string{
				"parent-app": {}, // No CDN setting
			},
			setupPlugin: func() plugins.PluginAssetPlugin {
				// Create parent plugin (non-CDN)
				parentJSON := plugins.JSONData{
					ID:   "parent-app",
					Type: plugins.TypeApp,
					Info: plugins.Info{Version: "1.0.0"},
				}
				parentFS := fakes.NewFakePluginFS("/plugins/parent-app")
				parentFS.RelFunc = func(childPath string) (string, error) {
					return "panels/child-panel", nil
				}
				parentPlugin := plugins.NewPluginWithAssets(parentJSON, parentFS, nil)

				// Create child plugin
				childJSON := plugins.JSONData{
					ID:   "child-panel",
					Type: plugins.TypePanel,
					Info: plugins.Info{Version: "1.0.0"},
				}
				childFS := fakes.NewFakePluginFS("/plugins/parent-app/panels/child-panel")
				return plugins.NewPluginWithAssets(childJSON, childFS, parentPlugin)
			},
			expectedBase:        "public/plugins/parent-app/panels/child-panel",
			expectedModule:      "public/plugins/parent-app/panels/child-panel/module.js",
			expectedRelativeURL: "public/plugins/parent-app/panels/child-panel/styles.css",
		},
		{
			name:        "child plugin with CDN-enabled parent",
			cdnTemplate: "https://cdn.example.com",
			pluginSettings: map[string]map[string]string{
				"cdn-parent-app": {"cdn": "true"},
			},
			setupPlugin: func() plugins.PluginAssetPlugin {
				// Create parent plugin (CDN-enabled)
				parentJSON := plugins.JSONData{
					ID:   "cdn-parent-app",
					Type: plugins.TypeApp,
					Info: plugins.Info{Version: "1.5.0"},
				}
				parentFS := fakes.NewFakePluginFS("/plugins/cdn-parent-app")
				parentFS.RelFunc = func(childPath string) (string, error) {
					return "extensions/child-extension", nil
				}
				parentPlugin := plugins.NewPluginWithAssets(parentJSON, parentFS, nil)

				// Create child plugin
				childJSON := plugins.JSONData{
					ID:   "child-extension",
					Type: plugins.TypePanel,
					Info: plugins.Info{Version: "1.0.0"},
				}
				childFS := fakes.NewFakePluginFS("/plugins/cdn-parent-app/extensions/child-extension")
				return plugins.NewPluginWithAssets(childJSON, childFS, parentPlugin)
			},
			expectedBase:        "public/plugins/cdn-parent-app/extensions/child-extension",
			expectedModule:      "https://cdn.example.com/cdn-parent-app/1.5.0/public/plugins/cdn-parent-app/extensions/child-extension/module.js",
			expectedRelativeURL: "https://cdn.example.com/cdn-parent-app/1.5.0/public/plugins/cdn-parent-app/extensions/child-extension/styles.css",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &config.PluginManagementCfg{
				PluginsCDNURLTemplate: tc.cdnTemplate,
				PluginSettings:        tc.pluginSettings,
			}
			cdnService := pluginscdn.ProvideService(cfg)
			localExternal := NewLocalExternal(cdnService)
			plugin := tc.setupPlugin()

			assetInfo, err := localExternal.PluginAssets(plugin)
			require.NoError(t, err)

			// Test BaseURL function
			baseURL, err := assetInfo.BaseURL()
			require.NoError(t, err)
			require.Equal(t, tc.expectedBase, baseURL)

			// Test ModuleURL function
			moduleURL, err := assetInfo.ModuleURL()
			require.NoError(t, err)
			require.Equal(t, tc.expectedModule, moduleURL)

			// Test RelativeURL function
			relativeURL, err := assetInfo.RelativeURL("styles.css")
			require.NoError(t, err)
			require.Equal(t, tc.expectedRelativeURL, relativeURL)
		})
	}
}
