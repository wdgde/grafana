package graphql

import (
	"context"
	"net/http"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/endpoints/handlers/responsewriters"
	"k8s.io/apiserver/pkg/endpoints/request"

	"github.com/grafana/grafana/pkg/apiserver/registry/graphql"
	"github.com/grafana/grafana/pkg/apiserver/rest/graphql"
	"github.com/grafana/grafana/pkg/apimachinery/identity"
	"github.com/grafana/grafana/pkg/infra/log"
)

// Handler provides GraphQL endpoint handling within the API server framework
type Handler struct {
	registry   *graphql.Registry
	restRouter *graphqlrest.Router
	serializer runtime.NegotiatedSerializer
	log        log.Logger
}

// NewHandler creates a new GraphQL endpoint handler
func NewHandler(registry *graphql.Registry, serializer runtime.NegotiatedSerializer) *Handler {
	restRouter := graphqlrest.NewRouter(registry, serializer)
	
	return &Handler{
		registry:   registry,
		restRouter: restRouter,
		serializer: serializer,
		log:        log.New("apiserver.endpoints.graphql"),
	}
}

// ServeHTTP handles HTTP requests for GraphQL endpoints
func (h *Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	
	// Extract request info if available (following Grafana patterns)
	if requestInfo, found := request.RequestInfoFrom(ctx); found {
		h.log.Debug("GraphQL request", 
			"path", requestInfo.Path,
			"verb", requestInfo.Verb,
			"resource", requestInfo.Resource,
		)
	}

	// Extract user info if available
	var userInfo identity.Requester
	if user, found := identity.RequesterFrom(ctx); found {
		userInfo = user
		h.log.Debug("GraphQL request user", 
			"userID", userInfo.GetUID(),
			"login", userInfo.GetLogin(),
		)
	}

	// Delegate to REST router which handles the actual logic
	h.restRouter.ServeHTTP(w, req)
}

// HealthCheck provides health checking for the GraphQL endpoint
func (h *Handler) HealthCheck(ctx context.Context) error {
	// Check if registry is accessible
	apps := h.registry.ListRegisteredApps()
	h.log.Debug("GraphQL health check", "registered_apps", len(apps))
	
	// Try to get the schema to ensure the registry is working
	_, err := h.registry.GetSchemaSDL(ctx)
	if err != nil {
		h.log.Error("GraphQL health check failed", "error", err)
		return err
	}
	
	return nil
}

// GetRegistry returns the GraphQL registry for external access
func (h *Handler) GetRegistry() *graphql.Registry {
	return h.registry
}

// handleIntrospection handles GraphQL introspection queries
func (h *Handler) handleIntrospection(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	
	// Standard GraphQL introspection query
	introspectionQuery := `
		query IntrospectionQuery {
			__schema {
				queryType { name }
				mutationType { name }
				subscriptionType { name }
				types {
					...FullType
				}
			}
		}
		
		fragment FullType on __Type {
			kind
			name
			description
			fields(includeDeprecated: true) {
				name
				description
				type { ...TypeRef }
				isDeprecated
				deprecationReason
			}
			inputFields {
				...InputValue
			}
			interfaces { ...TypeRef }
			enumValues(includeDeprecated: true) {
				name
				description
				isDeprecated
				deprecationReason
			}
			possibleTypes { ...TypeRef }
		}
		
		fragment InputValue on __InputValue {
			name
			description
			type { ...TypeRef }
			defaultValue
		}
		
		fragment TypeRef on __Type {
			kind
			name
			ofType {
				kind
				name
				ofType {
					kind
					name
					ofType {
						kind
						name
						ofType {
							kind
							name
							ofType {
								kind
								name
								ofType {
									kind
									name
									ofType {
										kind
										name
									}
								}
							}
						}
					}
				}
			}
		}
	`

	request := graphql.GraphQLRequest{
		Query: introspectionQuery,
	}

	response := h.registry.ExecuteQuery(ctx, request)
	
	// Write response using API server response writers
	responsewriters.WriteObjectNegotiated(
		h.serializer,
		responsewriters.DefaultNegotiation,
		responsewriters.DefaultScheme,
		responsewriters.DefaultGVK,
		w,
		req,
		http.StatusOK,
		response,
		false,
	)
}

// Metrics provides metrics about GraphQL operations
type Metrics struct {
	TotalQueries      int64            `json:"total_queries"`
	ErrorCount        int64            `json:"error_count"`
	AvgResponseTime   float64          `json:"avg_response_time_ms"`
	RegisteredApps    []string         `json:"registered_apps"`
	SchemaComplexity  int              `json:"schema_complexity"`
}

// GetMetrics returns operational metrics for the GraphQL endpoint
func (h *Handler) GetMetrics(ctx context.Context) (*Metrics, error) {
	apps := h.registry.ListRegisteredApps()
	schema := h.registry.GetUnifiedSchema()
	
	// Calculate schema complexity (simplified)
	complexity := 0
	if schema.QueryType() != nil {
		complexity += len(schema.QueryType().Fields())
	}
	if schema.MutationType() != nil {
		complexity += len(schema.MutationType().Fields())
	}

	metrics := &Metrics{
		RegisteredApps:   apps,
		SchemaComplexity: complexity,
		// TODO: Implement actual metrics collection for queries, errors, response times
		TotalQueries:    0,
		ErrorCount:      0,
		AvgResponseTime: 0.0,
	}

	return metrics, nil
} 