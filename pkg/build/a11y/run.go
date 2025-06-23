package main

import (
	"dagger.io/dagger"
)

func RunTest(
	d *dagger.Client,
	grafanaService *dagger.Service,
	hostSrc *dagger.Directory,
) *dagger.Container {

	pa11yContainer := d.Container().From("grafana/docker-puppeteer:1.1.0").
		WithWorkdir("/src").
		WithFile("/src/.pa11yci-pr.conf.js", hostSrc.File(".pa11yci-pr.conf.js")).
		WithFile("/src/.pa11yci.conf.js", hostSrc.File(".pa11yci.conf.js")).
		WithEnvVariable("HOST", "grafana").
		WithEnvVariable("PORT", "3001").
		WithExec([]string{"mkdir", "-p", "./screenshots"}). // not yet exported
		WithServiceBinding("grafana", grafanaService).
		WithExec([]string{"pa11y-ci", "--config", ".pa11yci-pr.conf.js"})

	return pa11yContainer
}
