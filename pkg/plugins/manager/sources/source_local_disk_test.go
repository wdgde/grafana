package sources

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/grafana/grafana/pkg/plugins"
	"github.com/grafana/grafana/pkg/plugins/config"
	"github.com/grafana/grafana/pkg/plugins/manager/pluginassets"
	"github.com/grafana/grafana/pkg/plugins/pluginscdn"
)

var compareOpts = []cmp.Option{
	cmp.AllowUnexported(LocalSource{}),
	cmp.AllowUnexported(pluginassets.LocalExternal{}),
	cmp.AllowUnexported(pluginscdn.Service{}),
}

func TestDirAsLocalSources(t *testing.T) {
	testdataDir := "../testdata"

	tests := []struct {
		name        string
		pluginsPath string
		cfg         *config.PluginManagementCfg
		expected    []*LocalSource
		err         error
	}{
		{
			name:        "Empty path returns an error",
			pluginsPath: "",
			expected:    []*LocalSource{},
			err:         errors.New("plugins path not configured"),
		},
		{
			name:        "Directory with subdirectories",
			pluginsPath: filepath.Join(testdataDir, "pluginRootWithDist"),
			cfg:         &config.PluginManagementCfg{},
			expected: []*LocalSource{
				{
					paths:         []string{filepath.Join(testdataDir, "pluginRootWithDist", "datasource")},
					strictMode:    true,
					class:         plugins.ClassExternal,
					assetProvider: pluginassets.NewLocalExternal(pluginscdn.ProvideService(&config.PluginManagementCfg{})),
				},
				{
					paths:         []string{filepath.Join(testdataDir, "pluginRootWithDist", "dist")},
					strictMode:    true,
					class:         plugins.ClassExternal,
					assetProvider: pluginassets.NewLocalExternal(pluginscdn.ProvideService(&config.PluginManagementCfg{})),
				},
				{
					paths:         []string{filepath.Join(testdataDir, "pluginRootWithDist", "panel")},
					strictMode:    true,
					class:         plugins.ClassExternal,
					assetProvider: pluginassets.NewLocalExternal(pluginscdn.ProvideService(&config.PluginManagementCfg{})),
				},
			},
		},
		{
			name: "Dev mode disables strict mode for source",
			cfg: &config.PluginManagementCfg{
				DevMode: true,
			},
			pluginsPath: filepath.Join(testdataDir, "pluginRootWithDist"),
			expected: []*LocalSource{
				{
					paths:      []string{filepath.Join(testdataDir, "pluginRootWithDist", "datasource")},
					class:      plugins.ClassExternal,
					strictMode: false,
					assetProvider: pluginassets.NewLocalExternal(pluginscdn.ProvideService(&config.PluginManagementCfg{
						DevMode: true,
					})),
				},
				{
					paths:      []string{filepath.Join(testdataDir, "pluginRootWithDist", "dist")},
					class:      plugins.ClassExternal,
					strictMode: false,
					assetProvider: pluginassets.NewLocalExternal(pluginscdn.ProvideService(&config.PluginManagementCfg{
						DevMode: true,
					})),
				},
				{
					paths:      []string{filepath.Join(testdataDir, "pluginRootWithDist", "panel")},
					class:      plugins.ClassExternal,
					strictMode: false,
					assetProvider: pluginassets.NewLocalExternal(pluginscdn.ProvideService(&config.PluginManagementCfg{
						DevMode: true,
					})),
				},
			},
		},
		{
			name:        "Directory with no subdirectories",
			cfg:         &config.PluginManagementCfg{},
			pluginsPath: filepath.Join(testdataDir, "pluginRootWithDist", "datasource"),
			expected:    []*LocalSource{},
		},
		{
			name:        "Directory with a symlink to a directory",
			pluginsPath: filepath.Join(testdataDir, "symbolic-plugin-dirs"),
			cfg:         &config.PluginManagementCfg{},
			expected: []*LocalSource{
				{
					paths:         []string{filepath.Join(testdataDir, "symbolic-plugin-dirs", "plugin")},
					class:         plugins.ClassExternal,
					strictMode:    true,
					assetProvider: pluginassets.NewLocalExternal(pluginscdn.ProvideService(&config.PluginManagementCfg{})),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DirAsLocalSources(tt.cfg, tt.pluginsPath, plugins.ClassExternal)
			if tt.err != nil {
				require.Errorf(t, err, tt.err.Error())
				return
			}
			require.NoError(t, err)
			if !cmp.Equal(got, tt.expected, compareOpts...) {
				t.Fatalf("Result mismatch (-want +got):\n%s", cmp.Diff(got, tt.expected, compareOpts...))
			}
		})
	}
}
