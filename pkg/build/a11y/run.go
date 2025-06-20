package main

import (
	"context"
	"fmt"

	"dagger.io/dagger"
)

func RunTest(
	d *dagger.Client,
	svc *dagger.Service,
	src *dagger.Directory, cache *dagger.CacheVolume,
	nodeVersion, runnerFlags string) *dagger.Container {
	command := fmt.Sprintf(
		"./e2e-runner a11y --start-grafana=false"+
			" --grafana-host grafana --grafana-port 3001 %s", runnerFlags)

	return GrafanaFrontend(d, cache, nodeVersion, src).
		WithExec([]string{"/bin/sh", "-c", "apt-get update && apt-get install -y git curl"}).
		WithExec([]string{"curl", "-LO", "https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb"}).
		WithExec([]string{"apt-get", "install", "-y", "./google-chrome-stable_current_amd64.deb"}).
		WithWorkdir("/src").
		WithServiceBinding("grafana", svc).
		WithExec([]string{"mkdir", "-p", "/src/screenshots"}).
		WithExec([]string{"/bin/bash", "-c", command}, dagger.ContainerWithExecOpts{Expect: dagger.ReturnTypeAny})
}

// ExportScreenshots exports the screenshots folder from the container to the host
func ExportScreenshots(ctx context.Context, container *dagger.Container, hostPath string) error {
	_, err := container.Directory("/src/screenshots").Export(ctx, hostPath)
	return err
}
