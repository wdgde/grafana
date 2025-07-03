package pluginassets

import (
	"net/url"
	"path"
	"path/filepath"
	"strings"

	"github.com/grafana/grafana/pkg/plugins"
)

var _ plugins.PluginAssetProvider = (*LocalCore)(nil)

const (
	coreModulePathPrefix          = "core:plugin"
	grafanaServerPublicPath       = "public"
	grafanaServerPluginsAssetPath = "public/plugins"
	grafanaCorePluginsPath        = "app/plugins"
)

// TODO maybe remove struct?
// TODO this file is actually really specific to core plugins based on filesystem (IE maybe using FS abstraction is misleading)
type LocalCore struct{}

func NewLocalCore() *LocalCore {
	return &LocalCore{}
}

func (l *LocalCore) PluginAssets(p plugins.PluginAssetPlugin) (plugins.AssetInfo, error) {
	baseURL, err := l.baseURL(p)
	if err != nil {
		return plugins.AssetInfo{}, err
	}

	moduleURL, err := l.moduleURL(p)
	if err != nil {
		return plugins.AssetInfo{}, err
	}

	return plugins.AssetInfo{
		BaseURLFunc: func() (string, error) {
			return baseURL, nil
		},
		ModuleURLFunc: func() (string, error) {
			return moduleURL, nil
		},
		RelativeURLFunc: func(s string) (string, error) {
			return l.relativeURL(p, s)
		},
	}, nil
}

func (l *LocalCore) baseURL(plugin plugins.PluginAssetPlugin) (string, error) {
	baseDir := getBaseDir(plugin.FS().Base())
	return path.Join(grafanaServerPublicPath, grafanaCorePluginsPath, string(plugin.JSONData().Type), baseDir), nil
}

func (l *LocalCore) moduleURL(plugin plugins.PluginAssetPlugin) (string, error) {
	if filepath.Base(plugin.FS().Base()) == "dist" {
		// Core plugin built externally - fall back to filesystem path
		return path.Join(grafanaServerPluginsAssetPath, plugin.JSONData().ID, "module.js"), nil
	}
	baseDir := getBaseDir(plugin.FS().Base())
	return path.Join(coreModulePathPrefix, baseDir), nil
}

func (l *LocalCore) relativeURL(plugin plugins.PluginAssetPlugin, assetPath string) (string, error) {
	if u, err := url.Parse(assetPath); err == nil && u.IsAbs() {
		return assetPath, nil
	}

	baseURL, err := l.baseURL(plugin)
	if err != nil {
		return "", err
	}

	// Avoid double-prefixing
	if strings.HasPrefix(assetPath, baseURL) {
		return assetPath, nil
	}

	return path.Join(baseURL, assetPath), nil
}

func getBaseDir(pluginDir string) string {
	baseDir := filepath.Base(pluginDir)
	// Decoupled core plugins will be suffixed with "dist" if they have been built
	if baseDir == "dist" {
		return filepath.Base(strings.TrimSuffix(pluginDir, baseDir))
	}
	return baseDir
}
