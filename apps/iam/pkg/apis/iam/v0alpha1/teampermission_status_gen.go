// Code generated - EDITING IS FUTILE. DO NOT EDIT.

package v0alpha1

// +k8s:openapi-gen=true
type TeamPermissionstatusOperatorState struct {
	// lastEvaluation is the ResourceVersion last evaluated
	LastEvaluation string `json:"lastEvaluation"`
	// state describes the state of the lastEvaluation.
	// It is limited to three possible states for machine evaluation.
	State TeamPermissionStatusOperatorStateState `json:"state"`
	// descriptiveState is an optional more descriptive state field which has no requirements on format
	DescriptiveState *string `json:"descriptiveState,omitempty"`
	// details contains any extra information that is operator-specific
	Details map[string]interface{} `json:"details,omitempty"`
}

// NewTeamPermissionstatusOperatorState creates a new TeamPermissionstatusOperatorState object.
func NewTeamPermissionstatusOperatorState() *TeamPermissionstatusOperatorState {
	return &TeamPermissionstatusOperatorState{}
}

// +k8s:openapi-gen=true
type TeamPermissionStatus struct {
	// operatorStates is a map of operator ID to operator state evaluations.
	// Any operator which consumes this kind SHOULD add its state evaluation information to this field.
	OperatorStates map[string]TeamPermissionstatusOperatorState `json:"operatorStates,omitempty"`
	// additionalFields is reserved for future use
	AdditionalFields map[string]interface{} `json:"additionalFields,omitempty"`
}

// NewTeamPermissionStatus creates a new TeamPermissionStatus object.
func NewTeamPermissionStatus() *TeamPermissionStatus {
	return &TeamPermissionStatus{}
}

// +k8s:openapi-gen=true
type TeamPermissionStatusOperatorStateState string

const (
	TeamPermissionStatusOperatorStateStateSuccess    TeamPermissionStatusOperatorStateState = "success"
	TeamPermissionStatusOperatorStateStateInProgress TeamPermissionStatusOperatorStateState = "in_progress"
	TeamPermissionStatusOperatorStateStateFailed     TeamPermissionStatusOperatorStateState = "failed"
)
