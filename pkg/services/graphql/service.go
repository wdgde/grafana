package graphql

import (
	"context"
	"encoding/json"
	"net/http"

	contextmodel "github.com/grafana/grafana/pkg/services/contexthandler/model"
	"github.com/grafana/grafana/pkg/web"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
)

// Service represents the GraphQL service
type Service struct {
	schema    graphql.Schema
	apiClient *APIClient
}

// NewService creates a new GraphQL service
func NewService() (*Service, error) {
	// Create API client (assuming Grafana is running on localhost:3000)
	apiClient := NewAPIClient("http://localhost:3000")

	// Define dashboard metadata type
	dashboardMetadataType := graphql.NewObject(graphql.ObjectConfig{
		Name: "DashboardMetadata",
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
			"creationTimestamp": &graphql.Field{
				Type: graphql.String,
			},
			"labels": &graphql.Field{
				Type: graphql.NewScalar(graphql.ScalarConfig{
					Name:         "JSON",
					Description:  "JSON scalar type",
					Serialize:    func(value interface{}) interface{} { return value },
					ParseValue:   func(value interface{}) interface{} { return value },
					ParseLiteral: func(valueAST ast.Value) interface{} { return nil },
				}),
			},
			"annotations": &graphql.Field{
				Type: graphql.NewScalar(graphql.ScalarConfig{
					Name:         "JSONAnnotations",
					Description:  "JSON scalar type for annotations",
					Serialize:    func(value interface{}) interface{} { return value },
					ParseValue:   func(value interface{}) interface{} { return value },
					ParseLiteral: func(valueAST ast.Value) interface{} { return nil },
				}),
			},
		},
	})

	// Define dashboard spec type
	dashboardSpecType := graphql.NewObject(graphql.ObjectConfig{
		Name: "DashboardSpec",
		Fields: graphql.Fields{
			"title": &graphql.Field{
				Type: graphql.String,
			},
			"description": &graphql.Field{
				Type: graphql.String,
			},
			"tags": &graphql.Field{
				Type: graphql.NewList(graphql.String),
			},
			"dashboard": &graphql.Field{
				Type: graphql.NewScalar(graphql.ScalarConfig{
					Name:         "DashboardJSON",
					Description:  "Dashboard JSON data",
					Serialize:    func(value interface{}) interface{} { return value },
					ParseValue:   func(value interface{}) interface{} { return value },
					ParseLiteral: func(valueAST ast.Value) interface{} { return nil },
				}),
			},
		},
	})

	// Define dashboard type
	dashboardType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Dashboard",
		Fields: graphql.Fields{
			"apiVersion": &graphql.Field{
				Type: graphql.String,
			},
			"kind": &graphql.Field{
				Type: graphql.String,
			},
			"metadata": &graphql.Field{
				Type: dashboardMetadataType,
			},
			"spec": &graphql.Field{
				Type: dashboardSpecType,
			},
		},
	})

	// Define a generic resource type for other resources
	resourceType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Resource",
		Fields: graphql.Fields{
			"name": &graphql.Field{
				Type: graphql.String,
			},
			"kind": &graphql.Field{
				Type: graphql.String,
			},
			"namespace": &graphql.Field{
				Type: graphql.String,
			},
		},
	})

	// Define the root query
	rootQuery := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			// Dashboard-specific queries
			"dashboards": &graphql.Field{
				Type: graphql.NewList(dashboardType),
				Args: graphql.FieldConfigArgument{
					"namespace": &graphql.ArgumentConfig{
						Type:        graphql.NewNonNull(graphql.String),
						Description: "Namespace to search for dashboards",
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
					dashboardList, err := apiClient.GetDashboards(p.Context, namespace, reqCtx)
					if err != nil {
						return nil, err
					}

					// Convert to GraphQL-friendly format
					var dashboards []map[string]interface{}
					for _, dashboard := range dashboardList.Items {
						dashboards = append(dashboards, map[string]interface{}{
							"apiVersion": dashboard.APIVersion,
							"kind":       dashboard.Kind,
							"metadata": map[string]interface{}{
								"name":              dashboard.Metadata.Name,
								"namespace":         dashboard.Metadata.Namespace,
								"uid":               dashboard.Metadata.UID,
								"creationTimestamp": dashboard.Metadata.CreationTime,
								"labels":            dashboard.Metadata.Labels,
								"annotations":       dashboard.Metadata.Annotations,
							},
							"spec": map[string]interface{}{
								"title":       dashboard.Spec.Title,
								"description": dashboard.Spec.Description,
								"tags":        dashboard.Spec.Tags,
								"dashboard":   dashboard.Spec.Dashboard,
							},
						})
					}

					return dashboards, nil
				},
			},
			"dashboard": &graphql.Field{
				Type: dashboardType,
				Args: graphql.FieldConfigArgument{
					"namespace": &graphql.ArgumentConfig{
						Type:        graphql.NewNonNull(graphql.String),
						Description: "Namespace of the dashboard",
					},
					"name": &graphql.ArgumentConfig{
						Type:        graphql.NewNonNull(graphql.String),
						Description: "Name of the dashboard",
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					namespace := p.Args["namespace"].(string)
					name := p.Args["name"].(string)

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
					dashboard, err := apiClient.GetDashboard(p.Context, namespace, name, reqCtx)
					if err != nil {
						return nil, err
					}

					// Convert to GraphQL-friendly format
					return map[string]interface{}{
						"apiVersion": dashboard.APIVersion,
						"kind":       dashboard.Kind,
						"metadata": map[string]interface{}{
							"name":              dashboard.Metadata.Name,
							"namespace":         dashboard.Metadata.Namespace,
							"uid":               dashboard.Metadata.UID,
							"creationTimestamp": dashboard.Metadata.CreationTime,
							"labels":            dashboard.Metadata.Labels,
							"annotations":       dashboard.Metadata.Annotations,
						},
						"spec": map[string]interface{}{
							"title":       dashboard.Spec.Title,
							"description": dashboard.Spec.Description,
							"tags":        dashboard.Spec.Tags,
							"dashboard":   dashboard.Spec.Dashboard,
						},
					}, nil
				},
			},
			// Generic resources query (for other resource types)
			"resources": &graphql.Field{
				Type: graphql.NewList(resourceType),
				Args: graphql.FieldConfigArgument{
					"group": &graphql.ArgumentConfig{
						Type:        graphql.String,
						Description: "API group",
					},
					"version": &graphql.ArgumentConfig{
						Type:        graphql.String,
						Description: "API version",
					},
					"namespace": &graphql.ArgumentConfig{
						Type:        graphql.String,
						Description: "Namespace (optional)",
					},
					"resource": &graphql.ArgumentConfig{
						Type:        graphql.String,
						Description: "Resource type",
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					group, _ := p.Args["group"].(string)
					version, _ := p.Args["version"].(string)
					namespace, _ := p.Args["namespace"].(string)
					resource, _ := p.Args["resource"].(string)

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

					// For backwards compatibility, return mock data if no specific parameters
					if group == "" && version == "" && resource == "" {
						return []map[string]interface{}{
							{
								"name":      "example-resource",
								"kind":      "Deployment",
								"namespace": "default",
							},
						}, nil
					}

					// Call the real API for generic resources
					result, err := apiClient.GetResources(p.Context, group, version, namespace, resource, reqCtx)
					if err != nil {
						return nil, err
					}

					// Try to extract items from the result (assuming it's a list response)
					if resultMap, ok := result.(map[string]interface{}); ok {
						if items, exists := resultMap["items"]; exists {
							return items, nil
						}
					}

					return result, nil
				},
			},
		},
	})

	// Create the schema
	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: rootQuery,
	})
	if err != nil {
		return nil, err
	}

	return &Service{
		schema:    schema,
		apiClient: apiClient,
	}, nil
}

// HandleGraphQL handles GraphQL HTTP requests
func (s *Service) HandleGraphQL(w http.ResponseWriter, r *http.Request) {
	var requestBody struct {
		Query     string                 `json:"query"`
		Variables map[string]interface{} `json:"variables"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// For now, we'll use the request context directly
	// TODO: Properly integrate with Grafana's request context
	ctx := context.WithValue(r.Context(), "httpRequest", r)

	result := graphql.Do(graphql.Params{
		Schema:         s.schema,
		RequestString:  requestBody.Query,
		VariableValues: requestBody.Variables,
		Context:        ctx,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
