package pluginassets

import (
	"net/url"
	"path"
	"strings"

	"github.com/grafana/grafana/pkg/plugins"
	"github.com/grafana/grafana/pkg/plugins/pluginscdn"
)

var _ plugins.PluginAssetProvider = (*LocalExternal)(nil)

type LocalExternal struct {
	cdn *pluginscdn.Service
}

func NewLocalExternal(cdn *pluginscdn.Service) *LocalExternal {
	return &LocalExternal{
		cdn: cdn,
	}
}

func (l *LocalExternal) PluginAssets(p plugins.PluginAssetPlugin) (plugins.AssetInfo, error) {
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

func (l *LocalExternal) baseURL(plugin plugins.PluginAssetPlugin) (string, error) {
	if l.cdn.PluginSupported(plugin.JSONData().ID) {
		return l.cdn.AssetURL(plugin.JSONData().ID, plugin.JSONData().Info.Version, "")
	}

	if plugin.Parent() != nil {
		relPath, err := plugin.Parent().FS().Rel(plugin.FS().Base())
		if err != nil {
			return "", err
		}

		return path.Join(grafanaServerPluginsAssetPath, plugin.Parent().JSONData().ID, relPath), nil
	}

	return path.Join(grafanaServerPluginsAssetPath, plugin.JSONData().ID), nil
}

func (l *LocalExternal) moduleURL(plugin plugins.PluginAssetPlugin) (string, error) {
	if l.cdn.PluginSupported(plugin.JSONData().ID) {
		return l.cdn.AssetURL(plugin.JSONData().ID, plugin.JSONData().Info.Version, "module.js")
	}

	if plugin.Parent() != nil {
		relPath, err := plugin.Parent().FS().Rel(plugin.FS().Base())
		if err != nil {
			return "", err
		}

		if l.cdn.PluginSupported(plugin.Parent().JSONData().ID) {
			return l.cdn.AssetURL(plugin.Parent().JSONData().ID, plugin.Parent().JSONData().Info.Version, path.Join(relPath, "module.js"))
		}

		return path.Join(grafanaServerPluginsAssetPath, plugin.Parent().JSONData().ID, relPath, "module.js"), nil
	}

	return path.Join(grafanaServerPluginsAssetPath, plugin.JSONData().ID, "module.js"), nil
}

func (l *LocalExternal) relativeURL(plugin plugins.PluginAssetPlugin, assetPath string) (string, error) {
	if l.cdn.PluginSupported(plugin.JSONData().ID) {
		return l.cdn.NewCDNURLConstructor(plugin.JSONData().ID, plugin.JSONData().ID).StringPath(assetPath)
	}
	if plugin.Parent() != nil {
		if l.cdn.PluginSupported(plugin.Parent().JSONData().ID) {
			relPath, err := plugin.Parent().FS().Rel(plugin.FS().Base())
			if err != nil {
				return "", err
			}
			return l.cdn.AssetURL(plugin.Parent().JSONData().ID, plugin.Parent().JSONData().Info.Version, path.Join(relPath, assetPath))
		}
	}

	// Handle absolute URLs
	if u, err := url.Parse(assetPath); err == nil && u.IsAbs() {
		return assetPath, nil
	}

	// Calculate base URL
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
