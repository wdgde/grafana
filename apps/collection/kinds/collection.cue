package collection

collectionv0alpha1: {
	kind:   "Collection"
	plural: "collections"
	scope:  "Namespaced"
	validation: {
		operations: [
			"CREATE",
			"UPDATE",
		]
	}
	schema: {
		#Item: {
			group: string
			kind: string
			name: string 
		}
		spec: {
			title:    string
			description: string
			items: [...#Item]
		}
	}
}
