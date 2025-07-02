package setting

import (
	"fmt"
	"net/url"
)

const (
	StaticProviderType     = "static"
	GOFFProviderType       = "goff"
	GrowthBookProviderType = "growthbook"
)

type OpenFeatureSettings struct {
	ProviderType string         `json:"providerType"`
	URL          *url.URL       `json:"url"`
	ClientKey    string         `json:"clientKey"`
	TargetingKey string         `json:"targetingKey"`
	ContextAttrs map[string]any `json:"contextAttrs"`
}

func (cfg *Cfg) readOpenFeatureSettings() error {
	cfg.OpenFeature = OpenFeatureSettings{}

	config := cfg.Raw.Section("feature_toggles.openfeature")
	cfg.OpenFeature.ProviderType = config.Key("provider").MustString(StaticProviderType)
	cfg.OpenFeature.TargetingKey = config.Key("targetingKey").MustString(cfg.AppURL)
	cfg.OpenFeature.ClientKey = config.Key("clientKey").MustString("")

	strURL := config.Key("url").MustString("")

	if strURL != "" && (cfg.OpenFeature.ProviderType == GOFFProviderType || cfg.OpenFeature.ProviderType == GrowthBookProviderType) {
		u, err := url.Parse(strURL)
		if err != nil {
			return fmt.Errorf("invalid feature provider url: %w", err)
		}
		cfg.OpenFeature.URL = u
	}

	// build the eval context attributes using [feature_toggles.openfeature.context] section
	ctxConf := cfg.Raw.Section("feature_toggles.openfeature.context")
	attrs := map[string]any{}
	for _, key := range ctxConf.KeyStrings() {
		attrs[key] = ctxConf.Key(key).String()
	}

	// Some default attributes
	if _, ok := attrs["grafana_version"]; !ok {
		attrs["grafana_version"] = cfg.BuildVersion
	}

	cfg.OpenFeature.ContextAttrs = attrs
	return nil
}
