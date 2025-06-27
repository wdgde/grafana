package apiserver

import (
	"context"
	"fmt"

	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/services/featuremgmt"
	graphqlregistry "github.com/grafana/grafana/pkg/apiserver/registry/graphql"
	graphqlrest "github.com/grafana/grafana/pkg/apiserver/rest/graphql"
	investigationapp "github.com/grafana/grafana/apps/investigations/pkg/app"
	"github.com/grafana/grafana-app-sdk/resource"
	"github.com/grafana/grafana-app-sdk/plugin/router"
)

// GraphQLService manages GraphQL functionality for the API server
type GraphQLService struct {
	log      log.Logger
	features featuremgmt.FeatureToggles
	registry *graphqlregistry.Registry
	router   *graphqlrest.Router
}

// ProvideGraphQLService creates a new GraphQL service
func ProvideGraphQLService(features featuremgmt.FeatureToggles) *GraphQLService {
	log := log.New("graphql")
	registry := graphqlregistry.NewRegistry()
	router := graphqlrest.NewRouter(registry)
	
	return &GraphQLService{
		log:      log,
		features: features,
		registry: registry,
		router:   router,
	}
}

// IsEnabled returns whether GraphQL is enabled
func (s *GraphQLService) IsEnabled() bool {
	return s.features.IsEnabledGlobally(featuremgmt.FlagApiServerGraphQL)
}

// Run starts the GraphQL service
func (s *GraphQLService) Run(ctx context.Context) error {
	if !s.IsEnabled() {
		s.log.Debug("GraphQL service is disabled via feature flag")
		return nil
	}

	s.log.Info("Starting GraphQL service")

	// TODO: For testing, manually register the investigations app
	// This should be replaced with proper app discovery
	s.registerTestInvestigationsApp()

	// Keep service running
	<-ctx.Done()
	s.log.Info("GraphQL service stopped")
	return nil
}

// registerTestInvestigationsApp manually registers the investigations app for testing
func (s *GraphQLService) registerTestInvestigationsApp() {
	s.log.Info("Registering investigations GraphQL provider")
	
	// For testing, create a simplified GraphQL provider directly
	// This bypasses the need for a full app instance with KubeConfig
	
	// Create resource collection for investigations
	investigationKinds := investigationapp.GetKinds()
	s.log.Info("Found investigation kinds", "count", len(investigationKinds))
	
	// Create a simple test provider using the app SDK components
	provider, err := s.createTestInvestigationsProvider()
	if err != nil {
		s.log.Error("Failed to create investigations GraphQL provider", "error", err)
		return
	}
	
	// Register the provider
	err = s.RegisterApp("investigations", provider)
	if err != nil {
		s.log.Error("Failed to register investigations GraphQL provider", "error", err)
	} else {
		s.log.Info("Successfully registered investigations GraphQL provider")
	}
}

// createTestInvestigationsProvider creates a test GraphQL provider for investigations
func (s *GraphQLService) createTestInvestigationsProvider() (graphqlregistry.AppGraphQLProvider, error) {
	// Create resource collection from investigations kinds
	kinds := investigationapp.GetKinds()
	var resourceKinds []resource.Kind
	
	for _, kindList := range kinds {
		for _, kind := range kindList {
			resourceKinds = append(resourceKinds, resource.Kind{
				Schema: kind,
				Codecs: map[resource.KindEncoding]resource.Codec{
					resource.KindEncodingJSON: resource.NewJSONCodec(),
				},
			})
		}
	}
	
	resourceCollection := resource.NewKindCollection(resourceKinds...)
	
	// Create a mock store for testing (in production this would be a real store)
	mockStore := &mockStore{}
	
	// Create the GraphQL provider
	provider, err := router.NewAppGraphQLProvider("investigations", resourceCollection, mockStore)
	if err != nil {
		return nil, fmt.Errorf("failed to create GraphQL provider: %w", err)
	}
	
	return provider, nil
}

// mockStore is a simple mock implementation of the Store interface for testing
type mockStore struct{}

func (m *mockStore) Add(ctx context.Context, obj resource.Object) (resource.Object, error) {
	return obj, nil
}

func (m *mockStore) Get(ctx context.Context, kind string, identifier resource.Identifier) (resource.Object, error) {
	return &resource.UntypedObject{}, fmt.Errorf("not found")
}

func (m *mockStore) List(ctx context.Context, kind string, options resource.StoreListOptions) (resource.ListObject, error) {
	return &resource.UntypedList{}, nil
}

func (m *mockStore) Update(ctx context.Context, obj resource.Object) (resource.Object, error) {
	return obj, nil
}

func (m *mockStore) Delete(ctx context.Context, kind string, identifier resource.Identifier) error {
	return nil
}

// RegisterApp registers an app's GraphQL provider
func (s *GraphQLService) RegisterApp(name string, provider graphqlregistry.AppGraphQLProvider) error {
	if !s.IsEnabled() {
		return fmt.Errorf("GraphQL service is not enabled")
	}

	s.log.Info("Registering GraphQL app", "app", name)
	return s.registry.RegisterApp(name, provider)
}

// UnregisterApp removes a GraphQL provider for an app
func (s *GraphQLService) UnregisterApp(appName string) error {
	if !s.IsEnabled() {
		return fmt.Errorf("GraphQL service is not enabled")
	}

	s.log.Info("Unregistering GraphQL app", "app", appName)
	return s.registry.UnregisterApp(appName)
}

// GetRegistry returns the GraphQL registry for route registration
func (s *GraphQLService) GetRegistry() *graphqlregistry.Registry {
	return s.registry
}

// HealthCheck performs a health check on the GraphQL service
func (s *GraphQLService) HealthCheck(ctx context.Context) error {
	if !s.IsEnabled() {
		return fmt.Errorf("GraphQL service is disabled")
	}

	return nil
}

// Integration point for the main API server service
func (s *service) initGraphQL() error {
	if s.features.IsEnabled(context.Background(), featuremgmt.FlagApiServerGraphQL) {
		s.log.Info("Initializing GraphQL support")
		
		// Create GraphQL router
		router := graphqlrest.NewRouter(s.graphqlService.GetRegistry())
		
		// Register GraphQL routes with the API server
		router.RegisterRoutes(s.rr)
		
		s.log.Info("GraphQL routes registered")
		
		// Register test investigations app
		s.graphqlService.registerTestInvestigationsApp()
	}
	return nil
}

// Feature flag definition (this would go in featuremgmt package, shown here for completeness)
// const FlagApiServerGraphQL = "apiServerGraphQL"

// RegisterAppsFromRegistry discovers and registers GraphQL providers from the app registry
func (s *GraphQLService) RegisterAppsFromRegistry() error {
	if !s.IsEnabled() {
		return fmt.Errorf("GraphQL service is not enabled")
	}

	s.log.Info("Discovering GraphQL providers from app registry")
	
	// TODO: This is a placeholder for discovering apps that implement GraphQL providers
	// In a real implementation, this would iterate through registered apps and check
	// if they implement a GraphQL provider interface
	
	return nil
} 