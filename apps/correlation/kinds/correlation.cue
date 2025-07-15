package kinds

correlation: {
	kind:		"Correlation"  // note: must be uppercase
	pluralName:	"Correlations" // note: must be uppercase
	current:	"v0alpha1"
	versions: {
		"v0alpha1": {
			codegen: {
				frontend: true
				backend:  true
			}
			schema: {
				spec: {
					source_uid: 	string
					target_uid: 	string
					label: 	string
					description: 	string
                    		config: string
                    		provisioned: int
                    		type: string
				}
			}
		}
	}
}