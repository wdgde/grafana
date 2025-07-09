// Code generated - EDITING IS FUTILE. DO NOT EDIT.

package v0alpha1

// +k8s:openapi-gen=true
type SettingSpec struct {
	// Settings section
	Section string `json:"section"`
	// Settings overrides
	Overrides map[string]string `json:"overrides"`
}

// NewSettingSpec creates a new SettingSpec object.
func NewSettingSpec() *SettingSpec {
	return &SettingSpec{
		Overrides: map[string]string{},
	}
}
