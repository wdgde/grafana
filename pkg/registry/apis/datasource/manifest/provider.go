package manifest

import (
	"fmt"

	datasourceV0 "github.com/grafana/grafana/pkg/apis/datasource/v0alpha1"
	"github.com/grafana/grafana/pkg/services/pluginsintegration/pluginstore"
)

// ManifestOpenAPIExtensionProvider provides OpenAPI extensions based on plugin manifests
type ManifestOpenAPIExtensionProvider struct {
	loader *PluginManifestLoader
}

// NewManifestExtensionProvider creates a new manifest extension provider
func NewManifestExtensionProvider() *ManifestOpenAPIExtensionProvider {
	return &ManifestOpenAPIExtensionProvider{
		loader: NewPluginManifestLoader(),
	}
}

// GetExtensionForPlugin attempts to load and convert a plugin manifest to OpenAPI extension
func (p *ManifestOpenAPIExtensionProvider) GetExtensionForPlugin(plugin *pluginstore.Plugin) (*datasourceV0.DataSourceOpenAPIExtension, error) {
	if plugin == nil {
		return nil, fmt.Errorf("plugin is nil")
	}

	// Try to load manifest from plugin
	manifestData, err := p.loader.LoadManifestFromPlugin(plugin)
	if err != nil {
		// Return nil if manifest doesn't exist - this is expected for plugins without manifests
		return nil, nil
	}

	// Convert manifest to OpenAPI extension
	extension, err := p.loader.ConvertManifestToOpenAPIExtension(manifestData)
	if err != nil {
		return nil, fmt.Errorf("failed to convert manifest to OpenAPI extension: %w", err)
	}

	return extension, nil
}
