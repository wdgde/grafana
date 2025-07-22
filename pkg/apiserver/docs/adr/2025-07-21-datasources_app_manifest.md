# ADR-001: Manifest-Based Datasource Schema System

**Date:** 2025-07-21
**Status:** Implemented
**Type:** Architecture Decision Record

## Context

The Grafana API Server currently uses hardcoded OpenAPI schemas for datasource plugins. This approach is inflexible and requires code changes to add schema support for new plugins. We need a more dynamic and extensible system that allows plugins to define their own schemas.

## Decision

We will implement a manifest-based schema system that derives OpenAPI schemas for datasource plugins from `app.Manifest` files. When a plugin has a `manifest.json` file in its package, the system will automatically read, parse, and load it as `app.ManifestData`, converting the `Kinds` field into `DataSourceOpenAPIExtension.Schemas`.

## Implementation

### 🔧 **Modular Components** (Separate files for easy removal)

1. **`loader.go`** - Core manifest loading and conversion logic
2. **`provider.go`** - Clean interface for extension providers
3. **`factory.go`** - Factory that combines manifest and hardcoded extensions
4. **`loader_test.go`** - Tests for the loader functionality
5. **`example_manifest.json`** - Example manifest file
6. **`README.md`** - Comprehensive documentation

### 🔄 **Integration Points**

- **`register.go`** - Added `extensionFactory` parameter to `RegisterAPIService`
- **`NewDataSourceAPIBuilder`** - Accepts `ExtensionFactoryInterface` as parameter
- **`RegisterAPIService`** - Uses factory instead of hardcoded extensions
- **`wireset.go`** - Added `ProvideExtensionFactory` to dependency injection

### 🎯 **Key Features**

1. **Dynamic Discovery**: Automatically reads `manifest.json` from plugin directories
2. **Schema Conversion**: Converts `app.VersionSchema` to OpenAPI `spec.Schema`
3. **Fallback Strategy**: Falls back to hardcoded extensions if no manifest exists
4. **Error Handling**: Graceful handling of missing manifests
5. **Type Safety**: Uses strongly typed `app.Manifest` structures

### 🔄 **How It Works**

1. When a plugin is loaded, the system checks for a `manifest.json` file
2. If found, it parses the manifest and converts the `Kinds` field to OpenAPI schemas
3. The schemas are added to `DataSourceOpenAPIExtension.Schemas`
4. If no manifest exists, it falls back to hardcoded extensions for specific plugins

### ✅ **Benefits**

- **Modular**: Each component is separate and easily removable
- **Backward Compatible**: Existing hardcoded extensions still work
- **Extensible**: Easy to add new extension sources
- **Type Safe**: Uses the `app.Manifest` structure from grafana-app-sdk
- **Tested**: Includes unit tests for core functionality

## Recent Improvements (2025-01-27)

### 🔧 **Interface-Based Design**

The `ExtensionFactory` has been refactored to use an interface-based design for better testability and dependency injection:

#### **Changes Made**

1. **Created `ExtensionFactoryInterface`**:
   - Defines the contract for extension factory implementations
   - Allows for easy mocking and testing
   - Enables dependency injection through wire

2. **Updated `ExtensionFactory`**:
   - Now implements `ExtensionFactoryInterface`
   - Maintains backward compatibility with existing functionality
   - Added compile-time interface compliance check

3. **Enhanced Dependency Injection**:
   - Added `ProvideExtensionFactory()` function for wire integration
   - Updated `NewDataSourceAPIBuilder` to accept interface parameter
   - Modified `RegisterAPIService` to receive factory through dependency injection

4. **Updated Wire Configuration**:
   - Added `ProvideExtensionFactory` to the wire set
   - Ensures proper dependency injection throughout the callstack

#### **Technical Benefits**

- **Testability**: Interface allows for easy mocking in unit tests
- **Flexibility**: Different implementations can be injected as needed
- **Dependency Injection**: Proper wire integration for clean architecture
- **Type Safety**: Compile-time interface compliance checking
- **Maintainability**: Clear separation of concerns and contracts

### 🔧 **Schema Conversion Enhancement**

The schema conversion logic has been improved to use the proper `AsKubeOpenAPI` method from the grafana-app-sdk instead of manual JSON marshaling/unmarshaling.

#### **Changes Made**

1. **Updated `convertSchemaToOpenAPI` method in `loader.go`**:
   - Replaced JSON marshaling/unmarshaling approach with `versionSchema.AsKubeOpenAPI(gvk, ref)`
   - Added proper imports for `schema.GroupVersionKind` and `common.ReferenceCallback`
   - Extracts the main schema from the returned definitions map (excluding kind and list objects)

2. **Fixed Type Compatibility Issues**:
   - Updated method signatures to accept `pluginstore.Plugin` instead of `*plugins.Plugin`
   - Updated `GetExtensionForPlugin` in `factory.go` and `provider.go`
   - Updated `LoadManifestFromPlugin` in `loader.go`
   - Added proper imports for `pluginstore` package

3. **Enhanced Test Coverage**:
   - Modified test to provide proper OpenAPI document structure with "spec" field
   - This allows the `AsKubeOpenAPI` method to work correctly

#### **Technical Benefits**

- **Proper API Usage**: Now uses the intended `AsKubeOpenAPI` method from grafana-app-sdk
- **Better Schema Handling**: Provides more robust schema conversion with proper OpenAPI structure
- **Type Safety**: Fixed compatibility issues between `pluginstore.Plugin` and `*plugins.Plugin`
- **Maintainability**: Uses the established SDK patterns instead of custom conversion logic

#### **Implementation Details**

The `convertSchemaToOpenAPI` method now:
1. Creates a dummy `schema.GroupVersionKind` for schema conversion
2. Calls `versionSchema.AsKubeOpenAPI(gvk, ref)` with a simple reference callback
3. Iterates through the returned definitions to find the main schema (excluding kind/list objects)
4. Returns the extracted `spec.Schema` for use in the OpenAPI extension

This approach leverages the full power of the grafana-app-sdk's schema conversion capabilities while maintaining backward compatibility.

## Example Usage

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
            "spec": {
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
        }
      ]
    }
  ]
}
```

**Note**: The schema should be structured with a "spec" field to work properly with the `AsKubeOpenAPI` method.

## Migration Path

The system is designed to be easily removable if necessary:

1. Remove the `extensionFactory` parameter from `NewDataSourceAPIBuilder`
2. Remove the `extensionFactory` parameter from `RegisterAPIService`
3. Remove `ProvideExtensionFactory` from the wire set
4. Restore the hardcoded extension logic in `RegisterAPIService`
5. Delete the `manifest/` directory

This ensures that the changes can be easily reverted if needed.

## Consequences

### Positive

- **Dynamic Schema Support**: Plugins can now define their own schemas without code changes
- **Better Developer Experience**: Plugin developers have more control over their API schemas
- **Reduced Maintenance**: Less need to maintain hardcoded schemas in Grafana core
- **Type Safety**: Uses the established `app.Manifest` structure
- **Proper SDK Integration**: Now uses the intended `AsKubeOpenAPI` method for schema conversion

### Negative

- **Additional Complexity**: Introduces new components to the codebase
- **Plugin Requirements**: Plugins need to include manifest files for full schema support
- **Learning Curve**: Plugin developers need to understand the manifest format

### Neutral

- **Backward Compatibility**: Existing plugins continue to work with hardcoded fallbacks
- **Modular Design**: Components can be easily removed if needed

## Alternatives Considered

1. **Code Generation**: Generate schemas from plugin code annotations
   - **Pros**: More integrated with existing development workflow
   - **Cons**: Requires changes to plugin development process, more complex

2. **API-First Approach**: Define schemas through API endpoints
   - **Pros**: Dynamic runtime configuration
   - **Cons**: More complex, potential security concerns

3. **Configuration Files**: Use separate configuration files for schemas
   - **Pros**: Simple, familiar approach
   - **Cons**: Duplicates manifest concept, less integrated

4. **Manual JSON Conversion**: Continue using manual JSON marshaling/unmarshaling
   - **Pros**: Simple implementation
   - **Cons**: Doesn't leverage SDK capabilities, less robust

## References

- [Grafana App SDK Documentation](https://github.com/grafana/grafana-app-sdk)
- [OpenAPI Specification](https://swagger.io/specification/)
- [Kubernetes CRD Schema](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/)
