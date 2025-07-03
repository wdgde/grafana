package pluginassets

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/grafana/grafana/pkg/plugins"
	"github.com/grafana/grafana/pkg/plugins/manager/fakes"
)

func TestLocalCore_PluginAssets(t *testing.T) {
	tests := []struct {
		name           string
		pluginID       string
		pluginType     plugins.Type
		basePath       string
		expectedBase   string
		expectedModule string
	}{
		{
			name:           "regular core plugin",
			pluginID:       "cloudwatch",
			pluginType:     plugins.TypeDataSource,
			basePath:       "cloudwatch",
			expectedBase:   "public/app/plugins/datasource/cloudwatch",
			expectedModule: "core:plugin/cloudwatch",
		},
		{
			name:           "externally-built plugin with dist directory",
			pluginID:       "external-build",
			pluginType:     plugins.TypePanel,
			basePath:       "/grafana/external-build/dist",
			expectedBase:   "public/app/plugins/panel/external-build",
			expectedModule: "public/plugins/external-build/module.js",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			localCore := NewLocalCore()

			jsonData := plugins.JSONData{
				ID:   tt.pluginID,
				Type: tt.pluginType,
				Info: plugins.Info{Version: "1.0.0"},
			}

			pluginFS := fakes.NewFakePluginFS(tt.basePath)
			plugin := plugins.NewPluginWithAssets(jsonData, pluginFS, nil)

			assetInfo, err := localCore.PluginAssets(plugin)
			require.NoError(t, err)

			// Test BaseURL function
			baseURL, err := assetInfo.BaseURL()
			require.NoError(t, err)
			require.Equal(t, tt.expectedBase, baseURL)

			// Test ModuleURL function
			moduleURL, err := assetInfo.ModuleURL()
			require.NoError(t, err)
			require.Equal(t, tt.expectedModule, moduleURL)

			// Test RelativeURL function
			relativeURL, err := assetInfo.RelativeURL("styles.css")
			require.NoError(t, err)
			require.Equal(t, tt.expectedBase+"/styles.css", relativeURL)
		})
	}
}

func TestLocalCore_relativeURL(t *testing.T) {
	localCore := NewLocalCore()

	jsonData := plugins.JSONData{
		ID:   "test-plugin",
		Type: plugins.TypePanel,
		Info: plugins.Info{Version: "1.0.0"},
	}

	pluginFS := fakes.NewFakePluginFS("/test-plugin")
	plugin := plugins.NewPluginWithAssets(jsonData, pluginFS, nil)

	tests := []struct {
		name        string
		assetPath   string
		expectedURL string
	}{
		{
			name:        "relative path gets prefixed",
			assetPath:   "styles.css",
			expectedURL: "public/app/plugins/panel/test-plugin/styles.css",
		},
		{
			name:        "absolute URL returned as-is",
			assetPath:   "https://example.com/asset.css",
			expectedURL: "https://example.com/asset.css",
		},
		{
			name:        "already prefixed path not double-prefixed",
			assetPath:   "public/app/plugins/panel/test-plugin/already-prefixed.css",
			expectedURL: "public/app/plugins/panel/test-plugin/already-prefixed.css",
		},
		{
			name:        "empty path returns base URL",
			assetPath:   "",
			expectedURL: "public/app/plugins/panel/test-plugin",
		},
		{
			name:        "absolute path gets prefixed",
			assetPath:   "/asset.css",
			expectedURL: "public/app/plugins/panel/test-plugin/asset.css",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			relativeURL, err := localCore.relativeURL(plugin, tt.assetPath)
			require.NoError(t, err)
			require.Equal(t, tt.expectedURL, relativeURL)
		})
	}
}

func TestGetBaseDir(t *testing.T) {
	tests := []struct {
		name        string
		pluginDir   string
		expectedDir string
	}{
		{
			name:        "regular plugin directory",
			pluginDir:   "plugins/panel/table",
			expectedDir: "table",
		},
		{
			name:        "plugin with dist directory",
			pluginDir:   "/grafana/public/app/plugins/panel/external-build/dist",
			expectedDir: "external-build",
		},
		{
			name:        "dist only directory",
			pluginDir:   "dist",
			expectedDir: ".",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getBaseDir(tt.pluginDir)
			require.Equal(t, tt.expectedDir, result)
		})
	}
}

func TestLocalCore_ErrorHandling(t *testing.T) {
	t.Run("malformed URL handling", func(t *testing.T) {
		localCore := NewLocalCore()

		jsonData := plugins.JSONData{
			ID:   "test-plugin",
			Type: plugins.TypePanel,
			Info: plugins.Info{Version: "1.0.0"},
		}

		pluginFS := fakes.NewFakePluginFS("/grafana/public/app/plugins/panel/test-plugin")
		plugin := plugins.NewPluginWithAssets(jsonData, pluginFS, nil)

		// Test with invalid URL that can't be parsed (path.Join cleans it)
		relativeURL, err := localCore.relativeURL(plugin, "://invalid-url")
		require.NoError(t, err)
		require.Equal(t, "public/app/plugins/panel/test-plugin/:/invalid-url", relativeURL)
	})
}
