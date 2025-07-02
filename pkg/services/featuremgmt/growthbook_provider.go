package featuremgmt

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/infra/log/slogadapter"
	gb "github.com/growthbook/growthbook-golang"
	gbprovider "github.com/growthbook/growthbook-openfeature-provider-go"

	"github.com/open-feature/go-sdk/openfeature"
)

func newGrowthBookProvider(url string, clientKey string) (openfeature.FeatureProvider, error) {
	if url == "" {
		return nil, fmt.Errorf("URL is required for GrowthBook provider")
	}

	if clientKey == "" {
		return nil, fmt.Errorf("ClientKey is required for GrowthBook provider")
	}

	gbClient, err := gb.NewClient(context.TODO(),
		gb.WithClientKey(clientKey),
		gb.WithApiHost(url),
		gb.WithSseDataSource(),
		gb.WithLogger(slog.New(slogadapter.New(log.New("growthbook-provider")))),
	)

	if err != nil {
		return nil, err
	}

	provider := gbprovider.NewProvider(gbClient)
	return provider, nil
}
