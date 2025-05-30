package user

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/internalversion"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apiserver/pkg/registry/rest"

	claims "github.com/grafana/authlib/types"
	iamv0alpha "github.com/grafana/grafana/apps/iam/pkg/apis/iam/v0alpha1"
	"github.com/grafana/grafana/pkg/apimachinery/utils"
	iamv0 "github.com/grafana/grafana/pkg/apis/iam/v0alpha1"
	"github.com/grafana/grafana/pkg/registry/apis/iam/common"
	"github.com/grafana/grafana/pkg/registry/apis/iam/legacy"
	"github.com/grafana/grafana/pkg/services/apiserver/endpoints/request"
	"github.com/grafana/grafana/pkg/services/user"
)

var (
	_ rest.Scoper               = (*LegacyStore)(nil)
	_ rest.SingularNameProvider = (*LegacyStore)(nil)
	_ rest.Getter               = (*LegacyStore)(nil)
	_ rest.Lister               = (*LegacyStore)(nil)
	_ rest.Storage              = (*LegacyStore)(nil)
	_ rest.CreaterUpdater       = (*LegacyStore)(nil)
	_ rest.GracefulDeleter      = (*LegacyStore)(nil)
	_ rest.CollectionDeleter    = (*LegacyStore)(nil)
	_ rest.TableConvertor       = (*LegacyStore)(nil)
)

var resource = iamv0.UserResourceInfo

func NewLegacyStore(store legacy.LegacyIdentityStore, ac claims.AccessClient, userSvc user.Service) *LegacyStore {
	return &LegacyStore{store, ac, userSvc}
}

type LegacyStore struct {
	store   legacy.LegacyIdentityStore
	ac      claims.AccessClient
	userSvc user.Service
}

// DeleteCollection implements rest.CollectionDeleter.
func (s *LegacyStore) DeleteCollection(ctx context.Context, deleteValidation rest.ValidateObjectFunc, options *metav1.DeleteOptions, listOptions *internalversion.ListOptions) (runtime.Object, error) {
	panic("unimplemented")
}

func (s *LegacyStore) New() runtime.Object {
	return resource.NewFunc()
}

func (s *LegacyStore) Destroy() {}

func (s *LegacyStore) NamespaceScoped() bool {
	return true // namespace == org
}

func (s *LegacyStore) GetSingularName() string {
	return resource.GetSingularName()
}

func (s *LegacyStore) NewList() runtime.Object {
	return resource.NewListFunc()
}

func (s *LegacyStore) ConvertToTable(ctx context.Context, object runtime.Object, tableOptions runtime.Object) (*metav1.Table, error) {
	return resource.TableConverter().ConvertToTable(ctx, object, tableOptions)
}

func (s *LegacyStore) List(ctx context.Context, options *internalversion.ListOptions) (runtime.Object, error) {
	res, err := common.List(
		ctx, resource.GetName(), s.ac, common.PaginationFromListOptions(options),
		func(ctx context.Context, ns claims.NamespaceInfo, p common.Pagination) (*common.ListResponse[iamv0alpha.User], error) {
			found, err := s.store.ListUsers(ctx, ns, legacy.ListUserQuery{
				Pagination: p,
			})

			if err != nil {
				return nil, err
			}

			users := make([]iamv0alpha.User, 0, len(found.Users))
			for _, u := range found.Users {
				users = append(users, toUserItem(&u, ns.Value))
			}

			return &common.ListResponse[iamv0alpha.User]{
				Items:    users,
				RV:       found.RV,
				Continue: found.Continue,
			}, nil
		},
	)

	if err != nil {
		return nil, err
	}

	obj := &iamv0alpha.UserList{Items: res.Items}
	obj.Continue = common.OptionalFormatInt(res.Continue)
	obj.ResourceVersion = common.OptionalFormatInt(res.RV)
	return obj, nil
}

// Create implements rest.CreaterUpdater.
func (s *LegacyStore) Create(ctx context.Context, obj runtime.Object, createValidation rest.ValidateObjectFunc, options *metav1.CreateOptions) (runtime.Object, error) {
	info, err := request.NamespaceInfoFrom(ctx, true)
	if err != nil {
		return nil, err
	}
	if createValidation != nil {
		if err := createValidation(ctx, obj.DeepCopyObject()); err != nil {
			return nil, err
		}
	}
	p, ok := obj.(*iamv0alpha.User)
	if !ok {
		return nil, fmt.Errorf("expected user but got %s", obj.GetObjectKind().GroupVersionKind())
	}

	usr, err := s.userSvc.Create(ctx, &user.CreateUserCommand{
		Email:         p.Spec.Email,
		EmailVerified: p.Spec.EmailVerified,
		Name:          p.Spec.Name,
		Login:         p.Spec.Login,
		OrgID:         info.OrgID,
		IsProvisioned: false,
		IsDisabled:    p.Spec.Disabled,
	})

	spec := iamv0alpha.UserSpec{
		Disabled:      usr.IsDisabled,
		Email:         usr.Email,
		EmailVerified: usr.EmailVerified,
		Login:         usr.Login,
		Name:          usr.Name,
		Provisioned:   usr.IsProvisioned,
	}

	result := &iamv0alpha.User{
		ObjectMeta: metav1.ObjectMeta{
			UID:             types.UID(usr.UID),
			Name:            usr.UID,
			Namespace:       info.Value,
			ResourceVersion: fmt.Sprintf("%d", usr.Version),
		},
		Spec: spec,
	}
	return result, nil
}

// Update implements rest.CreaterUpdater.
func (s *LegacyStore) Update(ctx context.Context, name string, objInfo rest.UpdatedObjectInfo, createValidation rest.ValidateObjectFunc, updateValidation rest.ValidateObjectUpdateFunc, forceAllowCreate bool, options *metav1.UpdateOptions) (runtime.Object, bool, error) {
	panic("unimplemented")
}

// Delete implements rest.GracefulDeleter.
func (s *LegacyStore) Delete(ctx context.Context, name string, deleteValidation rest.ValidateObjectFunc, options *metav1.DeleteOptions) (runtime.Object, bool, error) {
	// info, err := request.NamespaceInfoFrom(ctx, true)
	// if err != nil {
	// 	return nil, false, err
	// }

	old, err := s.Get(ctx, name, nil)
	if err != nil {
		return old, false, err
	}

	if deleteValidation != nil {
		if err := deleteValidation(ctx, old.DeepCopyObject()); err != nil {
			return nil, false, err
		}
	}

	oldUser, ok := old.(*iamv0alpha.User)
	if !ok {
		return nil, false, fmt.Errorf("expected user but got %s", old.GetObjectKind().GroupVersionKind())
	}

	meta, err := utils.MetaAccessor(oldUser)
	if err != nil {
		return nil, false, fmt.Errorf("failed to get metadata accessor: %w", err)
	}

	err = s.userSvc.Delete(ctx, &user.DeleteUserCommand{
		UserID: meta.GetDeprecatedInternalID(),
	})
	if err != nil {
		return nil, false, fmt.Errorf("failed to delete user: %w", err)
	}

	return old, true, nil
}

func (s *LegacyStore) Get(ctx context.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {
	ns, err := request.NamespaceInfoFrom(ctx, true)
	if err != nil {
		return nil, err
	}

	found, err := s.store.ListUsers(ctx, ns, legacy.ListUserQuery{
		OrgID:      ns.OrgID,
		UID:        name,
		Pagination: common.Pagination{Limit: 1},
	})
	if found == nil || err != nil {
		return nil, resource.NewNotFound(name)
	}
	if len(found.Users) < 1 {
		return nil, resource.NewNotFound(name)
	}

	obj := toUserItem(&found.Users[0], ns.Value)
	return &obj, nil
}

func toUserItem(u *user.User, ns string) iamv0alpha.User {
	item := &iamv0alpha.User{
		ObjectMeta: metav1.ObjectMeta{
			Name:              u.UID,
			Namespace:         ns,
			ResourceVersion:   fmt.Sprintf("%d", u.Updated.UnixMilli()),
			CreationTimestamp: metav1.NewTime(u.Created),
		},
		Spec: iamv0alpha.UserSpec{
			Name:          u.Name,
			Login:         u.Login,
			Email:         u.Email,
			EmailVerified: u.EmailVerified,
			Disabled:      u.IsDisabled,
		},
	}
	obj, _ := utils.MetaAccessor(item)
	obj.SetUpdatedTimestamp(&u.Updated)
	obj.SetDeprecatedInternalID(u.ID) // nolint:staticcheck
	return *item
}
