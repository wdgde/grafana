package collection

import (
	"context"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/internalversion"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/rest"

	collection "github.com/grafana/grafana/apps/collection/pkg/apis/collection/v0alpha1"
	"github.com/grafana/grafana/pkg/services/apiserver/endpoints/request"
)

var (
	_ rest.Scoper               = (*legacyStorage)(nil)
	_ rest.SingularNameProvider = (*legacyStorage)(nil)
	_ rest.Getter               = (*legacyStorage)(nil)
	_ rest.Lister               = (*legacyStorage)(nil)
	_ rest.Storage              = (*legacyStorage)(nil)
	_ rest.Creater              = (*legacyStorage)(nil)
	_ rest.Updater              = (*legacyStorage)(nil)
	_ rest.GracefulDeleter      = (*legacyStorage)(nil)
)

type legacyStorage struct {
	namespacer     request.NamespaceMapper
	tableConverter rest.TableConvertor
}

func (s *legacyStorage) New() runtime.Object {
	return collection.StarsKind().ZeroValue()
}

func (s *legacyStorage) Destroy() {}

func (s *legacyStorage) NamespaceScoped() bool {
	return true // namespace == org
}

func (s *legacyStorage) GetSingularName() string {
	return strings.ToLower(collection.StarsKind().Kind())
}

func (s *legacyStorage) NewList() runtime.Object {
	return collection.StarsKind().ZeroListValue()
}

func (s *legacyStorage) ConvertToTable(ctx context.Context, object runtime.Object, tableOptions runtime.Object) (*metav1.Table, error) {
	return s.tableConverter.ConvertToTable(ctx, object, tableOptions)
}

func (s *legacyStorage) List(ctx context.Context, options *internalversion.ListOptions) (runtime.Object, error) {
	// orgId, err := request.OrgIDForList(ctx)
	// if err != nil {
	// 	return nil, err
	// }

	list := &collection.StarsList{}
	// for idx := range res {
	// 	list.Items = append(list.Items, *convertToK8sResource(&res[idx], s.namespacer))
	// }
	return list, nil
}

func (s *legacyStorage) Get(ctx context.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {
	// info, err := request.NamespaceInfoFrom(ctx, true)
	// if err != nil {
	// 	return nil, err
	// }

	return nil, fmt.Errorf("todo")
}

func (s *legacyStorage) Create(ctx context.Context,
	obj runtime.Object,
	createValidation rest.ValidateObjectFunc,
	options *metav1.CreateOptions,
) (runtime.Object, error) {
	// info, err := request.NamespaceInfoFrom(ctx, true)
	// if err != nil {
	// 	return nil, err
	// }

	// p, ok := obj.(*collection.Playlist)
	// if !ok {
	// 	return nil, fmt.Errorf("expected collection?")
	// }
	// cmd, err := convertToLegacyUpdateCommand(p, info.OrgID)
	// if err != nil {
	// 	return nil, err
	// }
	// out, err := s.service.Create(ctx, &collectionsvc.CreatePlaylistCommand{
	// 	UID:      p.Name,
	// 	Name:     cmd.Name,
	// 	Interval: cmd.Interval,
	// 	Items:    cmd.Items,
	// 	OrgId:    cmd.OrgId,
	// })
	// if err != nil {
	// 	return nil, err
	// }
	// return s.Get(ctx, out.UID, nil)
	return nil, fmt.Errorf("TODO")
}

func (s *legacyStorage) Update(ctx context.Context,
	name string,
	objInfo rest.UpdatedObjectInfo,
	createValidation rest.ValidateObjectFunc,
	updateValidation rest.ValidateObjectUpdateFunc,
	forceAllowCreate bool,
	options *metav1.UpdateOptions,
) (runtime.Object, bool, error) {
	// info, err := request.NamespaceInfoFrom(ctx, true)
	// if err != nil {
	// 	return nil, false, err
	// }

	// created := false
	// old, err := s.Get(ctx, name, nil)
	// if err != nil {
	// 	return old, created, err
	// }

	// obj, err := objInfo.UpdatedObject(ctx, old)
	// if err != nil {
	// 	return old, created, err
	// }
	// p, ok := obj.(*collection.Playlist)
	// if !ok {
	// 	return nil, created, fmt.Errorf("expected collection after update")
	// }

	// cmd, err := convertToLegacyUpdateCommand(p, info.OrgID)
	// if err != nil {
	// 	return old, created, err
	// }
	// _, err = s.service.Update(ctx, cmd)
	// if err != nil {
	// 	return nil, false, err
	// }

	// r, err := s.Get(ctx, name, nil)
	// return r, created, err

	return nil, false, fmt.Errorf("TODO")
}

// GracefulDeleter
func (s *legacyStorage) Delete(ctx context.Context, name string, deleteValidation rest.ValidateObjectFunc, options *metav1.DeleteOptions) (runtime.Object, bool, error) {
	return nil, false, fmt.Errorf("TODO")
}

// CollectionDeleter
func (s *legacyStorage) DeleteCollection(ctx context.Context, deleteValidation rest.ValidateObjectFunc, options *metav1.DeleteOptions, listOptions *internalversion.ListOptions) (runtime.Object, error) {
	return nil, fmt.Errorf("DeleteCollection for collections not implemented")
}
