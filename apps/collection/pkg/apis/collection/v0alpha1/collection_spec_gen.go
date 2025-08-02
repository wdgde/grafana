// Code generated - EDITING IS FUTILE. DO NOT EDIT.

package v0alpha1

// +k8s:openapi-gen=true
type CollectionItem struct {
	Group string `json:"group"`
	Kind  string `json:"kind"`
	Name  string `json:"name"`
}

// NewCollectionItem creates a new CollectionItem object.
func NewCollectionItem() *CollectionItem {
	return &CollectionItem{}
}

// +k8s:openapi-gen=true
type CollectionSpec struct {
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Items       []CollectionItem `json:"items"`
}

// NewCollectionSpec creates a new CollectionSpec object.
func NewCollectionSpec() *CollectionSpec {
	return &CollectionSpec{
		Items: []CollectionItem{},
	}
}
