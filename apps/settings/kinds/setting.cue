package kinds

setting: {
	kind:       "Setting"
	pluralName: "Settings"
	current:    "v0alpha1"
	versions: {
		"v0alpha1": {
			schema: {
				spec: {
					group: string
					value: string
				}
				status: {
					lastAppliedGeneration: int
				}
				metadata: {
					group: string
				}
			}
			codegen: {
				ts: {
					enabled: false
				}
				go: {
					enabled: true
				}
			}
		}
	}
}
