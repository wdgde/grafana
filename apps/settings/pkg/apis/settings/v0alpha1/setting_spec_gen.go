// Code generated - EDITING IS FUTILE. DO NOT EDIT.

package v0alpha1

// +k8s:openapi-gen=true
type SettingSpec struct {
	Group string `json:"group"`
	Value string `json:"value"`
}

// NewSettingSpec creates a new SettingSpec object.
func NewSettingSpec() *SettingSpec {
	return &SettingSpec{}
}
