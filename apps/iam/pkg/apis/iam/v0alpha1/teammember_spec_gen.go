// Code generated - EDITING IS FUTILE. DO NOT EDIT.

package v0alpha1

// +k8s:openapi-gen=true
type TeamMemberTeamPermission string

const (
	TeamMemberTeamPermissionAdmin  TeamMemberTeamPermission = "admin"
	TeamMemberTeamPermissionMember TeamMemberTeamPermission = "member"
)

// +k8s:openapi-gen=true
type TeamMemberSpec struct {
	External   bool                     `json:"external"`
	Permission TeamMemberTeamPermission `json:"permission"`
}

// NewTeamMemberSpec creates a new TeamMemberSpec object.
func NewTeamMemberSpec() *TeamMemberSpec {
	return &TeamMemberSpec{}
}
