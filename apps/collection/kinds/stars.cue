package collection

starsv0alpha1: {
	kind:   "Stars"
	// pluralName: "Stars"
	scope:  "Namespaced"
	validation: {
		operations: [
			"CREATE",
			"UPDATE",
		]
	}
	schema: {
		#Resource: {
			group: string
			kind: string

			// The set of resources
			names: [...string]
		}
		spec: {
			resource: [...#Resource]
		}
	}
}
