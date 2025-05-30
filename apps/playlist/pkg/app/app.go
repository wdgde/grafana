package app

import (
	"bytes"
	"context"
	"fmt"

	"github.com/grafana/grafana-app-sdk/app"
	"github.com/grafana/grafana-app-sdk/k8s"
	"github.com/grafana/grafana-app-sdk/operator"
	"github.com/grafana/grafana-app-sdk/resource"
	"github.com/grafana/grafana-app-sdk/simple"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"

	"github.com/grafana/grafana/apps/playlist/pkg/reconcilers"

	playlistv0alpha1 "github.com/grafana/grafana/apps/playlist/pkg/apis/playlist/v0alpha1"
	playlistv1 "github.com/grafana/grafana/apps/playlist/pkg/apis/playlist/v1"
)

type PlaylistConfig struct {
	EnableReconcilers bool
}

func getPatchClient(restConfig rest.Config, playlistKind resource.Kind) (operator.PatchClient, error) {
	clientGenerator := k8s.NewClientRegistry(restConfig, k8s.ClientConfig{})
	return clientGenerator.ClientFor(playlistKind)
}

func New(cfg app.Config) (app.App, error) {
	var (
		playlistReconciler operator.Reconciler
		err                error
	)

	playlistConfig, ok := cfg.SpecificConfig.(*PlaylistConfig)
	if ok && playlistConfig.EnableReconcilers {
		patchClient, err := getPatchClient(cfg.KubeConfig, playlistv0alpha1.PlaylistKind())
		if err != nil {
			klog.ErrorS(err, "Error getting patch client for use with opinionated reconciler")
			return nil, err
		}

		playlistReconciler, err = reconcilers.NewPlaylistReconciler(patchClient)
		if err != nil {
			klog.ErrorS(err, "Error creating playlist reconciler")
			return nil, err
		}
	}

	simpleConfig := simple.AppConfig{
		Name:       "playlist",
		KubeConfig: cfg.KubeConfig,
		InformerConfig: simple.AppInformerConfig{
			ErrorHandler: func(ctx context.Context, err error) {
				klog.ErrorS(err, "Informer processing error")
			},
		},
		Converters: map[schema.GroupKind]simple.Converter{
			schema.GroupKind{Group: playlistv1.PlaylistKind().Group(), Kind: playlistv1.PlaylistKind().Kind()}: &PlaylistConverter{},
		},
		ManagedKinds: []simple.AppManagedKind{
			{
				Kind:       playlistv0alpha1.PlaylistKind(),
				Reconciler: playlistReconciler,
				Mutator: &simple.Mutator{
					MutateFunc: func(ctx context.Context, req *app.AdmissionRequest) (*app.MutatingResponse, error) {
						// modify req.Object if needed
						return &app.MutatingResponse{
							UpdatedObject: req.Object,
						}, nil
					},
				},
				Validator: &simple.Validator{
					ValidateFunc: func(ctx context.Context, req *app.AdmissionRequest) error {
						// do something here if needed
						return nil
					},
				},
			},
			{
				Kind:       playlistv1.PlaylistKind(),
				Reconciler: playlistReconciler,
				Mutator: &simple.Mutator{
					MutateFunc: func(ctx context.Context, req *app.AdmissionRequest) (*app.MutatingResponse, error) {
						// modify req.Object if needed
						return &app.MutatingResponse{
							UpdatedObject: req.Object,
						}, nil
					},
				},
				Validator: &simple.Validator{
					ValidateFunc: func(ctx context.Context, req *app.AdmissionRequest) error {
						// do something here if needed
						return nil
					},
				},
			},
		},
	}

	a, err := simple.NewApp(simpleConfig)
	if err != nil {
		return nil, err
	}

	err = a.ValidateManifest(cfg.ManifestData)
	if err != nil {
		return nil, err
	}

	return a, nil
}

func GetKinds() map[schema.GroupVersion][]resource.Kind {
	v0 := schema.GroupVersion{
		Group:   playlistv0alpha1.PlaylistKind().Group(),
		Version: playlistv0alpha1.PlaylistKind().Version(),
	}
	v1 := schema.GroupVersion{
		Group:   playlistv1.PlaylistKind().Group(),
		Version: playlistv1.PlaylistKind().Version(),
	}
	return map[schema.GroupVersion][]resource.Kind{
		v0: {playlistv0alpha1.PlaylistKind()},
		v1: {playlistv1.PlaylistKind()},
	}
}

// example from https://github.com/grafana/grafana-app-sdk/blob/main/docs/custom-kinds/managing-multiple-versions.md#conversion-code
type PlaylistConverter struct{}

func (p *PlaylistConverter) Convert(obj k8s.RawKind, targetAPIVersion string) ([]byte, error) {
	if targetAPIVersion == obj.APIVersion {
		return obj.Raw, nil
	}

	targetGVK := schema.FromAPIVersionAndKind(targetAPIVersion, obj.Kind)

	if obj.Version == "v1" {
		if targetGVK.Version != "v0alpha1" {
			return nil, fmt.Errorf("cannot convert into unknown version %s", targetGVK.Version)
		}
		src := playlistv1.Playlist{}
		err := playlistv1.PlaylistKind().Codecs[resource.KindEncodingJSON].Read(bytes.NewReader(obj.Raw), &src)
		if err != nil {
			return nil, fmt.Errorf("unable to parse kind")
		}
		dst := playlistv0alpha1.Playlist{}
		src.ObjectMeta.DeepCopyInto(&dst.ObjectMeta)
		dst.SetGroupVersionKind(targetGVK)
		dst.Spec.Title = src.Spec.Title
		dst.Spec.Interval = src.Spec.Interval
		dst.Spec.Items = make([]playlistv0alpha1.PlaylistItem, len(src.Spec.Items))
		for i, item := range src.Spec.Items {
			dst.Spec.Items[i] = playlistv0alpha1.PlaylistItem{
				Type:  playlistv0alpha1.PlaylistItemType(item.Type),
				Value: item.Value,
			}
		}
		buf := bytes.Buffer{}
		err = playlistv0alpha1.PlaylistKind().Codecs[resource.KindEncodingJSON].Write(&buf, &dst)
		return buf.Bytes(), err
	}

	if obj.Version == "v0alpha1" {
		if targetGVK.Version != "v1" {
			return nil, fmt.Errorf("cannot convert into unknown version %s", targetGVK.Version)
		}
		src := playlistv0alpha1.Playlist{}
		err := playlistv0alpha1.PlaylistKind().Codecs[resource.KindEncodingJSON].Read(bytes.NewReader(obj.Raw), &src)
		if err != nil {
			return nil, fmt.Errorf("unable to parse kind")
		}
		dst := playlistv1.Playlist{}
		src.ObjectMeta.DeepCopyInto(&dst.ObjectMeta)
		dst.SetGroupVersionKind(targetGVK)
		dst.Spec.Title = src.Spec.Title
		dst.Spec.Interval = src.Spec.Interval
		dst.Spec.Items = make([]playlistv1.PlaylistItem, len(src.Spec.Items))
		for i, item := range src.Spec.Items {
			if item.Type == playlistv0alpha1.PlaylistItemTypeDashboardById {
				return nil, fmt.Errorf("cannot convert dashboard by id to v1")
			}
			dst.Spec.Items[i] = playlistv1.PlaylistItem{
				Type:  playlistv1.PlaylistItemType(item.Type),
				Value: item.Value,
			}
		}
		buf := bytes.Buffer{}
		err = playlistv1.PlaylistKind().Codecs[resource.KindEncodingJSON].Write(&buf, &dst)
		return buf.Bytes(), err
	}

	return nil, fmt.Errorf("unknown source version %s", obj.Version)
}
