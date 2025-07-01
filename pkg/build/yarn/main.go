package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"

	"dagger.io/dagger"
	"github.com/urfave/cli/v3"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := NewApp().Run(ctx, os.Args); err != nil {
		cancel()
		fmt.Println(err)
		os.Exit(1)
	}
}

func NewApp() *cli.Command {
	return &cli.Command{
		Name:  "a11y",
		Usage: "Run Grafana playwright e2e tests",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:      "grafana-dir",
				Usage:     "Path to the grafana/grafana clone directory",
				Value:     ".",
				Validator: mustBeDir("grafana-dir", false, false),
				TakesFile: true,
			},
			&cli.StringFlag{
				Name:  "output",
				Usage: "Path to export node_modules.tar.gz to",
				Value: "node-modules.tar.gz",
			},
		},
		Action: run,
	}
}

func run(ctx context.Context, cmd *cli.Command) error {
	grafanaDir := cmd.String("grafana-dir")
	outPath := cmd.String("output")

	d, err := dagger.Connect(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to Dagger: %w", err)
	}

	// Minimal files needed to run yarn install
	hostSrc := d.Host().Directory(grafanaDir, dagger.HostDirectoryOpts{
		Include: []string{
			"package.json",
			"yarn.lock",
			".yarnrc.yml",
			".yarn",
			"packages/*/package.json",
			"public/app/plugins/*/*/package.json",
			"e2e/test-plugins/*/package.json",
			".nvmrc",
		},
	})

	nodeVersion := GetNodeVersion(ctx, hostSrc.File(".nvmrc"))
	if nodeVersion == "" {
		return cli.Exit("Unable to find node version", 1)
	}

	exportedPath, err := WithYarnInstall(d, WithNode(d, nodeVersion), hostSrc).
		WithExec([]string{"tar", "-czf", "node_modules.tar.gz", "node_modules"}).
		File("node_modules.tar.gz").
		Export(ctx, outPath)

	log.Printf("exportedPath %s", exportedPath)

	if err != nil {
		return fmt.Errorf("failed to export node_modules.tar.gz: %w", err)
	}

	return nil
}

func mustBeDir(arg string, emptyOk bool, notExistOk bool) func(string) error {
	return func(s string) error {
		if s == "" {
			if emptyOk {
				return nil
			}
			return cli.Exit(arg+" cannot be empty", 1)
		}
		stat, err := os.Stat(s)
		if err != nil {
			if notExistOk {
				return nil
			}
			return cli.Exit(arg+" does not exist or cannot be read: "+s, 1)
		}
		if !stat.IsDir() {
			return cli.Exit(arg+" must be a directory: "+s, 1)
		}
		return nil
	}
}

// TODO: put this in a reusable package
func GetNodeVersion(ctx context.Context, nvmrcFile *dagger.File) string {
	nvmrc, err := nvmrcFile.Contents(ctx)
	if err != nil {
		return ""
	}

	return strings.TrimSpace(strings.TrimPrefix(nvmrc, "v"))
}

func WithNode(d *dagger.Client, version string) *dagger.Container {
	nodeImage := fmt.Sprintf("node:%s-slim", strings.TrimPrefix(version, "v"))
	return d.Container().From(nodeImage)
}

func WithYarnInstall(d *dagger.Client, base *dagger.Container, yarnHostSrc *dagger.Directory) *dagger.Container {
	yarnCache := d.CacheVolume("yarn-cache")

	return base.
		WithWorkdir("/src").
		WithMountedCache("/.yarn", yarnCache).
		WithEnvVariable("YARN_CACHE_FOLDER", "/.yarn").
		// It's important to copy all files here because the whole src directory is then copied into the test runner container
		WithDirectory("/src", yarnHostSrc).
		WithExec([]string{"corepack", "enable"}).
		WithExec([]string{"corepack", "install"}).
		WithExec([]string{"yarn", "install", "--immutable"})
}
