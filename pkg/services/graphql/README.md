# Grafana GraphQL API Service

## Overview

The GraphQL API service provides a unified GraphQL layer over Grafana's `/apis` endpoints, solving the limitation where frontend applications need to make multiple waterfall requests to fetch related data. Instead of making separate API calls for each resource type, GraphQL allows querying multiple resources and their relationships in a single request.

## Architecture

The GraphQL service consists of three main components:

### 1. **API Client** (`client.go`)

- Handles HTTP requests to Grafana's `/apis` endpoints
- Manages authentication by forwarding request context
- Provides typed methods for specific resource types (e.g., `GetDashboards`, `GetDashboard`)
- Includes a generic `GetResources` method for other resource types

### 2. **GraphQL Service** (`service.go`)

- Defines GraphQL schema with types and resolvers
- Maps GraphQL queries to API client calls
- Handles request context propagation
- Provides schema introspection for GraphQL clients

### 3. **Integration** (`integration.go`)

- Registers GraphQL routes with Grafana's HTTP server
- Provides route registration functions

## Current Capabilities

### Available Queries

```graphql
# Get all dashboards in a namespace
query GetDashboards {
  dashboards(namespace: "default") {
    metadata {
      name
      namespace
      uid
      creationTimestamp
      labels
      annotations
    }
    spec {
      title
      description
      tags
      dashboard  # Full dashboard JSON
    }
  }
}

# Get a specific dashboard
query GetDashboard {
  dashboard(namespace: "default", name: "dashboard-name") {
    metadata {
      name
      uid
    }
    spec {
      title
      description
      tags
    }
  }
}

# Generic resources query (for other resource types)
query GetGenericResources {
  resources(
    group: "folder.grafana.app",
    version: "v1alpha1", 
    namespace: "default",
    resource: "folders"
  )
}
```

### GraphQL Types

- **Dashboard**: Represents dashboard resources from `/apis/dashboard.grafana.app/v1beta1`
- **DashboardMetadata**: Kubernetes-style metadata (name, namespace, uid, etc.)
- **DashboardSpec**: Dashboard-specific data (title, description, tags, dashboard JSON)
- **Resource**: Generic resource type for other API endpoints

## Adding New Resource Types

Follow these steps to add support for new `/apis` endpoints:

### Step 1: Define Resource Types

In `client.go`, add struct definitions for your resource:

```go
// Example: Adding support for folders
type FolderResource struct {
    APIVersion string `json:"apiVersion"`
    Kind       string `json:"kind"`
    Metadata   struct {
        Name            string            `json:"name"`
        Namespace       string            `json:"namespace"`
        UID             string            `json:"uid"`
        CreationTime    string            `json:"creationTimestamp"`
        Labels          map[string]string `json:"labels,omitempty"`
        Annotations     map[string]string `json:"annotations,omitempty"`
        ResourceVersion string            `json:"resourceVersion"`
    } `json:"metadata"`
    Spec struct {
        Title       string `json:"title"`
        Description string `json:"description,omitempty"`
        // Add folder-specific fields
    } `json:"spec"`
}

type FolderList struct {
    APIVersion string           `json:"apiVersion"`
    Kind       string           `json:"kind"`
    Items      []FolderResource `json:"items"`
    Metadata   struct {
        Continue        string `json:"continue,omitempty"`
        ResourceVersion string `json:"resourceVersion"`
    } `json:"metadata"`
}
```

### Step 2: Add API Client Methods

In `client.go`, add methods to fetch your resource:

```go
// GetFolders retrieves folders from the /apis endpoint
func (c *APIClient) GetFolders(ctx context.Context, namespace string, reqCtx *contextmodel.ReqContext) (*FolderList, error) {
    // Build the URL
    apiPath := fmt.Sprintf("/apis/folder.grafana.app/v1alpha1/namespaces/%s/folders", namespace)
    fullURL := c.baseURL + apiPath

    // Create the request
    req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }

    // Add authentication headers from the request context
    if reqCtx != nil && reqCtx.Req != nil {
        if auth := reqCtx.Req.Header.Get("Authorization"); auth != "" {
            req.Header.Set("Authorization", auth)
        }
        for _, cookie := range reqCtx.Req.Cookies() {
            req.AddCookie(cookie)
        }
    }

    req.Header.Set("Accept", "application/json")
    req.Header.Set("Content-Type", "application/json")

    // Make the request
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("failed to make request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
    }

    // Parse the response
    var folderList FolderList
    if err := json.NewDecoder(resp.Body).Decode(&folderList); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }

    return &folderList, nil
}
```

### Step 3: Define GraphQL Types

In `service.go`, add GraphQL type definitions:

```go
// Add to NewService() function
folderMetadataType := graphql.NewObject(graphql.ObjectConfig{
    Name: "FolderMetadata",
    Fields: graphql.Fields{
        "name": &graphql.Field{
            Type: graphql.String,
        },
        "namespace": &graphql.Field{
            Type: graphql.String,
        },
        "uid": &graphql.Field{
            Type: graphql.String,
        },
        // Add other metadata fields...
    },
})

folderSpecType := graphql.NewObject(graphql.ObjectConfig{
    Name: "FolderSpec",
    Fields: graphql.Fields{
        "title": &graphql.Field{
            Type: graphql.String,
        },
        "description": &graphql.Field{
            Type: graphql.String,
        },
        // Add folder-specific fields...
    },
})

folderType := graphql.NewObject(graphql.ObjectConfig{
    Name: "Folder",
    Fields: graphql.Fields{
        "apiVersion": &graphql.Field{
            Type: graphql.String,
        },
        "kind": &graphql.Field{
            Type: graphql.String,
        },
        "metadata": &graphql.Field{
            Type: folderMetadataType,
        },
        "spec": &graphql.Field{
            Type: folderSpecType,
        },
    },
})
```

### Step 4: Add GraphQL Resolvers

In `service.go`, add query fields to the `rootQuery`:

```go
// Add to the Fields map in rootQuery
"folders": &graphql.Field{
    Type: graphql.NewList(folderType),
    Args: graphql.FieldConfigArgument{
        "namespace": &graphql.ArgumentConfig{
            Type:        graphql.NewNonNull(graphql.String),
            Description: "Namespace to search for folders",
        },
    },
    Resolve: func(p graphql.ResolveParams) (interface{}, error) {
        namespace := p.Args["namespace"].(string)
        
        // Get request context from GraphQL context or create a minimal one
        reqCtx, ok := p.Context.Value("reqContext").(*contextmodel.ReqContext)
        if !ok {
            // Create a minimal request context from the HTTP request
            if httpReq, exists := p.Context.Value("httpRequest").(*http.Request); exists {
                reqCtx = &contextmodel.ReqContext{
                    Context: &web.Context{Req: httpReq},
                }
            }
        }

        // Call the real API
        folderList, err := apiClient.GetFolders(p.Context, namespace, reqCtx)
        if err != nil {
            return nil, err
        }

        // Convert to GraphQL-friendly format
        var folders []map[string]interface{}
        for _, folder := range folderList.Items {
            folders = append(folders, map[string]interface{}{
                "apiVersion": folder.APIVersion,
                "kind":       folder.Kind,
                "metadata": map[string]interface{}{
                    "name":              folder.Metadata.Name,
                    "namespace":         folder.Metadata.Namespace,
                    "uid":               folder.Metadata.UID,
                    "creationTimestamp": folder.Metadata.CreationTime,
                    "labels":            folder.Metadata.Labels,
                    "annotations":       folder.Metadata.Annotations,
                },
                "spec": map[string]interface{}{
                    "title":       folder.Spec.Title,
                    "description": folder.Spec.Description,
                },
            })
        }

        return folders, nil
    },
},
```

### Step 5: Test Your Implementation

1. **Build and start Grafana:**
   ```bash
   go build -o bin/grafana ./pkg/cmd/grafana
   ./bin/grafana server
   ```

2. **Test with curl:**
   ```bash
   curl -X POST http://localhost:3000/api/graphql \
     -H "Content-Type: application/json" \
     -u admin:admin \
     -d '{"query": "{ folders(namespace: \"default\") { metadata { name uid } spec { title } } }"}'
   ```

3. **Test with Bruno or other GraphQL clients:**
   - Endpoint: `http://localhost:3000/api/graphql`
   - Auth: Basic Auth (admin/admin)

## Using Bruno for GraphQL Development

[Bruno](https://usebruno.com) is the recommended GraphQL client for developing and testing the GraphQL API. Here's how to set it up and use it effectively:

### Setting up Bruno

1. **Install Bruno**:
   - Download from [usebruno.com](https://usebruno.com/)
   - Or install via package manager: `brew install bruno` (macOS)

2. **Create a New Collection**:
   - Open Bruno
   - Create a new collection (e.g., "Grafana GraphQL API")
   - Add a new GraphQL request

3. **Configure the GraphQL Endpoint**:
   - **Method**: POST
   - **URL**: `http://localhost:3000/api/graphql`
   - **Type**: GraphQL

4. **Setup Authentication**:
   - Go to the **Headers** tab
   - Add the header `Authorization: Basic YWRtaW46YWRtaW4=` (base64 encoded `admin:admin`)

### Schema Introspection

Bruno automatically loads the GraphQL schema through introspection:

1. Once configured, Bruno will fetch the schema automatically
2. You'll see the schema documentation in the right panel
3. Use **Ctrl+Space** for autocomplete based on the schema
4. Browse available queries, types, and fields in the schema explorer

### Example Queries in Bruno

Here are some example queries you can run in Bruno:

#### List All Dashboards

```graphql
query GetDashboards {
  dashboards(namespace: "default") {
    metadata {
      name
      namespace
      uid
      creationTimestamp
    }
    spec {
      title
      description
      tags
    }
  }
}
```

#### Get Specific Dashboard

```graphql
query GetDashboard($namespace: String!, $name: String!) {
  dashboard(namespace: $namespace, name: $name) {
    metadata {
      name
      uid
    }
    spec {
      title
      description
      tags
      dashboard
    }
  }
}
```

With query variables:

```json
{
  "namespace": "default",
  "name": "your-dashboard-name"
}
```

#### Schema Introspection Query

```graphql
query IntrospectionQuery {
  __schema {
    queryType {
      name
    }
    types {
      name
      kind
      description
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

### Bruno Features for GraphQL Development

1. **Schema Explorer**: Browse all available types and queries
2. **Autocomplete**: Type-aware suggestions as you write queries
3. **Syntax Highlighting**: GraphQL syntax highlighting
4. **Variable Management**: Easily manage query variables
5. **Response Formatting**: Beautiful JSON response formatting
6. **Request History**: Track your query history
7. **Collections**: Organize related queries
8. **Environment Variables**: Manage different endpoints (dev, staging, prod)

### Bruno Tips

- **Use Ctrl+Space** for autocomplete based on the GraphQL schema
- **Set up environment variables** for different Grafana instances
- **Save common queries** in your collection for reuse
- **Use query variables** instead of hardcoding values
- **Explore the schema** using the right-side panel documentation

### Alternative GraphQL Clients

If you prefer other tools, these also work well with the GraphQL API:

- **GraphiQL**: Web-based GraphQL IDE
- **Altair GraphQL Client**: Cross-platform GraphQL client
- **Insomnia**: REST/GraphQL client with good GraphQL support
- **Postman**: Now supports GraphQL queries

All of these clients support schema introspection and will work with the `/api/graphql` endpoint using basic authentication.

## Example: Dashboard Implementation

The dashboard implementation serves as a reference for adding new resource types:

### API Endpoints Mapped:

- `GET /apis/dashboard.grafana.app/v1beta1/namespaces/{namespace}/dashboards` → `dashboards` query
- `GET /apis/dashboard.grafana.app/v1beta1/namespaces/{namespace}/dashboards/{name}` → `dashboard` query

### GraphQL Schema Generated:

```graphql
type Dashboard {
  apiVersion: String
  kind: String
  metadata: DashboardMetadata
  spec: DashboardSpec
}

type DashboardMetadata {
  name: String
  namespace: String
  uid: String
  creationTimestamp: String
  labels: JSON
  annotations: JSONAnnotations
}

type DashboardSpec {
  title: String
  description: String
  tags: [String]
  dashboard: DashboardJSON  # Full dashboard definition
}
```

## Benefits Over Direct `/apis` Calls

1. **Reduced API Calls**: Single GraphQL query instead of multiple REST calls
2. **Field Selection**: Request only the data you need
3. **Type Safety**: Strong schema validation
4. **Query Flexibility**: Combine multiple resources in one request
5. **Future Joins**: Foundation for cross-resource relationships

## Common Patterns

### Authentication

All API calls automatically forward authentication from the original GraphQL request:

- Basic Auth headers
- Session cookies
- Authorization headers

### Error Handling

The service returns GraphQL-formatted errors for:

- Missing authentication
- API request failures
- Invalid parameters

### Namespace Scoping

Most resources are namespace-scoped. Always include `namespace` as a required argument for list queries.

## Development Tips

1. **Use Bruno or similar GraphQL clients** for interactive development and testing
2. **Check API documentation** for the specific `/apis` endpoint structure
3. **Test authentication** by verifying API calls work with proper credentials
4. **Handle list vs. single resource** patterns consistently
5. **Add proper error handling** for all API client methods

## Future Enhancements

Potential improvements for the GraphQL service:

1. **Cross-Resource Joins**: Query dashboards with their folder information in one request
2. **Filtering and Sorting**: Add GraphQL arguments for filtering results
3. **Pagination**: Implement cursor-based pagination for large result sets
4. **Subscriptions**: Real-time updates for resource changes
5. **Caching**: Add response caching for frequently accessed data
6. **Batch Loading**: Optimize N+1 query problems with DataLoader pattern

## Troubleshooting

### Common Issues

1. **Build Errors**: Ensure all imports are correct and struct fields match API responses
2. **Authentication Failures**: Verify request context is properly forwarded
3. **Empty Results**: Check API endpoint URLs and namespace parameters
4. **Schema Errors**: Validate GraphQL type definitions and resolver return types

### Debugging

1. **Use Bruno or similar GraphQL clients** for interactive query testing
2. **Check server logs** for API request errors
3. **Test direct API calls** first before implementing GraphQL layer
4. **Use curl** to verify authentication and response format

## Contributing

When adding new resource types:

- Follow the established patterns from dashboard implementation
- Update the schema documentation

For questions or help with implementation, refer to the dashboard implementation in `service.go` and `client.go` as reference examples.
