// Code generated - EDITING IS FUTILE. DO NOT EDIT.

package v0alpha1

// spec is the schema of our resource
// +k8s:openapi-gen=true
type PluginSpec struct {
	Id      string `json:"id"`
	Version string `json:"version"`
}

// NewPluginSpec creates a new PluginSpec object.
func NewPluginSpec() *PluginSpec {
	return &PluginSpec{}
}
