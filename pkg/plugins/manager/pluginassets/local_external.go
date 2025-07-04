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

func (l *LocalExternal) PluginAssets(target *plugins.FoundPlugin, parent *plugins.FoundPlugin) (plugins.AssetInfo, error) {
	baseURL, err := l.baseURL(target, parent)
	if err != nil {
		return plugins.AssetInfo{}, err
	}

	moduleURL, err := l.moduleURL(target, parent)
	if err != nil {
		return plugins.AssetInfo{}, err
	}

	return plugins.AssetInfo{
		Base:   baseURL,
		Module: moduleURL,
		RelativeURLFn: func(s string) (string, error) {
			return l.relativeURL(target, parent, s)
		},
	}, nil
}

func (l *LocalExternal) baseURL(plugin *plugins.FoundPlugin, parent *plugins.FoundPlugin) (string, error) {
	if l.cdn.PluginSupported(plugin.JSONData.ID) {
		return l.cdn.AssetURL(plugin.JSONData.ID, plugin.JSONData.Info.Version, "")
	}

	if parent != nil {
		relPath, err := parent.FS.Rel(plugin.FS.Base())
		if err != nil {
			return "", err
		}

		return path.Join(grafanaServerPluginsAssetPath, parent.JSONData.ID, relPath), nil
	}

	return path.Join(grafanaServerPluginsAssetPath, plugin.JSONData.ID), nil
}

func (l *LocalExternal) moduleURL(plugin *plugins.FoundPlugin, parent *plugins.FoundPlugin) (string, error) {
	if l.cdn.PluginSupported(plugin.JSONData.ID) {
		return l.cdn.AssetURL(plugin.JSONData.ID, plugin.JSONData.Info.Version, "module.js")
	}

	if parent != nil {
		relPath, err := parent.FS.Rel(plugin.FS.Base())
		if err != nil {
			return "", err
		}

		if l.cdn.PluginSupported(parent.JSONData.ID) {
			return l.cdn.AssetURL(parent.JSONData.ID, parent.JSONData.Info.Version, path.Join(relPath, "module.js"))
		}

		return path.Join(grafanaServerPluginsAssetPath, parent.JSONData.ID, relPath, "module.js"), nil
	}

	return path.Join(grafanaServerPluginsAssetPath, plugin.JSONData.ID, "module.js"), nil
}

func (l *LocalExternal) relativeURL(plugin *plugins.FoundPlugin, parent *plugins.FoundPlugin, assetPath string) (string, error) {
	if l.cdn.PluginSupported(plugin.JSONData.ID) {
		return l.cdn.NewCDNURLConstructor(plugin.JSONData.ID, plugin.JSONData.ID).StringPath(assetPath)
	}
	if parent != nil {
		if l.cdn.PluginSupported(parent.JSONData.ID) {
			relPath, err := parent.FS.Rel(plugin.FS.Base())
			if err != nil {
				return "", err
			}
			return l.cdn.AssetURL(parent.JSONData.ID, parent.JSONData.Info.Version, path.Join(relPath, assetPath))
		}
	}

	// Handle absolute URLs
	if u, err := url.Parse(assetPath); err == nil && u.IsAbs() {
		return assetPath, nil
	}

	// Calculate base URL
	baseURL, err := l.baseURL(plugin, parent)
	if err != nil {
		return "", err
	}

	// Avoid double-prefixing
	if strings.HasPrefix(assetPath, baseURL) {
		return assetPath, nil
	}

	return path.Join(baseURL, assetPath), nil
}
