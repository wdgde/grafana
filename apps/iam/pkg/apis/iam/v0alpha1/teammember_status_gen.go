// Code generated - EDITING IS FUTILE. DO NOT EDIT.

package v0alpha1

// +k8s:openapi-gen=true
type TeamMemberstatusOperatorState struct {
	// lastEvaluation is the ResourceVersion last evaluated
	LastEvaluation string `json:"lastEvaluation"`
	// state describes the state of the lastEvaluation.
	// It is limited to three possible states for machine evaluation.
	State TeamMemberStatusOperatorStateState `json:"state"`
	// descriptiveState is an optional more descriptive state field which has no requirements on format
	DescriptiveState *string `json:"descriptiveState,omitempty"`
	// details contains any extra information that is operator-specific
	Details map[string]interface{} `json:"details,omitempty"`
}

// NewTeamMemberstatusOperatorState creates a new TeamMemberstatusOperatorState object.
func NewTeamMemberstatusOperatorState() *TeamMemberstatusOperatorState {
	return &TeamMemberstatusOperatorState{}
}

// +k8s:openapi-gen=true
type TeamMemberStatus struct {
	// operatorStates is a map of operator ID to operator state evaluations.
	// Any operator which consumes this kind SHOULD add its state evaluation information to this field.
	OperatorStates map[string]TeamMemberstatusOperatorState `json:"operatorStates,omitempty"`
	// additionalFields is reserved for future use
	AdditionalFields map[string]interface{} `json:"additionalFields,omitempty"`
}

// NewTeamMemberStatus creates a new TeamMemberStatus object.
func NewTeamMemberStatus() *TeamMemberStatus {
	return &TeamMemberStatus{}
}

// +k8s:openapi-gen=true
type TeamMemberStatusOperatorStateState string

const (
	TeamMemberStatusOperatorStateStateSuccess    TeamMemberStatusOperatorStateState = "success"
	TeamMemberStatusOperatorStateStateInProgress TeamMemberStatusOperatorStateState = "in_progress"
	TeamMemberStatusOperatorStateStateFailed     TeamMemberStatusOperatorStateState = "failed"
)
