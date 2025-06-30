package upgrades

import (
	"context"
	"net/http"

	upgradesv0alpha1 "github.com/grafana/grafana/apps/upgrades/pkg/apis/upgrades/v0alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/rest"
)

type checkForUpgradesREST struct {
}

var (
	_ rest.Storage              = (*checkForUpgradesREST)(nil)
	_ rest.SingularNameProvider = (*checkForUpgradesREST)(nil)
	_ rest.Connecter            = (*checkForUpgradesREST)(nil)
	_ rest.Scoper               = (*checkForUpgradesREST)(nil)
	_ rest.StorageMetadata      = (*checkForUpgradesREST)(nil)
)

func (r *checkForUpgradesREST) New() runtime.Object {
	return &upgradesv0alpha1.UpgradeMetadata{}
}

func (r *checkForUpgradesREST) Destroy() {}

func (r *checkForUpgradesREST) NamespaceScoped() bool {
	return true
}

func (r *checkForUpgradesREST) GetSingularName() string {
	return "upgrademetadata"
}

func (r *checkForUpgradesREST) ProducesMIMETypes(verb string) []string {
	return []string{"application/json"}
}

func (r *checkForUpgradesREST) ProducesObject(verb string) interface{} {
	return &upgradesv0alpha1.UpgradeMetadata{}
}

func (r *checkForUpgradesREST) ConnectMethods() []string {
	return []string{"GET"}
}

func (r *checkForUpgradesREST) NewConnectOptions() (runtime.Object, bool, string) {
	return nil, false, "" // true means you can use the trailing path as a variable
}

func (r *checkForUpgradesREST) Connect(ctx context.Context, name string, opts runtime.Object, responder rest.Responder) (http.Handler, error) {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		responder.Object(200, nil)
	}), nil
}
