package collection

manifest: {
	appName:       "collection"
	groupOverride: "collection.grafana.app"
	versions: {
		"v0alpha1": {
			codegen: {
				ts: {enabled: false}
				go: {enabled: true}
			}
			kinds: [
				starsv0alpha1,
			]
		}
	}
}
