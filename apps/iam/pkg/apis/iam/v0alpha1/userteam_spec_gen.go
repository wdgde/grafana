// Code generated - EDITING IS FUTILE. DO NOT EDIT.

package v0alpha1

// +k8s:openapi-gen=true
type UserTeamTeamPermissionSpec struct {
	Name       string                 `json:"name"`
	Permission UserTeamTeamPermission `json:"permission"`
}

// NewUserTeamTeamPermissionSpec creates a new UserTeamTeamPermissionSpec object.
func NewUserTeamTeamPermissionSpec() *UserTeamTeamPermissionSpec {
	return &UserTeamTeamPermissionSpec{}
}

// +k8s:openapi-gen=true
type UserTeamTeamPermission string

const (
	UserTeamTeamPermissionAdmin  UserTeamTeamPermission = "admin"
	UserTeamTeamPermissionMember UserTeamTeamPermission = "member"
)

// +k8s:openapi-gen=true
type UserTeamSpec struct {
	Title      string                      `json:"title"`
	TeamRef    UserTeamV0alpha1SpecTeamRef `json:"teamRef"`
	Permission UserTeamTeamPermissionSpec  `json:"permission"`
}

// NewUserTeamSpec creates a new UserTeamSpec object.
func NewUserTeamSpec() *UserTeamSpec {
	return &UserTeamSpec{
		TeamRef:    *NewUserTeamV0alpha1SpecTeamRef(),
		Permission: *NewUserTeamTeamPermissionSpec(),
	}
}

// +k8s:openapi-gen=true
type UserTeamV0alpha1SpecTeamRef struct {
	// uid of the team
	Name string `json:"name"`
}

// NewUserTeamV0alpha1SpecTeamRef creates a new UserTeamV0alpha1SpecTeamRef object.
func NewUserTeamV0alpha1SpecTeamRef() *UserTeamV0alpha1SpecTeamRef {
	return &UserTeamV0alpha1SpecTeamRef{}
}
