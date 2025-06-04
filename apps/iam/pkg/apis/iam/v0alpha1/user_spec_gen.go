// Code generated - EDITING IS FUTILE. DO NOT EDIT.

package v0alpha1

// +k8s:openapi-gen=true
type UserSpec struct {
	Disabled      bool   `json:"disabled"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"emailVerified"`
	Login         string `json:"login"`
	Name          string `json:"name"`
	Provisioned   bool   `json:"provisioned"`
}

// NewUserSpec creates a new UserSpec object.
func NewUserSpec() *UserSpec {
	return &UserSpec{}
}
