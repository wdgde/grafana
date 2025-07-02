package featuremgmt

import (
	"fmt"
	"log/slog"
	"net/url"

	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/infra/log/slogadapter"
	"github.com/grafana/grafana/pkg/setting"

	"github.com/open-feature/go-sdk/openfeature"
	"github.com/open-feature/go-sdk/openfeature/hooks"
)

type OpenFeatureService struct {
	log      log.Logger
	provider openfeature.FeatureProvider
	Client   openfeature.IClient
}

// ProvideOpenFeatureService is used for wiring dependencies in single tenant grafana
func ProvideOpenFeatureService(cfg *setting.Cfg) (*OpenFeatureService, error) {
	confFlags, err := setting.ReadFeatureTogglesFromInitFile(cfg.Raw.Section("feature_toggles"))
	if err != nil {
		return nil, fmt.Errorf("failed to read feature toggles from config: %w", err)
	}

	openfeature.SetEvaluationContext(openfeature.NewEvaluationContext(cfg.OpenFeature.TargetingKey, cfg.OpenFeature.ContextAttrs))
	return newOpenFeatureService(cfg.OpenFeature.ProviderType, cfg.OpenFeature.URL, cfg.OpenFeature.ClientKey, confFlags)
}

// TODO: might need to be public, so other MT services could set up open feature client
func newOpenFeatureService(pType string, u *url.URL, key string, staticFlags map[string]bool) (*OpenFeatureService, error) {
	p, err := createProvider(pType, u, key, staticFlags)
	if err != nil {
		return nil, fmt.Errorf("failed to create feature provider: type %s, %w", pType, err)
	}

	loggingHook, err := hooks.NewCustomLoggingHook(false, slog.New(slogadapter.New(log.New("openfeature"))))
	if err != nil {
		return nil, fmt.Errorf("failed to create logging hook: %w", err)
	}

	if err := openfeature.SetProviderAndWait(p); err != nil {
		return nil, fmt.Errorf("failed to set global feature provider: %s, %w", pType, err)
	}

	client := openfeature.NewClient("grafana-openfeature-client")
	client.AddHooks(loggingHook)
	return &OpenFeatureService{
		log:      log.New("openfeatureservice"),
		provider: p,
		Client:   client,
	}, nil
}

func createProvider(providerType string, u *url.URL, clientKey string, staticFlags map[string]bool) (openfeature.FeatureProvider, error) {
	if providerType == setting.StaticProviderType {
		return newStaticProvider(staticFlags)
	}

	if providerType == setting.GOFFProviderType {
		return newGOFFProvider(u.String())
	}

	if providerType == setting.GrowthBookProviderType {
		return newGrowthBookProvider(u.String(), clientKey)
	}

	return nil, fmt.Errorf("invalid provider type: %s", providerType)
}

func createClient(provider openfeature.FeatureProvider) (openfeature.IClient, error) {
	if err := openfeature.SetProviderAndWait(provider); err != nil {
		return nil, fmt.Errorf("failed to set global feature provider: %w", err)
	}

	client := openfeature.NewClient("grafana-openfeature-client")
	return client, nil
}
