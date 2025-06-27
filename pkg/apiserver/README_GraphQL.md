# GraphQL Support for Grafana API Server

This document describes the GraphQL implementation for the Grafana API Server, which provides a unified GraphQL endpoint that aggregates schemas from multiple Grafana apps.

## Architecture Overview

The GraphQL implementation follows a clean separation of concerns:

```
┌─────────────────────────────────────────────────────────────┐
│                    Grafana API Server                       │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │              GraphQL Registry                           │ │
│  │  ┌─────────────────┐ ┌─────────────────┐ ┌─────────────┐ │ │
│  │  │   App Provider  │ │   App Provider  │ │   App Provider│ │ │
│  │  │  (Investigations)│ │   (Dashboards)  │ │    (Alerts) │ │ │
│  │  └─────────────────┘ └─────────────────┘ └─────────────┘ │ │
│  └─────────────────────────────────────────────────────────┘ │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │               HTTP Handler                              │ │
│  │           /apis/graphql endpoint                        │ │
│  └─────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────┐
│                Individual Apps (SDK)                        │
│  ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐ │
│  │AppGraphQLProvider│ │AppGraphQLProvider│ │AppGraphQLProvider│ │
│  │ - Schema Builder │ │ - Schema Builder │ │ - Schema Builder │ │
│  │ - Resource Store │ │ - Resource Store │ │ - Resource Store │ │
│  └─────────────────┘ └─────────────────┘ └─────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## Key Components

### 1. GraphQL Registry (`grafana/pkg/apiserver/registry/graphql/registry.go`)
- **Purpose**: Aggregates GraphQL schemas from multiple apps into a unified schema
- **Key Features**:
  - Thread-safe app registration/unregistration
  - Smart conflict resolution with intelligent field prefixing
  - Unified schema building and caching
  - GraphQL introspection support

### 2. HTTP Integration (`grafana/pkg/apiserver/rest/graphql/router.go`)
- **Purpose**: Provides HTTP transport integration for GraphQL endpoints
- **Endpoints**:
  - `POST /apis/graphql` - Execute GraphQL queries (POST only - GET is an antipattern)
- **Features**:
  - Standard GraphQL introspection via introspection queries
  - Proper Content-Type validation (application/json)
  - Request validation and error handling
- **Note**: Health checks should use the standard Grafana `/health` endpoint, not GraphQL-specific routes

### 3. API Server Integration (`grafana/pkg/apiserver/endpoints/graphql/handler.go`)
- **Purpose**: Integrates GraphQL with Grafana's API server framework
- **Features**:
  - Grafana authentication/authorization integration
  - Request context handling
  - Metrics and logging

### 4. Service Integration (`grafana/pkg/services/apiserver/graphql_integration.go`)
- **Purpose**: Manages GraphQL service lifecycle within API server
- **Features**:
  - Feature toggle support (`FlagAPIServerGraphQL`)
  - Service startup/shutdown management
  - Route registration

### 5. App GraphQL Provider (`grafana-app-sdk/plugin/router/graphql_app_provider.go`)
- **Purpose**: Enables individual apps to provide GraphQL schemas
- **Features**:
  - Automatic schema generation from CUE-derived Go structs
  - CRUD operations for all resource kinds
  - Integration with existing app-SDK storage layer

## Smart Conflict Resolution

The GraphQL registry uses intelligent field prefixing to avoid naming conflicts while keeping field names clean:

### How It Works

1. **First Pass**: Collect all field names from all registered apps
2. **Conflict Detection**: Identify fields that exist in multiple apps
3. **Smart Prefixing**: Only add app prefixes when there are actual conflicts

### Examples

```graphql
# Single app scenario - no prefixing needed
type Query {
  investigation(id: ID!): Investigation     # Clean field name
  investigations: [Investigation!]!         # Clean field name
}

# Multiple apps with conflicts - selective prefixing
type Query {
  investigation(id: ID!): Investigation     # No conflict, clean name
  dashboard(id: ID!): Dashboard             # No conflict, clean name
  
  # Only conflicting fields get prefixed
  investigations_search(query: String!): [Investigation!]!
  dashboards_search(query: String!): [Dashboard!]!
}
```

### Benefits

- **Clean API**: No awkward `investigations_investigation` naming
- **Conflict Prevention**: Automatic handling of actual naming conflicts
- **Backward Compatibility**: Existing apps continue to work
- **Developer Experience**: Intuitive field names for clients

## Schema Generation from CUE

The system automatically generates GraphQL schemas from CUE-derived Go structs:

```cue
// CUE definition
Investigation: {
    metadata: {
        name: string
        namespace: string
    }
    spec: {
        title: string
        createdByProfile: {
            uid: string
            name: string
        }
        isFavorite: bool
    }
}
```

Becomes:

```graphql
# Generated GraphQL schema
type Investigation {
  metadata: ObjectMeta!
  spec: InvestigationSpec!
}

type InvestigationSpec {
  title: String!
  createdByProfile: Person!
  isFavorite: Boolean!
}

type Query {
  investigation(name: String!, namespace: String): Investigation
  investigations(namespace: String, limit: Int): [Investigation!]!
}

type Mutation {
  createInvestigation(input: InvestigationInput!): Investigation
  updateInvestigation(input: InvestigationInput!): Investigation
  deleteInvestigation(name: String!, namespace: String): Boolean
}
```

## Usage Examples

### Basic Query
```graphql
query GetInvestigation {
  investigation(name: "my-investigation", namespace: "default") {
    metadata {
      name
      creationTimestamp
    }
    spec {
      title
      isFavorite
    }
  }
}
```

### List Query
```graphql
query ListInvestigations {
  investigations(namespace: "default", limit: 10) {
    metadata { name }
    spec { title }
  }
}
```

### Mutation
```graphql
mutation CreateInvestigation {
  createInvestigation(input: {
    metadata: { name: "new-investigation", namespace: "default" }
    spec: { title: "My Investigation", isFavorite: true }
  }) {
    metadata { name }
    spec { title }
  }
}
```

### Schema Introspection (Standard GraphQL)
```graphql
query IntrospectionQuery {
  __schema {
    queryType { name }
    mutationType { name }
    types {
      name
      kind
      fields {
        name
        type {
          name
          kind
        }
      }
    }
  }
}
```

### Type Introspection
```graphql
query GetInvestigationType {
  __type(name: "Investigation") {
    name
    kind
    fields {
      name
      type {
        name
        kind
      }
    }
  }
}
```

## Implementation Guide

### For App Developers

1. **Create GraphQL Provider**:
```go
func New(cfg app.Config) (app.App, error) {
    // ... existing app setup ...
    
    // Create GraphQL provider
    provider, err := router.NewAppGraphQLProvider(
        "myapp", 
        resourceCollection, 
        store,
    )
    if err != nil {
        return nil, err
    }
    
    return &MyApp{
        App: baseApp,
        graphqlProvider: provider,
    }, nil
}

func (a *MyApp) GetGraphQLProvider() router.AppGraphQLProvider {
    return a.graphqlProvider
}
```

2. **Register with Server**:
The GraphQL service automatically discovers and registers apps that implement the GraphQL provider interface.

### For Server Developers

The GraphQL service is automatically configured when the feature flag is enabled:

```go
// Feature flag in configuration
FlagAPIServerGraphQL = "apiserver-graphql"
```

## Security Considerations

- **Authentication**: Inherits Grafana's existing authentication mechanisms
- **Authorization**: Respects existing RBAC and permissions
- **Input Validation**: Automatic validation of GraphQL inputs
- **Rate Limiting**: Can be configured at the API server level

## Performance Considerations

- **Schema Caching**: Unified schema is cached and only rebuilt when apps change
- **DataLoader Foundation**: Architecture supports DataLoader pattern for N+1 prevention
- **Batching**: Can be implemented at the resolver level
- **Caching**: Integrates with existing Grafana caching mechanisms

## Development and Testing

### Local Development
1. Enable the feature flag: `FlagAPIServerGraphQL`
2. Start Grafana with API server enabled
3. Access GraphQL playground at `/apis/graphql`

### Testing
- Unit tests for individual components
- Integration tests for end-to-end scenarios
- Schema validation tests
- Performance benchmarks

## Migration Guide

### From Old Architecture (App-SDK Registry) to New Architecture (Server-Side Registry)

1. **Remove old GraphQL registry code** from app-SDK
2. **Update app implementations** to use `AppGraphQLProvider` instead of registry
3. **Register providers** with the main Grafana server
4. No changes needed to existing HTTP endpoints - GraphQL is additive

## Future Enhancements

### Planned Features
- **DataLoader Integration**: Built-in N+1 problem prevention
- **Subscriptions**: Real-time updates via GraphQL subscriptions
- **Federation**: Evolution toward Apollo Federation patterns
- **Custom Scalars**: Support for Grafana-specific data types
- **Metrics Integration**: Deep integration with Grafana metrics

### Potential Improvements
- **Query Complexity Analysis**: Prevent expensive queries
- **Persisted Queries**: Improve performance and security
- **Schema Stitching**: Advanced schema composition
- **Custom Directives**: Grafana-specific GraphQL directives

## Troubleshooting

### Common Issues

1. **Schema Composition Failures**
   - Check for type conflicts between apps
   - Verify all apps implement the provider interface correctly
   - Review logs for detailed error messages

2. **Field Resolution Errors**
   - Ensure stores are properly configured
   - Check that resource collections match schema definitions
   - Verify authentication and permissions

3. **Performance Issues**
   - Monitor query complexity
   - Implement DataLoader patterns for related data
   - Consider caching at appropriate levels

### Debug Tools

- **GraphQL Introspection**: Use `__schema` and `__type` queries for schema exploration
- **Standard GraphQL Tools**: Compatible with GraphiQL, Apollo Studio, etc.
- **Grafana Health Endpoint**: Use `/health` for service status
- **Detailed Logging**: Structured fields for debugging and monitoring

## References

- [GraphQL Specification](https://spec.graphql.org/)
- [Apollo Federation](https://www.apollographql.com/docs/federation/)
- [Grafana App SDK](../../../grafana-app-sdk/)
- [Kubernetes API Conventions](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md)

1. **Apps** provide `AppGraphQLProvider` implementations via the SDK
2. **Registry** aggregates all app providers into a unified GraphQL schema
3. **HTTP Handler** exposes the unified schema via `/apis/graphql` endpoint
4. **Endpoint Handler** integrates with Grafana's API server framework
5. **Service** manages the GraphQL functionality lifecycle 