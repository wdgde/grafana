package upgrades

import (
	"fmt"

	upgradesv0alpha1 "github.com/grafana/grafana/apps/upgrades/pkg/apis/upgrades/v0alpha1"
	grafanaregistry "github.com/grafana/grafana/pkg/apiserver/registry/generic"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/authorization/authorizer"
	"k8s.io/apiserver/pkg/registry/rest"
	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/kube-openapi/pkg/common"

	"github.com/grafana/grafana/pkg/services/apiserver/builder"
	"github.com/grafana/grafana/pkg/services/apiserver/builder/runner"
)

var (
	_ builder.APIGroupVersionProvider = (*UpgradesAPIBuilder)(nil)
	_ builder.APIGroupAuthorizer      = (*UpgradesAPIBuilder)(nil)
	_ builder.APIGroupBuilder         = (*UpgradesAPIBuilder)(nil)
)

type UpgradesAPIBuilder struct{}

func NewUpgradesAPIBuilder() (*UpgradesAPIBuilder, error) {
	return &UpgradesAPIBuilder{}, nil
}

func RegisterAPIService(apiregistration builder.APIRegistrar) (*UpgradesAPIBuilder, error) {
	builder, err := NewUpgradesAPIBuilder()
	if err != nil {
		panic(fmt.Errorf("failed to create UpgradesAPIBuilder: %v", err))
	}

	apiregistration.RegisterAPI(builder)

	return builder, err
}

func (b *UpgradesAPIBuilder) GetGroupVersion() schema.GroupVersion {
	return upgradesv0alpha1.GroupVersion
}

func (b *UpgradesAPIBuilder) InstallSchema(scheme *runtime.Scheme) error {
	gv := schema.GroupVersion{Group: upgradesv0alpha1.APIGroup, Version: upgradesv0alpha1.APIVersion}
	types := []runtime.Object{
		&upgradesv0alpha1.UpgradeMetadata{},
		&upgradesv0alpha1.UpgradeMetadataList{},
	}

	scheme.AddKnownTypes(gv, types...)
	scheme.AddKnownTypes(schema.GroupVersion{Group: upgradesv0alpha1.APIGroup, Version: runtime.APIVersionInternal}, types...)

	metav1.AddToGroupVersion(scheme, gv)

	if err := scheme.SetVersionPriority(gv); err != nil {
		return fmt.Errorf("scheme set version priority: %w", err)
	}

	return nil
}

func (b *UpgradesAPIBuilder) AllowedV0Alpha1Resources() []string {
	return nil
}

func (b *UpgradesAPIBuilder) UpdateAPIGroupInfo(apiGroupInfo *genericapiserver.APIGroupInfo, opts builder.APIGroupOptions) error {
	resourceInfo := runner.KindToResourceInfo(upgradesv0alpha1.UpgradeMetadataKind())

	unifiedStore, err := grafanaregistry.NewRegistryStore(opts.Scheme, resourceInfo, opts.OptsGetter)
	if err != nil {
		return fmt.Errorf("failed to create repository storage: %w", err)
	}

	storage := map[string]rest.Storage{}
	storage[resourceInfo.StoragePath()] = unifiedStore
	storage["check_for_upgrades"] = &checkForUpgradesREST{}

	apiGroupInfo.VersionedResourcesStorageMap[upgradesv0alpha1.APIVersion] = storage

	return nil
}

func (b *UpgradesAPIBuilder) GetOpenAPIDefinitions() common.GetOpenAPIDefinitions {
	return upgradesv0alpha1.GetOpenAPIDefinitions
}

func (b *UpgradesAPIBuilder) GetAuthorizer() authorizer.Authorizer {
	return nil
}
