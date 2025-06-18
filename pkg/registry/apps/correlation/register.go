package correlation

import (
	"fmt"

	"github.com/grafana/grafana-app-sdk/app"
	"github.com/grafana/grafana-app-sdk/simple"
	"github.com/grafana/grafana/apps/correlation/pkg/apis"
	correlationv0alpha1 "github.com/grafana/grafana/apps/correlation/pkg/apis/correlation/v0alpha1"
	correlationapp "github.com/grafana/grafana/apps/correlation/pkg/app"
	"github.com/grafana/grafana/pkg/apimachinery/utils"
	grafanarest "github.com/grafana/grafana/pkg/apiserver/rest"
	"github.com/grafana/grafana/pkg/services/apiserver/builder/runner"
	"github.com/grafana/grafana/pkg/services/apiserver/endpoints/request"
	correlationsvc "github.com/grafana/grafana/pkg/services/correlations"
	"github.com/grafana/grafana/pkg/setting"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type CorrelationAppProvider struct {
	app.Provider
	cfg     *setting.Cfg
	service correlationsvc.Service
}

func RegisterApp(
	cfg *setting.Cfg,
	correlationsvc correlationsvc.Service,
) *CorrelationAppProvider {
	provider := &CorrelationAppProvider{
		cfg:     cfg,
		service: correlationsvc,
	}
	appCfg := &runner.AppBuilderConfig{
		OpenAPIDefGetter:    correlationv0alpha1.GetOpenAPIDefinitions,
		ManagedKinds:        correlationapp.GetKinds(),
		LegacyStorageGetter: provider.legacyStorageGetter,
	}
	provider.Provider = simple.NewAppProvider(apis.LocalManifest(), appCfg, correlationapp.New)
	return provider
}

func (c *CorrelationAppProvider) legacyStorageGetter(requested schema.GroupVersionResource) grafanarest.Storage {
	gvr := schema.GroupVersionResource{
		Group:    correlationv0alpha1.CorrelationKind().Group(),
		Version:  correlationv0alpha1.CorrelationKind().Version(),
		Resource: correlationv0alpha1.CorrelationKind().Plural(),
	}
	if requested.String() != gvr.String() {
		return nil
	}
	legacyStore := &legacyStorage{
		service:    c.service,
		namespacer: request.GetNamespaceMapper(c.cfg),
	}
	legacyStore.tableConverter = utils.NewTableConverter(
		gvr.GroupResource(),
		utils.TableColumns{
			Definition: []metav1.TableColumnDefinition{
				{Name: "Name", Type: "string", Format: "name"},
				{Name: "Source UID", Type: "string", Format: "string"},
				{Name: "Target UID", Type: "string", Format: "string"},
				{Name: "Label", Type: "string", Format: "string"},
				{Name: "Description", Type: "string", Format: "string"},
				{Name: "Config", Type: "string", Format: "string"},
				{Name: "Provisioned", Type: "integer", Format: "int32"},
				{Name: "Type", Type: "string", Format: "string"},
			},
			Reader: func(obj any) ([]interface{}, error) {
				m, ok := obj.(*correlationv0alpha1.Correlation)
				if !ok {
					return nil, fmt.Errorf("expected correlation")
				}
				return []interface{}{
					m.Name,
					m.Spec.SourceUid,
					m.Spec.TargetUid,
					m.Spec.Label,
					m.Spec.Description,
					m.Spec.Config,
					m.Spec.Provisioned,
					m.Spec.Type,
				}, nil
			},
		},
	)
	return legacyStore
}
