package manifest

import (
	"testing"

	"github.com/grafana/grafana/pkg/plugins"
	"github.com/grafana/grafana/pkg/registry/apis/datasource"
	"github.com/grafana/grafana/pkg/services/pluginsintegration/pluginstore"
	"github.com/stretchr/testify/assert"
)

func TestExtensionFactory_ImplementsInterface(t *testing.T) {
	// This test ensures that ExtensionFactory implements ExtensionFactoryInterface
	var factory datasource.OpenAPIExtensionGetter = NewOpenAPIExtensionProvider()
	assert.NotNil(t, factory)
}

func TestExtensionFactory_GetExtensionForPlugin(t *testing.T) {
	factory := NewOpenAPIExtensionProvider()

	// Test with a plugin that has no manifest (should return nil, nil)
	plugin := pluginstore.Plugin{
		JSONData: plugins.JSONData{
			ID: "test-plugin",
		},
		FS: plugins.NewFakeFS(),
	}

	extension, err := factory.GetOpenAPIExtension(plugin)
	assert.NoError(t, err)
	assert.Nil(t, extension)
}

func TestExtensionFactory_GetExtensionForPlugin_EmptyID(t *testing.T) {
	factory := NewOpenAPIExtensionProvider()

	// Test with a plugin that has empty ID (should return error)
	plugin := pluginstore.Plugin{
		JSONData: plugins.JSONData{
			ID: "",
		},
	}

	extension, err := factory.GetOpenAPIExtension(plugin)
	assert.Error(t, err)
	assert.Nil(t, extension)
	assert.Contains(t, err.Error(), "plugin is nil")
}
