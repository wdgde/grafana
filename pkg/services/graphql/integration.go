package graphql

import (
	"github.com/grafana/grafana/pkg/api/routing"
	"github.com/grafana/grafana/pkg/web"
)

// RegisterRoutes registers GraphQL routes with the Grafana HTTP server
func RegisterRoutes(routeRegister routing.RouteRegister, graphqlService *Service) {
	// Register GraphQL endpoint with introspection support
	routeRegister.Post("/graphql", func(c *web.Context) {
		graphqlService.HandleGraphQL(c.Resp, c.Req)
	})

	// Note: Schema introspection is handled through the main GraphQL endpoint
	// using standard introspection queries like { __schema { types { name } } }
}

// Example usage:
// In your main HTTP server setup, you would do:
//
// graphqlService, err := NewService()
// if err != nil {
//     // handle error
// }
//
// RegisterRoutes(hs.RouteRegister, graphqlService)
