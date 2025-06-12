package enhancedregistry

import (
	"context"
	"errors"
	"fmt"
	"github.com/grafana/authlib/types"
	"github.com/grafana/grafana-app-sdk/k8s"
	sdkresource "github.com/grafana/grafana-app-sdk/resource"
	"github.com/grafana/grafana/pkg/plugins/manager/registry"
	"github.com/grafana/grafana/pkg/services/apiserver"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/endpoints/request"
	"strconv"
	"sync"

	pluginsv0alpha1 "github.com/grafana/grafana/apps/plugins/pkg/apis/plugins/v0alpha1"
	"github.com/grafana/grafana/pkg/plugins"
	"github.com/grafana/grafana/pkg/plugins/log"
	"github.com/grafana/grafana/pkg/setting"
)

// EnhancedRegistry is a registry that uses Kubernetes as the source of truth for plugins.
type EnhancedRegistry struct {
	pluginClient sdkresource.Client
	restConfig   apiserver.RestConfigProvider
	cfg          *setting.Cfg

	mem *registry.InMemory

	log      log.Logger
	initOnce sync.Once
	initErr  error
}

// ProvideService returns a new enhanced registry service.
func ProvideService(cfg *setting.Cfg, restCfgProvider apiserver.RestConfigProvider) (*EnhancedRegistry, error) {
	return NewEnhancedRegistry(cfg, restCfgProvider), nil
}

// NewEnhancedRegistry creates a new enhanced registry.
func NewEnhancedRegistry(cfg *setting.Cfg, restConfig apiserver.RestConfigProvider) *EnhancedRegistry {
	return &EnhancedRegistry{
		cfg:        cfg,
		restConfig: restConfig,
		mem:        registry.NewInMemory(),
		log:        log.New("plugins.registry.enhanced"),
	}
}

// ensureClient ensures the plugin client is initialized
func (r *EnhancedRegistry) ensureClient(ctx context.Context) error {
	r.initOnce.Do(func() {
		kubeConfig, err := r.restConfig.GetRestConfig(ctx)
		if err != nil {
			r.initErr = err
			return
		}
		clientGenerator := k8s.NewClientRegistry(*kubeConfig, k8s.ClientConfig{})
		r.pluginClient, err = clientGenerator.ClientFor(pluginsv0alpha1.PluginKind())
		if err != nil {
			r.initErr = fmt.Errorf("failed to create plugin client: %w", err)
			return
		}
		r.initErr = nil
	})
	return r.initErr
}

// Plugin returns a plugin by its ID.
func (r *EnhancedRegistry) Plugin(ctx context.Context, pluginID string, version string) (*plugins.Plugin, bool) {
	if err := r.ensureClient(ctx); err != nil {
		r.log.Error("Failed to initialize plugin client", "error", err)
		return nil, false
	}

	namespace, err := r.getNamespace(ctx)
	if err != nil {
		r.log.Error("Failed to get namespace")
		return nil, false
	}

	identifier := sdkresource.Identifier{
		Name:      pluginID,
		Namespace: namespace,
	}

	_, err = r.pluginClient.Get(ctx, identifier)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			r.log.Error("Failed to get plugin from Kubernetes", "plugin", pluginID, "error", err)
		}
		return nil, false
	}

	return r.mem.Plugin(ctx, pluginID, version)

	// Convert Kubernetes resource to plugin
	//if plugin, ok := pluginResource.(*pluginsv0alpha1.Plugin); ok {
	//	// Create a new plugin instance from the Kubernetes resource
	//	return specToPlugin(plugin.Spec), true
	//}
	//return nil, false
}

// Plugins returns all plugins.
func (r *EnhancedRegistry) Plugins(ctx context.Context) []*plugins.Plugin {
	if err := r.ensureClient(ctx); err != nil {
		r.log.Error("Failed to initialize plugin client", "error", err)
		return nil
	}

	var ps []*plugins.Plugin

	namespace, err := r.getNamespace(ctx)
	if err != nil {
		r.log.Error("Failed to get namespace")
		return ps
	}

	// List all plugins in the namespace
	_, err = r.pluginClient.List(ctx, namespace, sdkresource.ListOptions{})
	if err != nil {
		r.log.Error("Failed to list plugins from Kubernetes", "error", err)
		return ps
	}

	return r.mem.Plugins(ctx)
	//for _, plugin := range pluginList.GetItems() {
	//	p, ok := plugin.(*pluginsv0alpha1.Plugin)
	//	if !ok {
	//		r.log.Error("Failed to get plugin spec from Kubernetes", "error", err)
	//		continue
	//	}
	//	ps = append(ps, specToPlugin(p.Spec))
	//}
	//
	//return ps
}

// Add adds a plugin to the registry.
func (r *EnhancedRegistry) Add(ctx context.Context, p *plugins.Plugin) error {
	if err := r.ensureClient(ctx); err != nil {
		return fmt.Errorf("failed to initialize plugin client: %w", err)
	}

	namespace, err := r.getNamespace(ctx)
	if err != nil {
		return err
	}

	// Create plugin identifier
	identifier := sdkresource.Identifier{
		Name:      p.ID,
		Namespace: namespace,
	}

	// Create plugin resource
	pluginResource := &pluginsv0alpha1.Plugin{
		TypeMeta: metav1.TypeMeta{
			Kind:       pluginsv0alpha1.PluginKind().Kind(),
			APIVersion: pluginsv0alpha1.PluginKind().Version(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      p.ID,
			Namespace: namespace,
		},
		Spec: pluginsv0alpha1.PluginSpec{
			Id:      p.JSONData.ID,
			Version: p.JSONData.Info.Version,
			Type:    string(p.JSONData.Type),
		},
	}

	var statusErr *apierrors.StatusError
	// Check if plugin exists
	existingPlugin, err := r.pluginClient.Get(ctx, identifier)
	if err != nil {
		if !errors.As(err, &statusErr) || statusErr.ErrStatus.Code != 404 {
			r.log.Error("Failed to get plugin resource", "plugin", p.ID, "error", err)
			return err
		}
		// Plugin doesn't exist, create it
		_, err = r.pluginClient.Create(ctx, identifier, pluginResource, sdkresource.CreateOptions{})
		if err != nil {
			r.log.Error("Failed to create plugin resource", "plugin", p.ID, "error", err)
			return err
		}
	} else {
		// Update existing plugin
		if meta, ok := existingPlugin.(metav1.Object); ok {
			pluginResource.ObjectMeta.UID = meta.GetUID()
			pluginResource.ObjectMeta.ResourceVersion = meta.GetResourceVersion()
		}
		_, err := r.pluginClient.Update(ctx, identifier, pluginResource, sdkresource.UpdateOptions{})
		if err != nil {
			r.log.Error("Failed to update plugin resource", "plugin", p.ID, "error", err)
			return err
		}
	}

	return r.mem.Add(ctx, p)
}

// Remove removes a plugin from the registry.
func (r *EnhancedRegistry) Remove(ctx context.Context, pluginID string, version string) error {
	if err := r.ensureClient(ctx); err != nil {
		return fmt.Errorf("failed to initialize plugin client: %w", err)
	}

	namespace, err := r.getNamespace(ctx)
	if err != nil {
		return err
	}

	identifier := sdkresource.Identifier{
		Name:      pluginID,
		Namespace: namespace,
	}

	err = r.pluginClient.Delete(ctx, identifier, sdkresource.DeleteOptions{})
	if err != nil {
		r.log.Error("Failed to delete plugin resource", "plugin", pluginID, "error", err)
		return err
	}

	return r.mem.Remove(ctx, pluginID, version)
}

func (r *EnhancedRegistry) getNamespace(ctx context.Context) (string, error) {
	namespace, ok := request.NamespaceFrom(ctx)
	if ok {
		return namespace, nil
	}

	if r.cfg.StackID == "" {
		return metav1.NamespaceDefault, nil
	}
	stackID, err := strconv.ParseInt(r.cfg.StackID, 10, 64)
	if err != nil {
		return "", fmt.Errorf("invalid stack id: %s", r.cfg.StackID)
	}
	return types.CloudNamespaceFormatter(stackID), nil
}

func specToPlugin(spec pluginsv0alpha1.PluginSpec) *plugins.Plugin {
	p := &plugins.Plugin{
		JSONData: plugins.JSONData{
			ID:   spec.Id,
			Type: plugins.Type(spec.Type),
			Info: plugins.Info{
				Description: "Plugin loaded from Kubernetes",
				Version:     spec.Version,
			},
		},
	}
	p.SetLogger(log.New(fmt.Sprintf("plugin.%s", spec.Id)))
	return p
}
