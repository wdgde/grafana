package app

import (
	"context"

	"github.com/grafana/grafana/pkg/apimachinery/identity"
	"github.com/grafana/grafana/pkg/setting"
	"k8s.io/apiserver/pkg/authorization/authorizer"
	"k8s.io/klog/v2"
)

func GetAuthorizer(cfg *setting.Cfg) authorizer.Authorizer {
	return authorizer.AuthorizerFunc(func(
		ctx context.Context, attr authorizer.Attributes,
	) (authorized authorizer.Decision, reason string, err error) {
		if !attr.IsResourceRequest() {
			return authorizer.DecisionNoOpinion, "", nil
		}
		logger := klog.FromContext(ctx).WithValues("app", "settings", "component", "authorizer")

		// allow service calls
		if identity.IsServiceIdentity(ctx) {
			logger.Info("service access", "resource", attr.GetResource())
			return authorizer.DecisionAllow, "", nil
		}

		if !cfg.SettingsAllowAdminAccess {
			return authorizer.DecisionDeny, "forbidden", nil
		}

		// require a user
		u, err := identity.GetRequester(ctx)

		if err != nil {
			return authorizer.DecisionDeny, "valid user is required", err
		}
		logger.Info("user request", "namespace", u.GetNamespace())

		// check if is admin
		if u.HasRole(identity.RoleAdmin) {
			logger.Info("admin access", "resource", attr.GetResource())
			return authorizer.DecisionAllow, "", nil
		}

		return authorizer.DecisionDeny, "forbidden", nil
	})
}
