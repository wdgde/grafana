package decrypt

import (
	"context"
	"testing"

	"github.com/grafana/authlib/authn"
	"github.com/grafana/authlib/types"
	"github.com/stretchr/testify/require"
	"k8s.io/apiserver/pkg/authorization/authorizer"

	"github.com/grafana/grafana/pkg/apimachinery/identity"
	"github.com/grafana/grafana/pkg/registry/apis/secret/contracts"
)

func TestDecryptAuthorizer(t *testing.T) {
	t.Run("when no auth info is present, it returns false", func(t *testing.T) {
		ctx := context.Background()
		authz := NewDecryptersAuthorizer(nil)

		identity, decision, err := authz.Authorize(ctx, contracts.DecryptRequest{})
		require.Empty(t, identity)
		require.Equal(t, authorizer.DecisionDeny, decision)
		require.Error(t, err)
	})

	t.Run("when token permissions are empty, it returns false", func(t *testing.T) {
		ctx := createAuthContext(context.Background(), "identity", []string{})
		authz := NewDecryptersAuthorizer(nil)

		identity, decision, err := authz.Authorize(ctx, contracts.DecryptRequest{})
		require.NotEmpty(t, identity)
		require.Equal(t, authorizer.DecisionDeny, decision)
		require.Error(t, err)
	})

	t.Run("when service identity is empty, it returns false", func(t *testing.T) {
		ctx := createAuthContext(context.Background(), "", []string{})
		authz := NewDecryptersAuthorizer(nil)

		identity, decision, err := authz.Authorize(ctx, contracts.DecryptRequest{})
		require.Empty(t, identity)
		require.Equal(t, authorizer.DecisionDeny, decision)
		require.Error(t, err)
	})

	t.Run("when permission format is malformed (missing verb), it returns false", func(t *testing.T) {
		authz := NewDecryptersAuthorizer(nil)

		// nameless
		ctx := createAuthContext(context.Background(), "identity", []string{"secret.grafana.app/securevalues"})
		identity, decision, err := authz.Authorize(ctx, contracts.DecryptRequest{})
		require.NotEmpty(t, identity)
		require.Equal(t, authorizer.DecisionDeny, decision)
		require.Error(t, err)

		// named
		ctx = createAuthContext(context.Background(), "identity", []string{"secret.grafana.app/securevalues/name"})
		identity, decision, err = authz.Authorize(ctx, contracts.DecryptRequest{})
		require.NotEmpty(t, identity)
		require.Equal(t, authorizer.DecisionDeny, decision)
		require.Error(t, err)
	})

	t.Run("when permission verb is not exactly `decrypt`, it returns false", func(t *testing.T) {
		authz := NewDecryptersAuthorizer(nil)

		// nameless
		ctx := createAuthContext(context.Background(), "identity", []string{"secret.grafana.app/securevalues:*"})
		identity, decision, err := authz.Authorize(ctx, contracts.DecryptRequest{})
		require.NotEmpty(t, identity)
		require.Equal(t, authorizer.DecisionDeny, decision)
		require.Error(t, err)

		// named
		ctx = createAuthContext(context.Background(), "identity", []string{"secret.grafana.app/securevalues/name:something"})
		identity, decision, err = authz.Authorize(ctx, contracts.DecryptRequest{})
		require.NotEmpty(t, identity)
		require.Equal(t, authorizer.DecisionDeny, decision)
		require.Error(t, err)
	})

	t.Run("when permission does not have 2 or 3 parts, it returns false", func(t *testing.T) {
		ctx := createAuthContext(context.Background(), "identity", []string{"secret.grafana.app:decrypt"})
		authz := NewDecryptersAuthorizer(nil)

		identity, decision, err := authz.Authorize(ctx, contracts.DecryptRequest{})
		require.NotEmpty(t, identity)
		require.Equal(t, authorizer.DecisionDeny, decision)
		require.Error(t, err)
	})

	t.Run("when permission has group that is not `secret.grafana.app`, it returns false", func(t *testing.T) {
		ctx := createAuthContext(context.Background(), "identity", []string{"wrong.group/securevalues/invalid:decrypt"})
		authz := NewDecryptersAuthorizer(nil)

		identity, decision, err := authz.Authorize(ctx, contracts.DecryptRequest{})
		require.NotEmpty(t, identity)
		require.Equal(t, authorizer.DecisionDeny, decision)
		require.Error(t, err)
	})

	t.Run("when permission has resource that is not `securevalues`, it returns false", func(t *testing.T) {
		authz := NewDecryptersAuthorizer(nil)

		// nameless
		ctx := createAuthContext(context.Background(), "identity", []string{"secret.grafana.app/invalid-resource:decrypt"})
		identity, decision, err := authz.Authorize(ctx, contracts.DecryptRequest{})
		require.NotEmpty(t, identity)
		require.Equal(t, authorizer.DecisionDeny, decision)
		require.Error(t, err)

		// named
		ctx = createAuthContext(context.Background(), "identity", []string{"secret.grafana.app/invalid-resource/name:decrypt"})
		identity, decision, err = authz.Authorize(ctx, contracts.DecryptRequest{})
		require.NotEmpty(t, identity)
		require.Equal(t, authorizer.DecisionDeny, decision)
		require.Error(t, err)
	})

	t.Run("when the identity is not in the allow list, it returns false", func(t *testing.T) {
		ctx := createAuthContext(context.Background(), "identity", []string{"secret.grafana.app/securevalues:decrypt"})
		authz := NewDecryptersAuthorizer(map[string]struct{}{"allowed1": {}})

		identity, decision, err := authz.Authorize(ctx, contracts.DecryptRequest{})
		require.NotEmpty(t, identity)
		require.Equal(t, authorizer.DecisionDeny, decision)
		require.Error(t, err)
	})

	t.Run("when the identity doesn't match any allowed decrypters, it returns false", func(t *testing.T) {
		authz := NewDecryptersAuthorizer(map[string]struct{}{"identity": {}})

		// nameless
		ctx := createAuthContext(context.Background(), "identity", []string{"secret.grafana.app/securevalues:decrypt"})
		identity, decision, err := authz.Authorize(ctx, contracts.DecryptRequest{Decrypters: []string{"group2"}})
		require.NotEmpty(t, identity)
		require.Equal(t, authorizer.DecisionDeny, decision)
		require.Error(t, err)

		// named
		ctx = createAuthContext(context.Background(), "identity", []string{"secret.grafana.app/securevalues/name:decrypt"})
		identity, decision, err = authz.Authorize(ctx, contracts.DecryptRequest{Decrypters: []string{"group2"}})
		require.NotEmpty(t, identity)
		require.Equal(t, authorizer.DecisionDeny, decision)
		require.Error(t, err)
	})

	t.Run("when the identity matches an allowed decrypter, it returns true", func(t *testing.T) {
		authz := NewDecryptersAuthorizer(map[string]struct{}{"identity": {}})

		// nameless
		ctx := createAuthContext(context.Background(), "identity", []string{"secret.grafana.app/securevalues:decrypt"})
		identity, decision, err := authz.Authorize(ctx, contracts.DecryptRequest{Decrypters: []string{"identity"}})
		require.Equal(t, "identity", identity)
		require.Equal(t, authorizer.DecisionAllow, decision)
		require.NoError(t, err)

		// named
		ctx = createAuthContext(context.Background(), "identity", []string{"secret.grafana.app/securevalues/name:decrypt"})
		identity, decision, err = authz.Authorize(ctx, contracts.DecryptRequest{Name: "name", Decrypters: []string{"identity"}})
		require.Equal(t, "identity", identity)
		require.Equal(t, authorizer.DecisionAllow, decision)
		require.NoError(t, err)
	})

	t.Run("when there are multiple permissions, some invalid, only valid ones are considered", func(t *testing.T) {
		ctx := createAuthContext(context.Background(), "identity", []string{
			"secret.grafana.app/securevalues/name1:decrypt",
			"secret.grafana.app/securevalues/name2:decrypt",
			"secret.grafana.app/securevalues/invalid:read",
			"wrong.group/securevalues/group2:decrypt",
			"secret.grafana.app/securevalues/identity:decrypt", // old style of identity+permission
		})
		authz := NewDecryptersAuthorizer(map[string]struct{}{"identity": {}})

		identity, decision, err := authz.Authorize(ctx, contracts.DecryptRequest{Name: "name1", Decrypters: []string{"identity"}})
		require.Equal(t, "identity", identity)
		require.Equal(t, authorizer.DecisionAllow, decision)
		require.NoError(t, err)

		identity, decision, err = authz.Authorize(ctx, contracts.DecryptRequest{Name: "name2", Decrypters: []string{"identity"}})
		require.Equal(t, "identity", identity)
		require.Equal(t, authorizer.DecisionAllow, decision)
		require.NoError(t, err)
	})

	t.Run("when empty secure value name with specific permission, it returns false", func(t *testing.T) {
		ctx := createAuthContext(context.Background(), "identity", []string{"secret.grafana.app/securevalues/name:decrypt"})
		authz := NewDecryptersAuthorizer(map[string]struct{}{"identity": {}})

		identity, decision, err := authz.Authorize(ctx, contracts.DecryptRequest{Decrypters: []string{"identity"}})
		require.Equal(t, "identity", identity)
		require.Equal(t, authorizer.DecisionDeny, decision)
		require.Error(t, err)
	})

	t.Run("when permission has an extra / but no name, it returns false", func(t *testing.T) {
		ctx := createAuthContext(context.Background(), "identity", []string{"secret.grafana.app/securevalues/:decrypt"})
		authz := NewDecryptersAuthorizer(map[string]struct{}{"identity": {}})

		identity, decision, err := authz.Authorize(ctx, contracts.DecryptRequest{Decrypters: []string{"identity"}})
		require.Equal(t, "identity", identity)
		require.Equal(t, authorizer.DecisionDeny, decision)
		require.Error(t, err)
	})

	t.Run("when the decrypters list is empty, meaning nothing can decrypt the secure value, it returns false", func(t *testing.T) {
		ctx := createAuthContext(context.Background(), "identity", []string{"secret.grafana.app/securevalues:decrypt"})
		authz := NewDecryptersAuthorizer(map[string]struct{}{"identity": {}})

		identity, decision, err := authz.Authorize(ctx, contracts.DecryptRequest{Name: "name"})
		require.Equal(t, "identity", identity)
		require.Equal(t, authorizer.DecisionDeny, decision)
		require.Error(t, err)
	})

	t.Run("when one of decrypters matches the identity, it returns true", func(t *testing.T) {
		ctx := createAuthContext(context.Background(), "identity1", []string{"secret.grafana.app/securevalues:decrypt"})
		authz := NewDecryptersAuthorizer(map[string]struct{}{"identity1": {}, "identity2": {}})

		identity, decision, err := authz.Authorize(ctx, contracts.DecryptRequest{Decrypters: []string{"identity1", "identity2", "identity3"}})
		require.Equal(t, "identity1", identity)
		require.Equal(t, authorizer.DecisionAllow, decision)
		require.NoError(t, err)
	})

	t.Run("permissions must be case-sensitive and return false", func(t *testing.T) {
		authz := NewDecryptersAuthorizer(map[string]struct{}{"identity": {}})

		ctx := createAuthContext(context.Background(), "identity", []string{"SECRET.grafana.app/securevalues:decrypt"})
		identity, decision, err := authz.Authorize(ctx, contracts.DecryptRequest{Decrypters: []string{"identity"}})
		require.Equal(t, "identity", identity)
		require.Equal(t, authorizer.DecisionDeny, decision)
		require.Error(t, err)

		ctx = createAuthContext(context.Background(), "identity", []string{"secret.grafana.app/SECUREVALUES:decrypt"})
		identity, decision, err = authz.Authorize(ctx, contracts.DecryptRequest{Decrypters: []string{"identity"}})
		require.Equal(t, "identity", identity)
		require.Equal(t, authorizer.DecisionDeny, decision)
		require.Error(t, err)

		ctx = createAuthContext(context.Background(), "identity", []string{"secret.grafana.app/securevalues:DECRYPT"})
		identity, decision, err = authz.Authorize(ctx, contracts.DecryptRequest{Decrypters: []string{"identity"}})
		require.Equal(t, "identity", identity)
		require.Equal(t, authorizer.DecisionDeny, decision)
		require.Error(t, err)
	})
}

func createAuthContext(ctx context.Context, serviceIdentity string, permissions []string) context.Context {
	requester := &identity.StaticRequester{
		AccessTokenClaims: &authn.Claims[authn.AccessTokenClaims]{
			Rest: authn.AccessTokenClaims{
				Permissions:     permissions,
				ServiceIdentity: serviceIdentity,
			},
		},
	}

	return types.WithAuthInfo(ctx, requester)
}

// Adapted from https://github.com/grafana/authlib/blob/1492b99410603ca15730a1805a9220ce48232bc3/authz/client_test.go#L18
func TestHasPermissionInToken(t *testing.T) {
	t.Parallel()

	tests := []struct {
		test             string
		tokenPermissions []string
		name             string
		want             bool
	}{
		{
			test:             "Permission matches group/resource",
			tokenPermissions: []string{"secret.grafana.app/securevalues:decrypt"},
			want:             true,
		},
		{
			test:             "Permission does not match verb",
			tokenPermissions: []string{"secret.grafana.app/securevalues:create"},
			want:             false,
		},
		{
			test:             "Permission does not have support for wildcard verb",
			tokenPermissions: []string{"secret.grafana.app/securevalues:*"},
			want:             false,
		},
		{
			test:             "Invalid permission missing verb",
			tokenPermissions: []string{"secret.grafana.app/securevalues"},
			want:             false,
		},
		{
			test:             "Permission on the wrong group",
			tokenPermissions: []string{"other-group.grafana.app/securevalues:decrypt"},
			want:             false,
		},
		{
			test:             "Permission on the wrong resource",
			tokenPermissions: []string{"secret.grafana.app/other-resource:decrypt"},
			want:             false,
		},
		{
			test:             "Permission without group are skipped",
			tokenPermissions: []string{":decrypt"},
			want:             false,
		},
		{
			test:             "Group level permission is not supported",
			tokenPermissions: []string{"secret.grafana.app:decrypt"},
			want:             false,
		},
		{
			test:             "Permission with name matches group/resource/name",
			tokenPermissions: []string{"secret.grafana.app/securevalues/name:decrypt"},
			name:             "name",
			want:             true,
		},
		{
			test:             "Permission with name2 does not matche group/resource/name1",
			tokenPermissions: []string{"secret.grafana.app/securevalues/name1:decrypt"},
			name:             "name2",
			want:             false,
		},
		{
			test:             "Parts need an exact match",
			tokenPermissions: []string{"secret.grafana.app/secure:*"},
			want:             false,
		},
		{
			test:             "Resource specific permission should not allow access to all resources",
			tokenPermissions: []string{"secret.grafana.app/securevalues/name:decrypt"},
			name:             "",
			want:             false,
		},
		{
			test:             "Permission at group/resource should allow access to all resources",
			tokenPermissions: []string{"secret.grafana.app/securevalues:decrypt"},
			name:             "name",
			want:             true,
		},
		{
			test:             "Empty name trying to match everything is not allowed",
			tokenPermissions: []string{"secret.grafana.app/securevalues/:decrypt"},
			name:             "",
			want:             false,
		},
		{
			test:             "Empty name trying to match a specific name is not allowed",
			tokenPermissions: []string{"secret.grafana.app/securevalues/:decrypt"},
			name:             "name",
			want:             false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			t.Parallel()

			got := hasPermissionInToken(tt.tokenPermissions, tt.name)
			require.Equal(t, tt.want, got)
		})
	}
}
