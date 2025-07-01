package manager

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"

	"github.com/grafana/grafana/pkg/infra/tracing"
	"github.com/grafana/grafana/pkg/infra/usagestats"
	"github.com/grafana/grafana/pkg/registry/apis/secret/encryption"
	"github.com/grafana/grafana/pkg/services/featuremgmt"
	"github.com/grafana/grafana/pkg/services/sqlstore"
	"github.com/grafana/grafana/pkg/setting"
	"github.com/grafana/grafana/pkg/storage/secret/database"
	encryptionstorage "github.com/grafana/grafana/pkg/storage/secret/encryption"
	"github.com/grafana/grafana/pkg/storage/secret/migrator"
)

func setupTestService(tb testing.TB) *EncryptionManager {
	tb.Helper()

	testDB := sqlstore.NewTestStore(tb, sqlstore.WithMigrator(migrator.New()))
	database := database.ProvideDatabase(testDB)

	features := featuremgmt.WithFeatures(featuremgmt.FlagGrafanaAPIServerWithExperimentalAPIs, featuremgmt.FlagSecretsManagementAppPlatform)
	defaultKey := "SdlklWklckeLS"
	cfg := &setting.Cfg{
		SecretsManagement: setting.SecretsManagerSettings{
			SecretKey:          defaultKey,
			EncryptionProvider: "secretKey.v1",
		},
	}
	tracer := noop.NewTracerProvider().Tracer("test")
	store, err := encryptionstorage.ProvideDataKeyStorage(database, tracer, features)
	require.NoError(tb, err)

	usageStats := &usagestats.UsageStatsMock{T: tb}

	encMgr, err := ProvideEncryptionManager(
		tracing.InitializeTracerForTest(),
		store,
		cfg,
		usageStats,
		encryption.ProvideThirdPartyProviderMap(),
	)
	require.NoError(tb, err)

	return encMgr.(*EncryptionManager)
}
