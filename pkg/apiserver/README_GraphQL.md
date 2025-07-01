# GraphQL for Grafana Apps

Grafana provides zero-touch GraphQL support for apps. Apps automatically get GraphQL schemas generated from their resource kinds without any code changes.

## How It Works

1. **Auto-Discovery**: Apps are automatically discovered via Wire dependency injection
2. **Shared Storage**: GraphQL uses the same storage backend as REST APIs  
3. **Auto-Schema**: GraphQL schemas generated automatically from resource kinds
4. **Unified Endpoint**: All apps accessible via `/apis/graphql`

## Architecture

```
App Providers (Wire DI) → GraphQL Service → Unified Schema → /apis/graphql
```

### App Provider Interface

Apps implement this interface automatically:

```go
type AppProvider interface {
    GetAppName() string
    GetStorageBackend() interface{}  // Same as REST API
    GetResourceKinds() []interface{} // Auto-discovered
}
```

### Example Implementation

```go
// Zero-touch GraphQL - no code changes needed
func (p *MyAppProvider) GetGraphQLProvider() *router.AppGraphQLProvider {
    if cachedApp := p.Provider.GetCachedApp(); cachedApp != nil {
        graphqlProvider, _ := router.NewAppGraphQLProviderFromApp("myapp", cachedApp)
        return graphqlProvider
    }
    return nil
}
```

## Usage

### Enable GraphQL
```ini
[feature_toggles]
apiServerGraphQL = true
```

### Query Examples

```graphql
# Basic query
{ ping }

# Resource query  
{ 
  investigation(name: "test", namespace: "default") { 
    metadata { name }
    spec { title }
  }
}

# Introspection
{ __schema { queryType { fields { name } } } }
```

### Testing
```bash
curl -X POST http://localhost:3000/apis/graphql \
  -H "Content-Type: application/json" \
  -H "Authorization: Basic YWRtaW46YWRtaW4=" \
  -d '{"query": "{ ping }"}'
```

## Key Benefits

- **Zero Code Changes**: Apps get GraphQL automatically
- **Data Consistency**: Same storage as REST APIs
- **Type Safety**: Schemas match Kubernetes resource definitions  
- **Auto-Discovery**: New apps automatically included
- **Conflict Resolution**: Smart field prefixing when needed

## Implementation Details

### Schema Generation
- GraphQL types generated from resource kinds automatically
- Field names derived from resource metadata (name, namespace, spec, etc.)
- List and single queries created for each resource type

### Storage Integration  
- Uses app's existing `ClientGenerator` for data access
- Same authentication and authorization as REST APIs
- No separate database connections or caching layers

### Error Handling
- "Cannot query field" errors indicate app not ready - wait for full startup
- Connection errors mean server not started - check process status
- Schema rebuilds automatically when apps become available

This implementation provides true zero-touch GraphQL where app developers focus on domain logic and GraphQL support comes automatically. 