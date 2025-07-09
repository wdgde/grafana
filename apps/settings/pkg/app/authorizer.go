package app

import (
	"context"

	"github.com/grafana/grafana/pkg/apimachinery/identity"
	"k8s.io/apiserver/pkg/authorization/authorizer"
	"k8s.io/klog/v2"
)

func GetAuthorizer() authorizer.Authorizer {
	return authorizer.AuthorizerFunc(func(
		ctx context.Context, attr authorizer.Attributes,
	) (authorized authorizer.Decision, reason string, err error) {
		if !attr.IsResourceRequest() {
			return authorizer.DecisionNoOpinion, "", nil
		}

		// allow service calls
		if identity.IsServiceIdentity(ctx) {
			klog.InfoS("service access", "resource", attr.GetResource())
			return authorizer.DecisionAllow, "", nil
		}

		// require a user
		u, err := identity.GetRequester(ctx)

		if err != nil {
			return authorizer.DecisionDeny, "valid user is required", err
		}

		// check if is admin
		if u.HasRole(identity.RoleAdmin) {
			klog.InfoS("admin access", "resource", attr.GetResource())
			return authorizer.DecisionAllow, "", nil
		}

		return authorizer.DecisionDeny, "forbidden", nil
	})
}
