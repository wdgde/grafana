package settings

import (
	"context"
	"github.com/grafana/grafana/pkg/setting"

	"k8s.io/klog/v2"

	"github.com/grafana/grafana-app-sdk/operator"
	settings "github.com/grafana/grafana/apps/settings/pkg/apis/settings/v0alpha1"
)

func NewSettingReconciler(patchClient operator.PatchClient, processConfig *setting.Cfg) (operator.Reconciler, error) {
	inner := operator.TypedReconciler[*settings.Setting]{}

	inner.ReconcileFunc = func(ctx context.Context, request operator.TypedReconcileRequest[*settings.Setting]) (operator.ReconcileResult, error) {
		switch request.Action {
		case operator.ReconcileActionCreated:
			simpleConfigCheck(request, processConfig)
			return operator.ReconcileResult{}, nil
		case operator.ReconcileActionUpdated:
			klog.InfoS("Updated resource", "name", request.Object.GetStaticMetadata().Identifier().Name)
			return operator.ReconcileResult{}, nil
		case operator.ReconcileActionDeleted:
			klog.InfoS("Deleted resource", "name", request.Object.GetStaticMetadata().Identifier().Name)
			return operator.ReconcileResult{}, nil
		case operator.ReconcileActionResynced:
			klog.InfoS("Possibly updated resource", "name", request.Object.GetStaticMetadata().Identifier().Name)
			return operator.ReconcileResult{}, nil
		case operator.ReconcileActionUnknown:
			klog.InfoS("error reconciling unknown action for Setting", "action", request.Action, "object", request.Object)
			return operator.ReconcileResult{}, nil
		}

		klog.InfoS("error reconciling invalid action for Setting", "action", request.Action, "object", request.Object)
		return operator.ReconcileResult{}, nil
	}

	// prefixing the finalizer with <group>-<kind> similar to how OpinionatedWatcher does
	reconciler, err := operator.NewOpinionatedReconciler(patchClient, "setting-settings-finalizer")
	if err != nil {
		klog.ErrorS(err, "Error creating opinionated reconciler for settings")
		return nil, err
	}
	reconciler.Reconciler = &inner
	return reconciler, nil
}

func simpleConfigCheck(request operator.TypedReconcileRequest[*settings.Setting], processConfig *setting.Cfg) {
	settingObject := request.Object
	if settingObject.Spec.Section == "server" && settingObject.Name == "enable_gzip" {
		klog.InfoS("Adding override for [server]enable_gzip", "current_value", processConfig.EnableGzip)
	}
	klog.InfoS("Added resource", "name", request.Object.GetStaticMetadata().Identifier().Name)
}
