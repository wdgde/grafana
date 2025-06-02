package correlation

import (
	"context"
	"errors"

	"github.com/prometheus/client_golang/prometheus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/admission"
	"k8s.io/apiserver/pkg/authorization/authorizer"
	"k8s.io/apiserver/pkg/registry/rest"
	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/kube-openapi/pkg/common"
	"k8s.io/kube-openapi/pkg/spec3"

	authtypes "github.com/grafana/authlib/types"

	correlation "github.com/grafana/grafana/apps/correlation/pkg/apis/correlation/v0alpha1"
	grafanaregistry "github.com/grafana/grafana/pkg/apiserver/registry/generic"
	grafanarest "github.com/grafana/grafana/pkg/apiserver/rest"
	"github.com/grafana/grafana/pkg/services/apiserver/builder"
	"github.com/grafana/grafana/pkg/services/apiserver/endpoints/request"
	correlationService "github.com/grafana/grafana/pkg/services/correlations"
	"github.com/grafana/grafana/pkg/services/featuremgmt"
	"github.com/grafana/grafana/pkg/setting"
	"github.com/grafana/grafana/pkg/storage/unified/apistore"
	"github.com/grafana/grafana/pkg/storage/unified/resource"
)

var _ builder.APIGroupBuilder = (*CorrelationAPIBuilder)(nil)
var _ builder.APIGroupValidation = (*CorrelationAPIBuilder)(nil)

var resourceInfo = correlation.CorrelationResourceInfo

var errNoUser = errors.New("valid user is required")
var errNoResource = errors.New("resource name is required")

// This is used just so wire has something unique to return
type CorrelationAPIBuilder struct {
	gv                 schema.GroupVersion
	features           featuremgmt.FeatureToggles
	namespacer         request.NamespaceMapper
	storage            grafanarest.Storage
	authorizer         authorizer.Authorizer
	correlationService *correlationService.CorrelationsService
	cfg                *setting.Cfg
}

func RegisterAPIService(cfg *setting.Cfg,
	features featuremgmt.FeatureToggles,
	apiregistration builder.APIRegistrar,
	registerer prometheus.Registerer,
	correlationService *correlationService.CorrelationsService,
	unified resource.ResourceClient,
) *CorrelationAPIBuilder {
	builder := &CorrelationAPIBuilder{
		gv:                 resourceInfo.GroupVersion(),
		features:           features,
		namespacer:         request.GetNamespaceMapper(cfg),
		cfg:                cfg,
		correlationService: correlationService,
	}
	apiregistration.RegisterAPI(builder)
	return builder
}

func NewAPIService(ac authtypes.AccessClient) *CorrelationAPIBuilder {
	return &CorrelationAPIBuilder{
		gv:         resourceInfo.GroupVersion(),
		namespacer: request.GetNamespaceMapper(nil),
	}
}

func (b *CorrelationAPIBuilder) GetGroupVersion() schema.GroupVersion {
	return b.gv
}

func addKnownTypes(scheme *runtime.Scheme, gv schema.GroupVersion) {
	scheme.AddKnownTypes(gv,
		&correlation.Correlation{},
		&correlation.CorrelationList{},
	)
}

func (b *CorrelationAPIBuilder) InstallSchema(scheme *runtime.Scheme) error {
	addKnownTypes(scheme, b.gv)

	// Link this version to the internal representation.
	// This is used for server-side-apply (PATCH), and avoids the error:
	//   "no kind is registered for the type"
	addKnownTypes(scheme, schema.GroupVersion{
		Group:   b.gv.Group,
		Version: runtime.APIVersionInternal,
	})

	// If multiple versions exist, then register conversions from zz_generated.conversion.go
	// if err := playlist.RegisterConversions(scheme); err != nil {
	//   return err
	// }
	metav1.AddToGroupVersion(scheme, b.gv)
	return scheme.SetVersionPriority(b.gv)
}

func (b *CorrelationAPIBuilder) UpdateAPIGroupInfo(apiGroupInfo *genericapiserver.APIGroupInfo, opts builder.APIGroupOptions) error {
	scheme := opts.Scheme
	optsGetter := opts.OptsGetter
	storage := map[string]rest.Storage{}

	opts.StorageOptsRegister(resourceInfo.GroupResource(), apistore.StorageOptions{
		EnableFolderSupport:         false,
		RequireDeprecatedInternalID: true,
	})

	store, err := grafanaregistry.NewRegistryStore(scheme, resourceInfo, optsGetter)
	if err != nil {
		return err
	}

	legacyStore, err := NewStore(b.correlationService, resourceInfo, scheme, optsGetter)
	if err != nil {
		return err
	}
	storage[resourceInfo.StoragePath()], err = opts.DualWriteBuilder(resourceInfo.GroupResource(), legacyStore, store)
	if err != nil {
		return err
	}
	apiGroupInfo.VersionedResourcesStorageMap[correlation.VERSION] = storage
	b.storage = storage[resourceInfo.StoragePath()].(grafanarest.Storage)
	return nil
}

func (b *CorrelationAPIBuilder) GetOpenAPIDefinitions() common.GetOpenAPIDefinitions {
	return correlation.GetOpenAPIDefinitions
}

func (b *CorrelationAPIBuilder) PostProcessOpenAPI(oas *spec3.OpenAPI) (*spec3.OpenAPI, error) {
	oas.Info.Description = "Grafana correlation"
	return oas, nil
}

func (b *CorrelationAPIBuilder) GetAuthorizer() authorizer.Authorizer {
	return b.authorizer
}

func (b *CorrelationAPIBuilder) Mutate(ctx context.Context, a admission.Attributes, _ admission.ObjectInterfaces) error {
	return nil
}

func (b *CorrelationAPIBuilder) Validate(ctx context.Context, a admission.Attributes, _ admission.ObjectInterfaces) error {

	return nil
}
