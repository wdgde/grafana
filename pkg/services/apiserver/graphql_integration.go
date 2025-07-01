package apiserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/graphql-go/graphql"
	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana-app-sdk/plugin/router"
)

// GraphQLProviderApp interface for apps that provide GraphQL functionality
type GraphQLProviderApp interface {
	GetAppName() string
	GetGraphQLProvider() *router.AppGraphQLProvider
}

// AppProvider represents an app that can provide GraphQL capabilities
type AppProvider interface {
	GetAppName() string
}

// GraphQLService provides GraphQL endpoints for the API server
type GraphQLService struct {
	logger      log.Logger
	schema      graphql.Schema
	mutex       sync.RWMutex
	appRegistry map[string]AppProvider
}

// NewGraphQLService creates a new GraphQL service with auto-discovery
func NewGraphQLService() *GraphQLService {
	logger := log.New("graphql-service")
	service := &GraphQLService{
		logger:      logger,
		appRegistry: make(map[string]AppProvider),
	}
	
	// Initialize with basic schema
	service.initializeBaseSchema()
	
	return service
}

// RegisterAppProvider registers an app provider for automatic GraphQL support
func (s *GraphQLService) RegisterAppProvider(provider AppProvider) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	appName := provider.GetAppName()
	s.logger.Info("Registering app provider for GraphQL", "app", appName)
	
	// Check if provider also implements GraphQLProviderApp interface
	if graphqlProvider, ok := provider.(GraphQLProviderApp); ok {
		s.logger.Info("App provider implements GraphQLProviderApp interface", "app", appName)
		if gqlProvider := graphqlProvider.GetGraphQLProvider(); gqlProvider != nil {
			s.logger.Info("Successfully obtained GraphQL provider from app", "app", appName)
		} else {
			s.logger.Warn("App provider returned nil GraphQL provider (app may not be initialized yet)", "app", appName)
		}
	} else {
		s.logger.Info("App provider does not implement GraphQLProviderApp interface", "app", appName)
	}
	
	s.appRegistry[appName] = provider
	
	// Rebuild schema with all registered apps
	// Note: This will skip apps that aren't ready yet, but they can be added later
	return s.rebuildSchema()
}

// UnregisterAppProvider removes an app provider
func (s *GraphQLService) UnregisterAppProvider(appName string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	delete(s.appRegistry, appName)
	s.logger.Info("Unregistered app provider from GraphQL", "app", appName)
	
	return s.rebuildSchema()
}

// HandleGraphQL handles GraphQL requests
func (s *GraphQLService) HandleGraphQL(w http.ResponseWriter, r *http.Request) {
	s.mutex.RLock()
	schema := s.schema
	s.mutex.RUnlock()
	
	var request struct {
		Query         string                 `json:"query"`
		Variables     map[string]interface{} `json:"variables"`
		OperationName string                 `json:"operationName"`
	}
	
	if r.Method == http.MethodGet {
		// Handle GET requests with query parameters
		request.Query = r.URL.Query().Get("query")
		request.OperationName = r.URL.Query().Get("operationName")
		if variables := r.URL.Query().Get("variables"); variables != "" {
			json.Unmarshal([]byte(variables), &request.Variables)
		}
	} else {
		// Handle POST requests with JSON body
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
			return
		}
	}
	
	// Execute GraphQL query
	result := graphql.Do(graphql.Params{
		Schema:         schema,
		RequestString:  request.Query,
		VariableValues: request.Variables,
		OperationName:  request.OperationName,
		Context:        r.Context(),
	})
	
	// Check if there are "Cannot query field" errors, which might indicate apps have become available
	hasFieldErrors := false
	if len(result.Errors) > 0 {
		for _, err := range result.Errors {
			if strings.Contains(err.Error(), "Cannot query field") {
				hasFieldErrors = true
				break
			}
		}
	}
	
	// If we have field errors, try refreshing the schema and re-executing
	if hasFieldErrors {
		s.logger.Info("GraphQL query failed with field errors, attempting schema refresh")
		if refreshErr := s.RefreshSchema(); refreshErr == nil {
			s.logger.Info("Schema refreshed, retrying GraphQL query")
			// Get the updated schema and retry
			s.mutex.RLock()
			updatedSchema := s.schema
			s.mutex.RUnlock()
			
			result = graphql.Do(graphql.Params{
				Schema:         updatedSchema,
				RequestString:  request.Query,
				VariableValues: request.Variables,
				OperationName:  request.OperationName,
				Context:        r.Context(),
			})
		} else {
			s.logger.Warn("Failed to refresh GraphQL schema", "error", refreshErr)
		}
	}
	
	// Prepare response
	response := map[string]interface{}{
		"data": result.Data,
	}
	
	if len(result.Errors) > 0 {
		errors := make([]interface{}, len(result.Errors))
		for i, err := range result.Errors {
			errors[i] = map[string]interface{}{
				"message": err.Error(),
			}
		}
		response["errors"] = errors
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// initializeBaseSchema creates the base GraphQL schema
func (s *GraphQLService) initializeBaseSchema() {
	queryFields := graphql.Fields{
		"ping": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return "pong", nil
			},
		},
	}
	
	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: graphql.NewObject(graphql.ObjectConfig{
			Name:   "Query",
			Fields: queryFields,
		}),
	})
	
	if err != nil {
		s.logger.Error("Failed to create base GraphQL schema", "error", err)
		return
	}
	
	s.schema = schema
}

// rebuildSchema rebuilds the GraphQL schema with all registered apps
func (s *GraphQLService) rebuildSchema() error {
	queryFields := graphql.Fields{
		"ping": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return "pong", nil
			},
		},
	}
	
	mutationFields := graphql.Fields{}
	
	// Use app's actual GraphQL providers instead of manual schema generation
	for appName, provider := range s.appRegistry {
		s.logger.Info("Building GraphQL schema for app", "app", appName)
		
		// Check if provider implements GraphQLProviderApp interface
		graphqlProvider, ok := provider.(GraphQLProviderApp)
		if !ok {
			s.logger.Info("App provider doesn't implement GraphQLProviderApp interface, skipping", "app", appName)
			continue
		}
		
		// Get the app's GraphQL provider
		appGraphQLProvider := graphqlProvider.GetGraphQLProvider()
		if appGraphQLProvider == nil {
			s.logger.Info("App GraphQL provider not available yet, skipping", "app", appName)
			continue
		}
		
		// Get the proper GraphQL schema from the app
		appSchema, err := appGraphQLProvider.GetGraphQLSchema()
		if err != nil {
			s.logger.Error("Failed to get GraphQL schema from app", "app", appName, "error", err)
			continue
		}
		
		s.logger.Info("Successfully obtained GraphQL schema from app", "app", appName)
		
		// Merge the app's schema fields into the unified schema with conflict resolution
		if err := s.mergeAppSchema(appSchema, appName, queryFields, mutationFields); err != nil {
			s.logger.Error("Failed to merge app schema", "app", appName, "error", err)
			continue
		}
		
		s.logger.Info("Successfully merged GraphQL schema from app", "app", appName)
	}
	
	// Create the new schema
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
		return fmt.Errorf("failed to create GraphQL schema: %w", err)
	}
	
	s.schema = schema
	s.logger.Info("Successfully rebuilt GraphQL schema", "apps", len(s.appRegistry))
	
	return nil
}

// mergeAppSchema merges an app's GraphQL schema into the unified schema with conflict resolution
func (s *GraphQLService) mergeAppSchema(appSchema graphql.Schema, appName string, queryFields, mutationFields graphql.Fields) error {
	// First pass: collect all field names and detect conflicts
	allQueryFields := make(map[string][]string)  // fieldName -> []appName
	allMutationFields := make(map[string][]string)
	
	// Collect existing field names from unified schema
	for fieldName := range queryFields {
		allQueryFields[fieldName] = []string{"unified"}
	}
	for fieldName := range mutationFields {
		allMutationFields[fieldName] = []string{"unified"}
	}
	
	// Collect field names from other registered apps
	for otherAppName, otherProvider := range s.appRegistry {
		if otherAppName == appName {
			continue // Skip self
		}
		
		if otherGraphqlProvider, ok := otherProvider.(GraphQLProviderApp); ok {
			if otherAppGraphQLProvider := otherGraphqlProvider.GetGraphQLProvider(); otherAppGraphQLProvider != nil {
				if otherSchema, err := otherAppGraphQLProvider.GetGraphQLSchema(); err == nil {
					// Collect query field names
					if otherSchema.QueryType() != nil {
						for fieldName := range otherSchema.QueryType().Fields() {
							allQueryFields[fieldName] = append(allQueryFields[fieldName], otherAppName)
						}
					}
					
					// Collect mutation field names  
					if otherSchema.MutationType() != nil {
						for fieldName := range otherSchema.MutationType().Fields() {
							allMutationFields[fieldName] = append(allMutationFields[fieldName], otherAppName)
						}
					}
				}
			}
		}
	}
	
	// Add query fields with smart prefixing
	if appSchema.QueryType() != nil {
		for fieldName, fieldDef := range appSchema.QueryType().Fields() {
			// Skip built-in GraphQL fields
			if strings.HasPrefix(fieldName, "__") {
				continue
			}
			
			finalFieldName := s.getSmartFieldName(fieldName, appName, allQueryFields[fieldName])
			
			// Convert arguments from []*graphql.Argument to graphql.FieldConfigArgument
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
			
			s.logger.Info("Added query field from app", "app", appName, "original", fieldName, "final", finalFieldName)
		}
	}
	
	// Add mutation fields with smart prefixing
	if appSchema.MutationType() != nil {
		for fieldName, fieldDef := range appSchema.MutationType().Fields() {
			finalFieldName := s.getSmartFieldName(fieldName, appName, allMutationFields[fieldName])
			
			// Convert arguments from []*graphql.Argument to graphql.FieldConfigArgument
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
			
			s.logger.Info("Added mutation field from app", "app", appName, "original", fieldName, "final", finalFieldName)
		}
	}
	
	return nil
}

// getSmartFieldName generates clean field names with conflict resolution
func (s *GraphQLService) getSmartFieldName(fieldName, appName string, conflictingApps []string) string {
	// If no conflicts, use the original field name
	if len(conflictingApps) <= 1 {
		return fieldName
	}
	
	// If there are conflicts, prefix with app name
	resolvedName := appName + "_" + fieldName
	s.logger.Info("Field name conflict resolved", "original", fieldName, "resolved", resolvedName, "app", appName, "conflicts", len(conflictingApps))
	
	return resolvedName
}

// RefreshSchema rebuilds the GraphQL schema, useful when apps become available
func (s *GraphQLService) RefreshSchema() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	s.logger.Info("Refreshing GraphQL schema for all registered apps")
	return s.rebuildSchema()
}

// GetRegisteredApps returns the list of registered app names
func (s *GraphQLService) GetRegisteredApps() []string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	apps := make([]string, 0, len(s.appRegistry))
	for appName := range s.appRegistry {
		apps = append(apps, appName)
	}
	return apps
}

// IsEnabled returns whether GraphQL is enabled
func (s *GraphQLService) IsEnabled() bool {
	return true // GraphQL is enabled when the service exists
} 