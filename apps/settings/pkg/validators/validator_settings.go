package settings

import (
	"context"
	"fmt"
	"github.com/grafana/grafana-app-sdk/app"
	"github.com/grafana/grafana-app-sdk/k8s"
	"github.com/grafana/grafana-app-sdk/logging"
	"github.com/grafana/grafana-app-sdk/resource"
	"github.com/grafana/grafana-app-sdk/simple"
	settings "github.com/grafana/grafana/apps/settings/pkg/apis/settings/v0alpha1"
	"github.com/grafana/grafana/pkg/setting"
	"gopkg.in/ini.v1"
)

func NewSettingsValidator(log logging.Logger) simple.KindValidator {
	return &simple.Validator{
		ValidateFunc: func(ctx context.Context, request *app.AdmissionRequest) error {
			logger := log.WithContext(ctx).With("validator", "settings")
			if request.Action != resource.AdmissionActionCreate && request.Action != resource.AdmissionActionUpdate {
				logger.Info("called for unsupported action", "action", request.Action)
				return nil
			}
			logger.Info("called for action", "action", request.Action)
			cast, ok := request.Object.(*settings.Setting)
			if !ok {
				return fmt.Errorf("object is not of type *settings.Setting (%s %s)", request.Object.GetName(), request.Object.GroupVersionKind().String())
			}

			config := ini.Empty()
			section, _ := config.NewSection(cast.Spec.Section)

			for key, val := range cast.Spec.Overrides {
				_, err := section.NewKey(key, val)
				if err != nil {
					return fmt.Errorf("unable to validate settings value: %w", err)
				}
			}

			// Add a basic static_root_path configuration to prevent unrelated error logs
			ensureStaticRootPath(config)

			cfg, err := setting.NewCfgFromINIFile(config)
			if err != nil {
				return k8s.NewServerResponseError(fmt.Errorf("unable to validate settings value: %w", err), 400)
			}
			if cfg == nil {
				return fmt.Errorf("empty section: %s", cast.Spec.Section)
			}
			logger.Debug("parsed Config", "cfg", cfg)

			return nil
		},
	}
}

func ensureStaticRootPath(config *ini.File) {
	if !config.HasSection("server") {
		_, _ = config.NewSection("server")
	}
	srvSection, _ := config.GetSection("server")

	if !srvSection.HasKey("static_root_path") {
		_, _ = srvSection.NewKey("static_root_path", "public")
	}
}
