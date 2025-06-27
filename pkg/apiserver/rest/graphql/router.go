package graphql

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/grafana/grafana/pkg/api/routing"
	"github.com/grafana/grafana/pkg/middleware"
	contextmodel "github.com/grafana/grafana/pkg/services/contexthandler/model"

	"github.com/grafana/grafana/pkg/apiserver/registry/graphql"
)

// Router provides REST API integration for GraphQL
type Router struct {
	registry *graphql.Registry
}

// NewRouter creates a new GraphQL REST router
func NewRouter(registry *graphql.Registry) *Router {
	return &Router{
		registry: registry,
	}
}

// RegisterRoutes registers GraphQL routes with the routing system
func (r *Router) RegisterRoutes(rr routing.RouteRegister) {
	// Main GraphQL endpoint - POST only (GET is an antipattern)
	rr.Post("/apis/graphql", middleware.ReqSignedIn, r.handleGraphQL)
}

// handleGraphQL handles GraphQL requests (POST only)
func (r *Router) handleGraphQL(c *contextmodel.ReqContext) {
	if c.Req.Method != http.MethodPost {
		c.JsonApiErr(http.StatusMethodNotAllowed, "GraphQL only supports POST requests", nil)
		return
	}

	// Validate Content-Type
	contentType := c.Req.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		c.JsonApiErr(http.StatusUnsupportedMediaType, "Content-Type must be application/json", nil)
		return
	}

	var request graphql.GraphQLRequest
	if err := json.NewDecoder(c.Req.Body).Decode(&request); err != nil {
		c.JsonApiErr(http.StatusBadRequest, "Invalid JSON request body", err)
		return
	}

	// Validate that query is provided
	if strings.TrimSpace(request.Query) == "" {
		c.JsonApiErr(http.StatusBadRequest, "Query is required", nil)
		return
	}

	response := r.registry.ExecuteQuery(c.Req.Context(), request)
	
	// GraphQL always returns 200 OK, even for errors
	// Errors are indicated in the response.Errors field
	c.JSON(http.StatusOK, response)
} 