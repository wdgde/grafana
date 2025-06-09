package contracts

import (
	"context"
	"errors"

	secretv0alpha1 "github.com/grafana/grafana/pkg/apis/secret/v0alpha1"
	"github.com/grafana/grafana/pkg/registry/apis/secret/xkube"
	"k8s.io/apiserver/pkg/authorization/authorizer"
)

var (
	ErrDecryptNotFound      = errors.New("not found")
	ErrDecryptNotAuthorized = errors.New("not authorized")
	ErrDecryptFailed        = errors.New("decryption failed")
)

// DecryptStorage is the interface for wiring and dependency injection.
type DecryptStorage interface {
	Decrypt(ctx context.Context, namespace xkube.Namespace, name string) (secretv0alpha1.ExposedSecureValue, error)
}

// DecryptAuthorizer is the interface for authorizing decryption requests.
type DecryptRequest struct {
	Namespace  xkube.Namespace
	Name       string
	Decrypters []string
}

type DecryptAuthorizer interface {
	Authorize(ctx context.Context, req DecryptRequest) (identity string, decision authorizer.Decision, err error)
}

// TEMPORARY: Needed to pass it with wire.
type DecryptAllowList map[string]struct{}
