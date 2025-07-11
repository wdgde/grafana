package kinds

setting: {
	group:      "settings.grafana.app"
	kind:       "Setting"
	pluralName: "Settings"
	current:    "v0alpha1"
	mutation: {
		operations: [
			"CREATE",
			"UPDATE",
		]
	}
	validation: {
		operations: [
			"CREATE",
			"UPDATE",
		]
	}
	versions: {
		"v0alpha1": {
			codegen: {
				ts: {
					enabled: false
				}
				go: {
					enabled: true
				}
			}
			schema: {
				#SettingsSection: {
					// Settings section
					section: string
					// Settings overrides
					overrides: [string]: string
				}
				spec: #SettingsSection
				status: {
					lastAppliedGeneration: int
				}
				metadata: {
					section: string
				}
			}
		}
	}
}
