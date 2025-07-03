package assetpath

import (
	"fmt"
	"net/url"
	"path"

	"github.com/grafana/grafana/pkg/plugins"
)

// DefaultLogoPath returns the default logo path for the specified plugin type.
func DefaultLogoPath(pluginType plugins.Type) string {
	return path.Join("public/img", fmt.Sprintf("icn-%s.svg", string(pluginType)))
}

func GetTranslations(pluginJSON plugins.JSONData, n plugins.AssetInfo) (map[string]string, error) {
	pathToTranslations, err := n.RelativeURL("locales")
	if err != nil {
		return nil, fmt.Errorf("get locales: %w", err)
	}

	// loop through all the languages specified in the plugin.json and add them to the list
	translations := map[string]string{}
	for _, language := range pluginJSON.Languages {
		file := fmt.Sprintf("%s.json", pluginJSON.ID)
		translations[language], err = url.JoinPath(pathToTranslations, language, file)
		if err != nil {
			return nil, fmt.Errorf("join path: %w", err)
		}
	}

	return translations, nil
}
