// Code generated - EDITING IS FUTILE. DO NOT EDIT.

package v0alpha1

// +k8s:openapi-gen=true
type TeamPermissionTeamPermission string

const (
	TeamPermissionTeamPermissionAdmin  TeamPermissionTeamPermission = "admin"
	TeamPermissionTeamPermissionMember TeamPermissionTeamPermission = "member"
)

// +k8s:openapi-gen=true
type TeamPermissionSpec struct {
	Name       string                       `json:"name"`
	Permission TeamPermissionTeamPermission `json:"permission"`
}

// NewTeamPermissionSpec creates a new TeamPermissionSpec object.
func NewTeamPermissionSpec() *TeamPermissionSpec {
	return &TeamPermissionSpec{}
}
