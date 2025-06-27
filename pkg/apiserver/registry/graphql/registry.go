package graphql

import (
	"context"
	"fmt"
	"sync"

	"github.com/graphql-go/graphql"
	"github.com/grafana/grafana/pkg/infra/log"
)

// GraphQLRequest represents a GraphQL request
type GraphQLRequest struct {
	Query         string                 `json:"query"`
	Variables     map[string]interface{} `json:"variables,omitempty"`
	OperationName string                 `json:"operationName,omitempty"`
}

// GraphQLResponse represents a GraphQL response
type GraphQLResponse struct {
	Data       interface{}   `json:"data,omitempty"`
	Errors     []interface{} `json:"errors,omitempty"`
	Extensions interface{}   `json:"extensions,omitempty"`
}

// AppGraphQLProvider interface that apps must implement to provide GraphQL functionality
type AppGraphQLProvider interface {
	GetGraphQLSchema() (graphql.Schema, error)
	GetAppName() string
	GetResourceCollections() map[string]interface{} // For future resource access
}

// Registry manages GraphQL schemas from multiple apps
type Registry struct {
	mu           sync.RWMutex
	apps         map[string]AppGraphQLProvider
	unifiedSchema graphql.Schema
	needsRebuild bool
	log          log.Logger
}

// NewRegistry creates a new GraphQL registry
func NewRegistry() *Registry {
	registry := &Registry{
		apps:         make(map[string]AppGraphQLProvider),
		needsRebuild: true,
		log:          log.New("apiserver.graphql.registry"),
	}

	// Create initial empty schema
	schema, _ := graphql.NewSchema(graphql.SchemaConfig{
		Query: graphql.NewObject(graphql.ObjectConfig{
			Name: "Query",
			Fields: graphql.Fields{
				"ping": &graphql.Field{
					Type: graphql.String,
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						return "pong", nil
					},
				},
			},
		}),
		Mutation: graphql.NewObject(graphql.ObjectConfig{
			Name:   "Mutation",
			Fields: graphql.Fields{},
		}),
	})
	registry.unifiedSchema = schema

	return registry
}

// RegisterApp registers a GraphQL provider from an app
func (r *Registry) RegisterApp(appName string, provider AppGraphQLProvider) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.apps[appName]; exists {
		return fmt.Errorf("app %s is already registered", appName)
	}

	r.apps[appName] = provider
	r.needsRebuild = true

	r.log.Info("Registered GraphQL app", "app", appName)
	return r.rebuildUnifiedSchema()
}

// UnregisterApp removes a GraphQL provider for an app
func (r *Registry) UnregisterApp(appName string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.apps[appName]; !exists {
		return fmt.Errorf("app %s is not registered", appName)
	}

	delete(r.apps, appName)
	r.needsRebuild = true

	r.log.Info("Unregistered GraphQL app", "app", appName)
	return r.rebuildUnifiedSchema()
}

// ExecuteQuery executes a GraphQL query against the unified schema
func (r *Registry) ExecuteQuery(ctx context.Context, request GraphQLRequest) GraphQLResponse {
	r.mu.RLock()
	schema := r.unifiedSchema
	r.mu.RUnlock()

	// Execute GraphQL query
	result := graphql.Do(graphql.Params{
		Schema:         schema,
		RequestString:  request.Query,
		VariableValues: request.Variables,
		OperationName:  request.OperationName,
		Context:        ctx,
	})

	response := GraphQLResponse{
		Data: result.Data,
	}

	if len(result.Errors) > 0 {
		errors := make([]interface{}, len(result.Errors))
		for i, err := range result.Errors {
			errors[i] = map[string]interface{}{
				"message": err.Error(),
			}
		}
		response.Errors = errors
		r.log.Error("GraphQL execution errors", "errors", result.Errors)
	}

	if len(result.Extensions) > 0 {
		response.Extensions = result.Extensions
	}

	return response
}

// GetUnifiedSchema returns the current unified schema
func (r *Registry) GetUnifiedSchema() graphql.Schema {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.unifiedSchema
}

// ListRegisteredApps returns a list of registered app names
func (r *Registry) ListRegisteredApps() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	apps := make([]string, 0, len(r.apps))
	for appName := range r.apps {
		apps = append(apps, appName)
	}
	return apps
}

// rebuildUnifiedSchema rebuilds the unified schema from all registered apps
func (r *Registry) rebuildUnifiedSchema() error {
	if !r.needsRebuild {
		return nil
	}

	queryFields := graphql.Fields{
		"ping": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return "pong", nil
			},
		},
	}

	mutationFields := graphql.Fields{}

	// First pass: collect all field names and detect conflicts
	allQueryFields := make(map[string][]string)  // fieldName -> []appName
	allMutationFields := make(map[string][]string)

	for appName, provider := range r.apps {
		appSchema, err := provider.GetGraphQLSchema()
		if err != nil {
			r.log.Error("Failed to get schema from app", "app", appName, "error", err)
			continue
		}

		// Collect query field names
		if appSchema.QueryType() != nil {
			for fieldName := range appSchema.QueryType().Fields() {
				allQueryFields[fieldName] = append(allQueryFields[fieldName], appName)
			}
		}

		// Collect mutation field names
		if appSchema.MutationType() != nil {
			for fieldName := range appSchema.MutationType().Fields() {
				allMutationFields[fieldName] = append(allMutationFields[fieldName], appName)
			}
		}
	}

	// Second pass: add fields with smart prefixing
	for appName, provider := range r.apps {
		appSchema, err := provider.GetGraphQLSchema()
		if err != nil {
			continue // Already logged above
		}

		// Add query fields with smart prefixing
		if appSchema.QueryType() != nil {
			for fieldName, fieldDef := range appSchema.QueryType().Fields() {
				finalFieldName := r.getSmartFieldName(fieldName, appName, allQueryFields[fieldName])
				
				// Convert arguments from []*Argument to FieldConfigArgument
				args := graphql.FieldConfigArgument{}
				for _, arg := range fieldDef.Args {
					args[arg.Name()] = &graphql.ArgumentConfig{
						Type:         arg.Type,
						DefaultValue: arg.DefaultValue,
						Description:  arg.Description(),
					}
				}
				
				// Convert FieldDefinition to Field
				queryFields[finalFieldName] = &graphql.Field{
					Type:        fieldDef.Type,
					Args:        args,
					Resolve:     fieldDef.Resolve,
					Description: fieldDef.Description,
				}
			}
		}

		// Add mutation fields with smart prefixing
		if appSchema.MutationType() != nil {
			for fieldName, fieldDef := range appSchema.MutationType().Fields() {
				finalFieldName := r.getSmartFieldName(fieldName, appName, allMutationFields[fieldName])
				
				// Convert arguments from []*Argument to FieldConfigArgument
				args := graphql.FieldConfigArgument{}
				for _, arg := range fieldDef.Args {
					args[arg.Name()] = &graphql.ArgumentConfig{
						Type:         arg.Type,
						DefaultValue: arg.DefaultValue,
						Description:  arg.Description(),
					}
				}
				
				// Convert FieldDefinition to Field
				mutationFields[finalFieldName] = &graphql.Field{
					Type:        fieldDef.Type,
					Args:        args,
					Resolve:     fieldDef.Resolve,
					Description: fieldDef.Description,
				}
			}
		}
	}

	// Build unified schema
	schemaConfig := graphql.SchemaConfig{
		Query: graphql.NewObject(graphql.ObjectConfig{
			Name:   "Query",
			Fields: queryFields,
		}),
	}

	if len(mutationFields) > 0 {
		schemaConfig.Mutation = graphql.NewObject(graphql.ObjectConfig{
			Name:   "Mutation",
			Fields: mutationFields,
		})
	}

	schema, err := graphql.NewSchema(schemaConfig)
	if err != nil {
		return fmt.Errorf("failed to build unified schema: %w", err)
	}

	r.unifiedSchema = schema
	r.needsRebuild = false
	r.log.Info("Rebuilt unified GraphQL schema", "apps", len(r.apps))

	return nil
}

// getSmartFieldName determines the final field name using smart prefixing logic
func (r *Registry) getSmartFieldName(fieldName, appName string, conflictingApps []string) string {
	// If there's no conflict (only one app uses this field name), use the original name
	if len(conflictingApps) <= 1 {
		return fieldName
	}

	// If there are conflicts, use app prefix
	return fmt.Sprintf("%s_%s", appName, fieldName)
} 