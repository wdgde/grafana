package main

import (
	"context"
	"os"
	"strings"

	"dagger.io/dagger"
)

type GrafanaServiceOpts struct {
	HostSrc      *dagger.Directory
	GrafanaTarGz *dagger.File
	License      *dagger.File
}

func GrafanaService(ctx context.Context, d *dagger.Client, opts GrafanaServiceOpts) (*dagger.Service, error) {
	container := d.Container().From("alpine:3").
		WithExec([]string{"apk", "add", "--no-cache", "bash", "tar", "netcat-openbsd"}).
		WithMountedFile("/src/grafana.tar.gz", opts.GrafanaTarGz).
		WithExec([]string{"mkdir", "-p", "/src/grafana"}).
		WithExec([]string{"tar", "--strip-components=1", "-xzf", "/src/grafana.tar.gz", "-C", "/src/grafana"}).
		WithDirectory("/src/grafana/devenv", opts.HostSrc.Directory("./devenv")).
		WithDirectory("/src/grafana/e2e", opts.HostSrc.Directory("./e2e")).
		WithDirectory("/src/grafana/scripts", opts.HostSrc.Directory("./scripts")).
		WithDirectory("/src/grafana/tools", opts.HostSrc.Directory("./tools")).
		WithWorkdir("/src/grafana").
		WithEnvVariable("GF_APP_MODE", "development").
		WithEnvVariable("GF_SERVER_HTTP_PORT", "3001").
		WithEnvVariable("GF_SERVER_ROUTER_LOGGING", "1").
		WithExposedPort(3001)

	var licenseArg string
	if opts.License != nil {
		container = container.WithMountedFile("/src/license.jwt", opts.License)
		licenseArg = "/src/license.jwt"
	}

	// We add all GF_ environment variables to allow for overriding Grafana configuration.
	// It is unlikely the runner has any such otherwise.
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, "GF_") {
			parts := strings.SplitN(env, "=", 2)
			container = container.WithEnvVariable(parts[0], parts[1])
		}
	}

	svc := container.AsService(dagger.ContainerAsServiceOpts{Args: []string{"bash", "-x", "scripts/grafana-server/start-server", licenseArg}})

	return svc, nil
}
