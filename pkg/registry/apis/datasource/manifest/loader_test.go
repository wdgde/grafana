package manifest

import (
	"testing"

	"github.com/grafana/grafana-app-sdk/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPluginManifestLoader_ConvertManifestToOpenAPIExtension(t *testing.T) {
	loader := NewPluginManifestLoader()

	// Create a schema map that represents a full OpenAPI document structure
	schemaMap := map[string]interface{}{
		"spec": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"testField": map[string]interface{}{
					"type": "string",
				},
			},
			"required": []string{"testField"},
		},
	}

	// Create VersionSchema from map
	versionSchema, err := app.VersionSchemaFromMap(schemaMap)
	require.NoError(t, err)

	manifestData := &app.ManifestData{
		AppName: "test-plugin",
		Group:   "test.grafana.app",
		Kinds: []app.ManifestKind{
			{
				Kind:  "TestKind",
				Scope: "Namespaced",
				Versions: []app.ManifestKindVersion{
					{
						Name:   "v1",
						Schema: versionSchema,
					},
				},
			},
		},
	}

	extension, err := loader.ConvertManifestToOpenAPIExtension(manifestData)
	require.NoError(t, err)
	assert.NotNil(t, extension)
	assert.Len(t, extension.Schemas, 1)
	assert.Contains(t, extension.Schemas, "TestKindv1")
}
