package sources

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/grafana/grafana/pkg/infra/fs"
	"github.com/grafana/grafana/pkg/plugins"
	"github.com/grafana/grafana/pkg/plugins/config"
	"github.com/grafana/grafana/pkg/plugins/log"
	"github.com/grafana/grafana/pkg/plugins/manager/pluginassets"
	"github.com/grafana/grafana/pkg/plugins/pluginscdn"
	"github.com/grafana/grafana/pkg/util"
)

var walk = util.Walk

var (
	ErrInvalidPluginJSONFilePath = errors.New("invalid plugin.json filepath was provided")
	logger                       = log.New("local.source")
)

type LocalSource struct {
	paths         []string
	class         plugins.Class
	assetProvider plugins.PluginAssetProvider
	strictMode    bool // If true, tracks files via a StaticFS
}

// NewLocalSource represents a plugin with a fixed set of files.
func NewLocalSource(class plugins.Class, paths []string, cfg *config.PluginManagementCfg) *LocalSource {
	return newLocalSource(paths, class, nil, cfg)
}

func newLocalSource(paths []string, class plugins.Class, assetProvider plugins.PluginAssetProvider,
	cfg *config.PluginManagementCfg) *LocalSource {
	if assetProvider == nil {
		assetProvider = pluginassets.NewLocalExternal(pluginscdn.ProvideService(cfg))
	}
	if class == plugins.ClassCore {
		assetProvider = pluginassets.NewLocalCore()
	}
	return &LocalSource{
		paths:         paths,
		class:         class,
		assetProvider: assetProvider,
		strictMode:    !cfg.DevMode,
	}
}

func (s *LocalSource) PluginClass(_ context.Context) plugins.Class {
	return s.class
}

// Paths returns the file system paths that this source will search for plugins.
// This method is primarily intended for testing purposes.
func (s *LocalSource) Paths() []string {
	return s.paths
}

func (s *LocalSource) DefaultSignature(_ context.Context, _ string) (plugins.Signature, bool) {
	switch s.class {
	case plugins.ClassCore:
		return plugins.Signature{
			Status: plugins.SignatureStatusInternal,
		}, true
	default:
		return plugins.Signature{}, false
	}
}

func (s *LocalSource) Discover(_ context.Context) ([]*plugins.FoundBundle, error) {
	if len(s.paths) == 0 {
		return []*plugins.FoundBundle{}, nil
	}

	pluginJSONPaths := make([]string, 0, len(s.paths))
	for _, path := range s.paths {
		exists, err := fs.Exists(path)
		if err != nil {
			logger.Warn("Skipping finding plugins as an error occurred", "path", path, "error", err)
			continue
		}
		if !exists {
			logger.Warn("Skipping finding plugins as directory does not exist", "path", path)
			continue
		}

		paths, err := getAbsPluginJSONPaths(path)
		if err != nil {
			return nil, err
		}
		pluginJSONPaths = append(pluginJSONPaths, paths...)
	}

	// load plugin.json files and map directory to JSON data
	foundPlugins := make(map[string]plugins.JSONData)
	for _, pluginJSONPath := range pluginJSONPaths {
		plugin, err := readPluginJSON(pluginJSONPath)
		if err != nil {
			logger.Warn("Skipping plugin loading as its plugin.json could not be read", "path", pluginJSONPath, "error", err)
			continue
		}

		pluginJSONAbsPath, err := filepath.Abs(pluginJSONPath)
		if err != nil {
			logger.Warn("Skipping plugin loading as absolute plugin.json path could not be calculated", "pluginId", plugin.ID, "error", err)
			continue
		}

		foundPlugins[filepath.Dir(pluginJSONAbsPath)] = plugin
	}

	res := make(map[string]*plugins.FoundBundle)
	for pluginDir, data := range foundPlugins {
		var pluginFs plugins.FS
		pluginFs = plugins.NewLocalFS(pluginDir)
		if s.strictMode {
			// Tighten up security by allowing access only to the files present up to this point.
			// Any new file "sneaked in" won't be allowed and will act as if the file does not exist.
			var err error
			pluginFs, err = plugins.NewStaticFS(pluginFs)
			if err != nil {
				return nil, err
			}
		}
		res[pluginDir] = &plugins.FoundBundle{
			Primary: plugins.FoundPlugin{
				JSONData: data,
				FS:       pluginFs,
			},
		}
	}

	// Track child plugins and add them to their parent.
	childPlugins := make(map[string]struct{})
	for dir, p := range res {
		// Check if this plugin is the parent of another plugin.
		for dir2, p2 := range res {
			if dir == dir2 {
				continue
			}

			relPath, err := filepath.Rel(dir, dir2)
			if err != nil {
				logger.Error("Cannot calculate relative path. Skipping", "pluginId", p2.Primary.JSONData.ID, "err", err)
				continue
			}
			if !strings.Contains(relPath, "..") {
				child := p2.Primary
				logger.Debug("Adding child", "parent", p.Primary.JSONData.ID, "child", child.JSONData.ID, "relPath", relPath)
				p.Children = append(p.Children, &child)
				childPlugins[dir2] = struct{}{}
			}
		}
	}

	// Remove child plugins from the result (they are already tracked via their parent).
	result := make([]*plugins.FoundBundle, 0, len(res))
	for k := range res {
		if _, ok := childPlugins[k]; !ok {
			result = append(result, res[k])
		}
	}

	return result, nil
}

func (s *LocalSource) AssetProvider(_ context.Context) plugins.PluginAssetProvider {
	return s.assetProvider
}

func readPluginJSON(pluginJSONPath string) (plugins.JSONData, error) {
	reader, err := readFile(pluginJSONPath)
	defer func() {
		if reader == nil {
			return
		}
		if err = reader.Close(); err != nil {
			logger.Warn("Failed to close plugin JSON file", "path", pluginJSONPath, "error", err)
		}
	}()
	if err != nil {
		logger.Warn("Skipping plugin loading as its plugin.json could not be read", "path", pluginJSONPath, "error", err)
		return plugins.JSONData{}, err
	}
	plugin, err := plugins.ReadPluginJSON(reader)
	if err != nil {
		logger.Warn("Skipping plugin loading as its plugin.json could not be read", "path", pluginJSONPath, "error", err)
		return plugins.JSONData{}, err
	}

	return plugin, nil
}

func getAbsPluginJSONPaths(path string) ([]string, error) {
	var pluginJSONPaths []string

	var err error
	path, err = filepath.Abs(path)
	if err != nil {
		return []string{}, err
	}

	if err = walk(path, true, true,
		func(currentPath string, fi os.FileInfo, err error) error {
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					logger.Error("Couldn't scan directory since it doesn't exist", "pluginDir", path, "error", err)
					return nil
				}
				if errors.Is(err, os.ErrPermission) {
					logger.Error("Couldn't scan directory due to lack of permissions", "pluginDir", path, "error", err)
					return nil
				}

				return fmt.Errorf("filepath.Walk reported an error for %q: %w", currentPath, err)
			}

			if fi.Name() == "node_modules" {
				return util.ErrWalkSkipDir
			}

			if fi.IsDir() {
				return nil
			}

			if fi.Name() != "plugin.json" {
				return nil
			}

			pluginJSONPaths = append(pluginJSONPaths, currentPath)
			return nil
		}); err != nil {
		return []string{}, err
	}

	return pluginJSONPaths, nil
}

func readFile(pluginJSONPath string) (io.ReadCloser, error) {
	logger.Debug("Loading plugin", "path", pluginJSONPath)

	if !strings.EqualFold(filepath.Ext(pluginJSONPath), ".json") {
		return nil, ErrInvalidPluginJSONFilePath
	}

	absPluginJSONPath, err := filepath.Abs(pluginJSONPath)
	if err != nil {
		return nil, err
	}

	// Wrapping in filepath.Clean to properly handle
	// gosec G304 Potential file inclusion via variable rule.
	return os.Open(filepath.Clean(absPluginJSONPath))
}

func DirAsLocalSources(cfg *config.PluginManagementCfg, pluginsPath string, class plugins.Class) ([]*LocalSource, error) {
	pluginDirs, err := ReadDir(pluginsPath)
	if err != nil {
		return nil, err
	}

	sources := make([]*LocalSource, len(pluginDirs))
	for i, dir := range pluginDirs {
		sources[i] = NewLocalSource(class, []string{dir}, cfg)
	}

	return sources, nil
}

func ReadDir(pluginsPath string) ([]string, error) {
	if pluginsPath == "" {
		return []string{}, errors.New("plugins path not configured")
	}

	// It's safe to ignore gosec warning G304 since the variable part of the file path comes from a configuration
	// variable.
	// nolint:gosec
	d, err := os.ReadDir(pluginsPath)
	if err != nil {
		return []string{}, errors.New("failed to open plugins path")
	}

	var pluginDirs []string
	for _, dir := range d {
		if dir.IsDir() || dir.Type()&os.ModeSymlink == os.ModeSymlink {
			pluginDirs = append(pluginDirs, filepath.Join(pluginsPath, dir.Name()))
		}
	}
	slices.Sort(pluginDirs)

	return pluginDirs, nil
}
