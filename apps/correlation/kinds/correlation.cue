package kinds

correlation: {
	kind:		"Correlation"
	pluralName:	"Correlations"
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
