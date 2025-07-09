package settings

import (
	"context"
	"fmt"

	"k8s.io/klog/v2"

	"github.com/grafana/grafana-app-sdk/app"
	"github.com/grafana/grafana-app-sdk/resource"
	"github.com/grafana/grafana-app-sdk/simple"
	settings "github.com/grafana/grafana/apps/settings/pkg/apis/settings/v0alpha1"
)

func NewSettingsMutator() simple.KindMutator {
	return &simple.Mutator{
		MutateFunc: func(ctx context.Context, request *app.AdmissionRequest) (*app.MutatingResponse, error) {
			if request.Action != resource.AdmissionActionCreate && request.Action != resource.AdmissionActionUpdate {
				klog.InfoS("[Mutator] Called for unsupported action", "action", request.Action)
				return &app.MutatingResponse{
					UpdatedObject: request.Object,
				}, nil
			}
			cast, ok := request.Object.(*settings.Setting)
			if !ok {
				return nil, fmt.Errorf("object is not of type *settings.Setting (%s %s)", request.Object.GetName(), request.Object.GroupVersionKind().String())
			}
			if cast.Labels == nil {
				cast.Labels = make(map[string]string)
			}
			cast.Labels["section"] = cast.Spec.Section

			if request.Action == resource.AdmissionActionCreate {
				cast.Name = cast.Spec.Section
			}
			return &app.MutatingResponse{
				UpdatedObject: cast,
			}, nil
		},
	}
}
