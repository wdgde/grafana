// Code generated - EDITING IS FUTILE. DO NOT EDIT.

package v0alpha1

// +k8s:openapi-gen=true
type UserSpec struct {
	Name          string `json:"name"`
	Login         string `json:"login"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"emailVerified"`
	Disabled      bool   `json:"disabled"`
	InternalID    int64  `json:"internalID"`
}

// NewUserSpec creates a new UserSpec object.
func NewUserSpec() *UserSpec {
	return &UserSpec{}
}
