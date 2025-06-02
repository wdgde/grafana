// Code generated - EDITING IS FUTILE. DO NOT EDIT.

package v0alpha1

// +k8s:openapi-gen=true
type CorrelationSpec struct {
	Uuid      string `json:"uuid"`
	SourceUID string `json:"sourceUID"`
	TargetUID string `json:"targetUID"`
}

// NewCorrelationSpec creates a new CorrelationSpec object.
func NewCorrelationSpec() *CorrelationSpec {
	return &CorrelationSpec{}
}
