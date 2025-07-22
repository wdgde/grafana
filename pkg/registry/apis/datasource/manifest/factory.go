package manifest

import (
	"fmt"

	datasourceV0 "github.com/grafana/grafana/pkg/apis/datasource/v0alpha1"
	"github.com/grafana/grafana/pkg/registry/apis/datasource"
	"github.com/grafana/grafana/pkg/registry/apis/datasource/hardcoded"
	"github.com/grafana/grafana/pkg/services/pluginsintegration/pluginstore"
)

// OpenAPIExtensionProvider creates OpenAPI extensions for datasource plugins
// It tries manifest-based extension first, then falls back to hardcoded extensions.
type OpenAPIExtensionProvider struct {
	manifestProvider *ManifestOpenAPIExtensionProvider
}

// Ensure OpenAPIExtensionProvider implements OpenAPIExtensionGetter
var _ datasource.OpenAPIExtensionGetter = (*OpenAPIExtensionProvider)(nil)

// NewOpenAPIExtensionProvider creates a new extension factory
func NewOpenAPIExtensionProvider() datasource.OpenAPIExtensionGetter {
	return &OpenAPIExtensionProvider{
		manifestProvider: NewManifestExtensionProvider(),
	}
}

// GetOpenAPIExtension attempts to get OpenAPI extension for a plugin
// It tries manifest-based extension first, then falls back to hardcoded extensions
func (f *OpenAPIExtensionProvider) GetOpenAPIExtension(plugin pluginstore.Plugin) (*datasourceV0.DataSourceOpenAPIExtension, error) {
	if plugin.ID == "" {
		return nil, fmt.Errorf("plugin is nil")
	}

	// First, try to get manifest-based extension
	manifestExtension, err := f.manifestProvider.GetExtensionForPlugin(&plugin)
	if err != nil {
		return nil, fmt.Errorf("failed to get manifest extension: %w", err)
	}

	// If manifest extension exists, return it
	if manifestExtension != nil {
		return manifestExtension, nil
	}

	// Fall back to hardcoded extensions for specific plugins
	return f.getHardcodedExtension(plugin.ID)
}

// getHardcodedExtension returns hardcoded extensions for specific plugins
func (f *OpenAPIExtensionProvider) getHardcodedExtension(pluginID string) (*datasourceV0.DataSourceOpenAPIExtension, error) {
	switch pluginID {
	case "grafana-testdata-datasource":
		return hardcoded.TestdataOpenAPIExtension()
	default:
		// Return nil for plugins without hardcoded extensions
		return nil, nil
	}
}

// ProvideOpenAPIExtensionProvider creates a new extension factory for dependency injection
func ProvideOpenAPIExtensionProvider() datasource.OpenAPIExtensionGetter {
	return NewOpenAPIExtensionProvider()
}
