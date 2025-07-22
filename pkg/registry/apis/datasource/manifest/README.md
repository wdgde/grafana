# Manifest-Based Datasource Schema System

This package implements a system for deriving OpenAPI schemas for datasource plugins from `app.Manifest` files. When a plugin has a `manifest.json` file in its package, this system will automatically read, parse, and load it as `app.ManifestData`, converting the `Kinds` field into `DataSourceOpenAPIExtension.Schemas`.

## Architecture

The system is designed to be modular and easily removable if necessary:

### Components

1. **PluginManifestLoader** (`loader.go`)
   - Loads `manifest.json` files from plugin directories
   - Converts `app.ManifestData` to `DataSourceOpenAPIExtension`
   - Handles schema conversion from `app.VersionSchema` to OpenAPI `spec.Schema`

2. **ManifestExtensionProvider** (`provider.go`)
   - Provides a clean interface for getting manifest-based extensions
   - Handles error cases gracefully (returns nil if manifest doesn't exist)

3. **ExtensionFactory** (`factory.go`)
   - Combines manifest-based extensions with hardcoded extensions
   - Provides fallback to hardcoded extensions for specific plugins
   - Implements the strategy pattern for extension resolution

4. **Integration** (`register.go`)
   - Integrates the factory into the datasource API registration process
   - Replaces hardcoded extension logic with manifest-based approach

## Usage

### For Plugin Developers

To add schema support to your datasource plugin, create a `manifest.json` file in your plugin's root directory:

```json
{
  "appName": "my-datasource",
  "group": "my-datasource.grafana.app",
  "kinds": [
    {
      "kind": "DataSourceConfig",
      "scope": "Namespaced",
      "versions": [
        {
          "name": "v1",
          "schema": {
            "type": "object",
            "properties": {
              "url": {
                "type": "string",
                "description": "The URL of the datasource"
              },
              "apiKey": {
                "type": "string",
                "description": "API key for authentication",
                "x-secure": true
              }
            },
            "required": ["url", "apiKey"]
          }
        }
      ]
    }
  ]
}
```

### For Grafana Developers

The system automatically integrates with the existing datasource API registration:

1. When a plugin is loaded, the system checks for a `manifest.json` file
2. If found, it parses the manifest and converts schemas to OpenAPI format
3. The schemas are added to the `DataSourceOpenAPIExtension.Schemas` field
4. If no manifest is found, it falls back to hardcoded extensions

## File Structure

```
pkg/registry/apis/datasource/manifest/
├── loader.go              # Core manifest loading and conversion logic
├── provider.go            # Clean interface for extension providers
├── factory.go             # Factory that combines manifest and hardcoded extensions
├── loader_test.go         # Tests for the loader functionality
├── example_manifest.json  # Example manifest file
└── README.md             # This documentation
```

## Integration Points

The system integrates with the existing datasource API registration in `pkg/registry/apis/datasource/register.go`:

- **DataSourceAPIBuilder**: Now includes an `extensionFactory` field
- **RegisterAPIService**: Uses the factory instead of hardcoded extensions
- **NewDataSourceAPIBuilder**: Initializes the extension factory

## Benefits

1. **Modularity**: Each component is separate and easily removable
2. **Backward Compatibility**: Falls back to hardcoded extensions
3. **Extensibility**: Easy to add new extension sources
4. **Type Safety**: Uses strongly typed `app.Manifest` structures
5. **Error Handling**: Graceful handling of missing manifests

## Migration Path

The system is designed to be easily removable if necessary:

1. Remove the `extensionFactory` field from `DataSourceAPIBuilder`
2. Remove the factory initialization in `NewDataSourceAPIBuilder`
3. Restore the hardcoded extension logic in `RegisterAPIService`
4. Delete the `manifest/` directory

This ensures that the changes can be easily reverted if needed. 