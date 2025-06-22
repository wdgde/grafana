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
		"yarn pa11y-ci %s > /src/pa11y-ci-results.json", runnerFlags)

	return GrafanaFrontend(d, cache, nodeVersion, src).
		WithExec([]string{"/bin/sh", "-c", "apt-get update && apt-get install -y git curl"}).
		WithExec([]string{"curl", "-LO", "https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb"}).
		WithExec([]string{"apt-get", "install", "-y", "./google-chrome-stable_current_amd64.deb"}).
		WithWorkdir("/src").
		WithServiceBinding("grafana", svc).
		WithExec([]string{"mkdir", "-p", "/src/screenshots"}).
		WithExec([]string{"yarn", "pa11y-ci", "--version"}).
		WithExec([]string{"yarn", "why", "pa11y-ci", "-R"}).
		WithEnvVariable("HOST", "grafana").
		WithEnvVariable("PORT", "3001").
		WithExec([]string{"/bin/bash", "-c", command}, dagger.ContainerWithExecOpts{Expect: dagger.ReturnTypeAny})
}

func ExportScreenshots(ctx context.Context, container *dagger.Container, hostPath string) error {
	_, err := container.Directory("/src/screenshots").Export(ctx, hostPath)
	return err
}

func ExportReport(ctx context.Context, container *dagger.Container, hostPath string) error {
	_, err := container.File("/src/pa11y-ci-results.json").Export(ctx, hostPath)
	return err
}
