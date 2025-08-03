package collection

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/internalversion"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/rest"

	authlib "github.com/grafana/authlib/types"

	collection "github.com/grafana/grafana/apps/collection/pkg/apis/collection/v0alpha1"
	dashboardsV1 "github.com/grafana/grafana/apps/dashboard/pkg/apis/dashboard/v1beta1"
	"github.com/grafana/grafana/pkg/registry/apps/collection/legacy"
	"github.com/grafana/grafana/pkg/services/apiserver/endpoints/request"
)

var (
	_ rest.Scoper               = (*legacyStorage)(nil)
	_ rest.SingularNameProvider = (*legacyStorage)(nil)
	_ rest.Getter               = (*legacyStorage)(nil)
	_ rest.Lister               = (*legacyStorage)(nil)
	_ rest.Storage              = (*legacyStorage)(nil)
	// _ rest.Creater              = (*legacyStorage)(nil)
	// _ rest.Updater              = (*legacyStorage)(nil)
	// _ rest.GracefulDeleter      = (*legacyStorage)(nil)
)

type legacyStorage struct {
	namespacer     request.NamespaceMapper
	tableConverter rest.TableConvertor
	sql            *legacy.LegacyStarSQL
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
	ns, err := request.NamespaceInfoFrom(ctx, false)
	if err != nil {
		return nil, err
	}

	if ns.Value == "" {
		// TODO -- make sure the user can list across *all* namespaces
		return nil, fmt.Errorf("TODO... get stars for all orgs")
	}

	list := &collection.StarsList{}
	found, rv, err := s.sql.GetStars(ctx, ns.OrgID, "")
	if err != nil {
		return nil, err
	}
	for _, v := range found {
		list.Items = append(list.Items, asResource(s.namespacer(v.OrgID), &v))
	}
	if rv > 0 {
		list.ResourceVersion = strconv.FormatInt(rv, 10)
	}
	return list, nil
}

func (s *legacyStorage) Get(ctx context.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {
	info, err := request.NamespaceInfoFrom(ctx, true)
	if err != nil {
		return nil, err
	}

	ut, uid, err := authlib.ParseTypeID(name)
	if err != nil {
		return nil, fmt.Errorf("invalid name %w", err)
	}
	if ut != authlib.TypeUser {
		return nil, fmt.Errorf("expecting name with prefix: %s", authlib.TypeUser)
	}

	found, _, err := s.sql.GetStars(ctx, info.OrgID, uid)
	if err != nil || len(found) == 0 {
		return nil, err
	}
	obj := asResource(info.Value, &found[0])
	return &obj, nil
}

func asResource(ns string, v *legacy.DashboardStars) collection.Stars {
	return collection.Stars{
		ObjectMeta: metav1.ObjectMeta{
			Name:              fmt.Sprintf("user:%s", v.UserUID),
			Namespace:         ns,
			ResourceVersion:   strconv.FormatInt(v.Last, 10),
			CreationTimestamp: metav1.NewTime(time.UnixMilli(v.First)),
		},
		Spec: collection.StarsSpec{
			Resource: []collection.StarsResource{{
				Group: dashboardsV1.APIGroup,
				Kind:  "Dashboard",
				Names: v.Dashboards,
			}},
		},
	}
}
