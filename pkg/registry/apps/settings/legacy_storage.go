package settings

import (
	"context"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/internalversion"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/rest"

	settings "github.com/grafana/grafana/apps/settings/pkg/apis/settings/v0alpha1"
	"github.com/grafana/grafana/pkg/services/apiserver/endpoints/request"
	"github.com/grafana/grafana/pkg/setting"
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

const separator = ":"

type legacyStorage struct {
	setting        *setting.Cfg
	namespacer     request.NamespaceMapper
	tableConverter rest.TableConvertor
}

func (s *legacyStorage) New() runtime.Object {
	return settings.SettingKind().ZeroValue()
}

func (s *legacyStorage) Destroy() {}

func (s *legacyStorage) NamespaceScoped() bool {
	return true // namespace == org
}

func (s *legacyStorage) GetSingularName() string {
	return strings.ToLower(settings.SettingKind().Kind())
}

func (s *legacyStorage) NewList() runtime.Object {
	return settings.SettingKind().ZeroListValue()
}

func (s *legacyStorage) ConvertToTable(ctx context.Context, object runtime.Object, tableOptions runtime.Object) (*metav1.Table, error) {
	return s.tableConverter.ConvertToTable(ctx, object, tableOptions)
}

func (s *legacyStorage) List(ctx context.Context, options *internalversion.ListOptions) (runtime.Object, error) {
	list := &settings.SettingList{}
	for _, section := range s.setting.Raw.Sections() {
		for _, key := range section.Keys() {
			list.Items = append(list.Items, settings.Setting{
				ObjectMeta: metav1.ObjectMeta{
					Name: fmt.Sprintf("%s%s%s", section.Name(), separator, key.Name()),
				},
				Spec: settings.SettingSpec{
					Group: section.Name(),
					Key:   key.Name(),
					// consider redacting sensitive values?
					Value: key.Value(),
				},
				Status: settings.SettingStatus{},
			})
		}
	}
	return list, nil
}

func (s *legacyStorage) Get(ctx context.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {
	parts := strings.Split(name, separator)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid name: %s", name)
	}
	groupName := parts[0]
	keyName := parts[1]

	section, err := s.setting.Raw.GetSection(groupName)
	if err != nil {
		return nil, fmt.Errorf("section %s not found", section)
	}

	key, err := section.GetKey(keyName)
	if err != nil {
		return nil, fmt.Errorf("key %s not found", keyName)
	}

	return &settings.Setting{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: settings.SettingSpec{
			Group: section.Name(),
			Key:   keyName,
			Value: key.Value(),
		},
		Status: settings.SettingStatus{},
	}, nil
}

func (s *legacyStorage) Create(ctx context.Context,
	obj runtime.Object,
	createValidation rest.ValidateObjectFunc,
	options *metav1.CreateOptions,
) (runtime.Object, error) {
	return nil, fmt.Errorf("not supported")
}

func (s *legacyStorage) Update(ctx context.Context,
	name string,
	objInfo rest.UpdatedObjectInfo,
	createValidation rest.ValidateObjectFunc,
	updateValidation rest.ValidateObjectUpdateFunc,
	forceAllowCreate bool,
	options *metav1.UpdateOptions,
) (runtime.Object, bool, error) {
	return nil, false, fmt.Errorf("not supported")
}

// GracefulDeleter
func (s *legacyStorage) Delete(ctx context.Context, name string, deleteValidation rest.ValidateObjectFunc, options *metav1.DeleteOptions) (runtime.Object, bool, error) {
	return nil, false, fmt.Errorf("not supported")
}

// CollectionDeleter
func (s *legacyStorage) DeleteCollection(ctx context.Context, deleteValidation rest.ValidateObjectFunc, options *metav1.DeleteOptions, listOptions *internalversion.ListOptions) (runtime.Object, error) {
	return nil, fmt.Errorf("not supported")
}
