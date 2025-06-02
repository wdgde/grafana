package kinds

correlation: {
	kind:       "Correlation"
	pluralName: "Correlations"
	current:    "v0alpha1"
	versions: {
		"v0alpha1": {
			codegen: {
				frontend: false
				backend:  true
			}
			schema: {
				spec: {
					uuid:    string
					sourceUID: string
					targetUID: string
				}
			}
		}
	}
}
