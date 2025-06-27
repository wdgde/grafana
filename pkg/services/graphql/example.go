package graphql

// Example demonstrates how to integrate GraphQL into Grafana

// Step 1: Add to HTTPServer dependencies
// In pkg/api/http_server.go, add the GraphQL service:
//
// type HTTPServer struct {
//     // ... existing fields ...
//     GraphQLService *graphql.Service
// }

// Step 2: Initialize the service in ProvideHTTPServer
// In the ProvideHTTPServer function:
//
// graphqlService, err := graphql.NewService()
// if err != nil {
//     return nil, err
// }
//
// hs := &HTTPServer{
//     // ... existing assignments ...
//     GraphQLService: graphqlService,
// }

// Step 3: Register routes in registerRoutes()
// In pkg/api/api.go, in the registerRoutes() function:
//
// // GraphQL API endpoints
// graphql.RegisterRoutes(r, hs.GraphQLService)

// Example GraphQL Queries you can test:
//
// Basic query:
// {
//   resources(group: "apps", version: "v1", resource: "deployments") {
//     name
//     kind
//     namespace
//   }
// }
//
// Single resource query:
// {
//   resource(group: "apps", version: "v1", resource: "deployments", name: "my-app") {
//     name
//     kind
//     namespace
//   }
// }

// Once integrated, you can access:
// - GraphQL API: POST /graphql
// - Schema introspection: GET /graphql/schema

// The GraphQL endpoint will be available at:
// http://localhost:3000/graphql
