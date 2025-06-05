package correlation

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/grafana/authlib/claims"
	correlation "github.com/grafana/grafana/apps/correlation/pkg/apis/correlation/v0alpha1"
	correlationv0alpha1 "github.com/grafana/grafana/apps/correlation/pkg/apis/correlation/v0alpha1"
	"github.com/grafana/grafana/pkg/apimachinery/utils"
	grafanaregistry "github.com/grafana/grafana/pkg/apiserver/registry/generic"
	grafanarest "github.com/grafana/grafana/pkg/apiserver/rest"
	"github.com/grafana/grafana/pkg/registry/apis/dashboard/legacy"
	correlationService "github.com/grafana/grafana/pkg/services/correlations"
	"github.com/grafana/grafana/pkg/storage/unified/apistore"
	"github.com/grafana/grafana/pkg/storage/unified/resource"
	"github.com/grafana/grafana/pkg/storage/unified/resourcepb"
	"github.com/prometheus/client_golang/prometheus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/registry/generic"
)

type correlationBackend struct {
	cs *correlationService.CorrelationsService
}

func (b *correlationBackend) WriteEvent(context.Context, resource.WriteEvent) (int64, error) {
	return 0, fmt.Errorf("not implemented")
}

func (b *correlationBackend) ReadResource(context.Context, *resourcepb.ReadRequest) *resource.BackendReadResponse {
	return nil
}

type it struct {
	correlations []correlationService.Correlation
	index        int
}

func (i *it) Next() bool {
	i.index++
	return i.index < len(i.correlations)
}

func (i *it) Error() error {
	return nil
}
func (i *it) ContinueToken() string {
	return ""
}

func (i *it) Folder() string {
	return ""
}

func (i *it) Name() string {
	return ""
}
func (i *it) Namespace() string {
	return ""
}

func (i *it) ResourceVersion() int64 {
	return 1
}

// Value implements resource.ListIterator.
func (i *it) Value() []byte {
	data := i.correlations[i.index]

	c := &correlation.Correlation{
		ObjectMeta: metav1.ObjectMeta{
			Name: data.UID,
		},
		Spec: correlation.CorrelationSpec{
			SourceUid: data.SourceUID,
		},
	}

	if data.TargetUID != nil {
		c.Spec.SourceUid = *data.TargetUID
	}

	b, err := json.Marshal(c)
	if err != nil {
		return nil
	}
	return b
}

func (b *correlationBackend) ListIterator(ctx context.Context, req *resourcepb.ListRequest, cb func(resource.ListIterator) error) (int64, error) {
	info, err := claims.ParseNamespace(req.Options.Key.Namespace)
	if err != nil {
		return 0, err
	}
	c, err := b.cs.GetCorrelations(ctx, correlationService.GetCorrelationsQuery{
		OrgId: info.OrgID,
		Limit: 10000,
	})
	if err != nil {
		return 0, err
	}
	return 1, cb(&it{
		correlations: c.Correlations,
		index:        -1,
	})
}

// Get all events from the store
// For HA setups, this will be more events than the local WriteEvent above!
func (b *correlationBackend) WatchWriteEvents(ctx context.Context) (<-chan *resource.WrittenEvent, error) {
	ch := make(chan *resource.WrittenEvent)
	return ch, nil
}

// ListHistory is like ListIterator, but it returns the history of a resource
func (b *correlationBackend) ListHistory(context.Context, *resourcepb.ListRequest, func(resource.ListIterator) error) (int64, error) {
	return 0, fmt.Errorf("not implemented")
}

// Get resource stats within the storage backend.  When namespace is empty, it will apply to all
func (b *correlationBackend) GetResourceStats(ctx context.Context, namespace string, minCount int) ([]resource.ResourceStats, error) {
	return []resource.ResourceStats{}, nil
}

func NewStore(cs *correlationService.CorrelationsService, scheme *runtime.Scheme, defaultOptsGetter generic.RESTOptionsGetter) (grafanarest.Storage, error) {
	server, err := resource.NewResourceServer(resource.ResourceServerOptions{
		Backend: &correlationBackend{
			cs: cs,
		},
		Reg: prometheus.DefaultRegisterer,
	})
	if err != nil {
		return nil, err
	}
	gr := schema.GroupResource{
		Group:    correlationv0alpha1.CorrelationKind().Group(),
		Resource: correlationv0alpha1.CorrelationKind().Plural(),
	}

	defaultOpts, err := defaultOptsGetter.GetRESTOptions(gr, nil)
	if err != nil {
		return nil, err
	}
	client := legacy.NewDirectResourceClient(server) // same context
	optsGetter := apistore.NewRESTOptionsGetterForClient(client,
		defaultOpts.StorageConfig.Config, nil,
	)

	optsGetter.RegisterOptions(gr, apistore.StorageOptions{
		EnableFolderSupport:         true,
		RequireDeprecatedInternalID: true,
	})

	ri := utils.ResourceInfo{}

	return grafanaregistry.NewRegistryStore(scheme, ri, optsGetter)
}
