package decrypt

import (
	"context"
	"errors"
	"strings"

	"github.com/grafana/authlib/authn"
	claims "github.com/grafana/authlib/types"
	secretv0alpha1 "github.com/grafana/grafana/pkg/apis/secret/v0alpha1"
	"github.com/grafana/grafana/pkg/registry/apis/secret/contracts"
	"k8s.io/apiserver/pkg/authorization/authorizer"
)

type decryptersAuthorizer struct {
	allowList contracts.DecryptAllowList
}

func NewDecryptersAuthorizer(allowList contracts.DecryptAllowList) contracts.DecryptAuthorizer {
	return &decryptersAuthorizer{
		allowList: allowList,
	}
}

func (a *decryptersAuthorizer) Authorize(ctx context.Context, req contracts.DecryptRequest) (identity string, allowed authorizer.Decision, err error) {
	authInfo, ok := claims.AuthInfoFrom(ctx)
	if !ok {
		return "", authorizer.DecisionDeny, errors.New("no auth info found")
	}

	serviceIdentityList, ok := authInfo.GetExtra()[authn.ServiceIdentityKey]
	if !ok {
		return "", authorizer.DecisionDeny, errors.New("service identity not found in auth info extra claims")
	}

	// If there's more than one service identity, something is suspicious and we reject it.
	if len(serviceIdentityList) != 1 {
		return "", authorizer.DecisionDeny, errors.New("multiple service identities found in auth info extra claims")
	}

	serviceIdentity := serviceIdentityList[0]

	// TEMPORARY: while we can't onboard every app into secrets, we can block them from decrypting
	// securevalues preemptively here before even reaching out to the database.
	// This check can be removed once we open the gates for any service to use secrets.
	if _, exists := a.allowList[serviceIdentity]; !exists || serviceIdentity == "" {
		return serviceIdentity, authorizer.DecisionDeny, errors.New("service identity not allowed to decrypt secure values")
	}

	// Checks whether the token has the permission to decrypt secure values.
	if !hasPermissionInToken(authInfo.GetTokenPermissions(), req.Name) {
		return serviceIdentity, authorizer.DecisionDeny, errors.New("service identity does not have permission to decrypt secure value")
	}

	// Finally check whether the service identity is allowed to decrypt this secure value.
	for _, decrypter := range req.Decrypters {
		if decrypter == serviceIdentity {
			return serviceIdentity, authorizer.DecisionAllow, nil
		}
	}

	return serviceIdentity, authorizer.DecisionDeny, errors.New("service identity does not have permission to decrypt secure value")
}

// Adapted from https://github.com/grafana/authlib/blob/1492b99410603ca15730a1805a9220ce48232bc3/authz/client.go#L138
// Changes: 1) we don't support `*` for verbs; 2) we support specific names in the permission.
func hasPermissionInToken(tokenPermissions []string, name string) bool {
	var (
		group    = secretv0alpha1.GROUP
		resource = secretv0alpha1.SecureValuesResourceInfo.GetName()
		verb     = "decrypt"
	)

	for _, p := range tokenPermissions {
		tokenGR, tokenVerb, found := strings.Cut(p, ":")
		if !found || tokenVerb != verb {
			continue
		}

		parts := strings.SplitN(tokenGR, "/", 3)

		switch len(parts) {
		// secret.grafana.app/securevalues:decrypt
		case 2:
			if parts[0] == group && parts[1] == resource {
				return true
			}

		// secret.grafana.app/securevalues/<name>:decrypt
		case 3:
			if parts[0] == group && parts[1] == resource && parts[2] == name && name != "" {
				return true
			}
		}
	}

	return false
}
