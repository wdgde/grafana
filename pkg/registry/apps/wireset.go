package appregistry

import (
	"github.com/google/wire"

	"github.com/grafana/grafana/apps/advisor/pkg/app/checkregistry"
	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/registry/apps/advisor"
	"github.com/grafana/grafana/pkg/registry/apps/alerting/notifications"
	"github.com/grafana/grafana/pkg/registry/apps/investigations"
	"github.com/grafana/grafana/pkg/registry/apps/playlist"
	"github.com/grafana/grafana/pkg/services/apiserver"
)

// ProvideAppProviders provides the slice of AppProviders for GraphQL auto-discovery
func ProvideAppProviders(
	investigationsProvider *investigations.InvestigationsAppProvider,
) []apiserver.AppProvider {
	logger := log.New("app-providers")
	logger.Info("ProvideAppProviders called", "investigationsProvider", investigationsProvider != nil)
	providers := []apiserver.AppProvider{
		investigationsProvider,
	}
	logger.Info("Returning app providers", "count", len(providers))
	return providers
}

var WireSet = wire.NewSet(
	ProvideRegistryServiceSink,
	playlist.RegisterApp,
	investigations.RegisterApp,
	advisor.RegisterApp,
	checkregistry.ProvideService,
	notifications.RegisterApp,
	wire.Bind(new(checkregistry.CheckService), new(*checkregistry.Service)),
	
	// App providers for GraphQL auto-discovery
	wire.Bind(new(apiserver.AppProvider), new(*investigations.InvestigationsAppProvider)),
	ProvideAppProviders,
)
