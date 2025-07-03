// Code generated - EDITING IS FUTILE. DO NOT EDIT.

package v0alpha1

// +k8s:openapi-gen=true
type SettingstatusOperatorState struct {
	// lastEvaluation is the ResourceVersion last evaluated
	LastEvaluation string `json:"lastEvaluation"`
	// state describes the state of the lastEvaluation.
	// It is limited to three possible states for machine evaluation.
	State SettingStatusOperatorStateState `json:"state"`
	// descriptiveState is an optional more descriptive state field which has no requirements on format
	DescriptiveState *string `json:"descriptiveState,omitempty"`
	// details contains any extra information that is operator-specific
	Details map[string]interface{} `json:"details,omitempty"`
}

// NewSettingstatusOperatorState creates a new SettingstatusOperatorState object.
func NewSettingstatusOperatorState() *SettingstatusOperatorState {
	return &SettingstatusOperatorState{}
}

// +k8s:openapi-gen=true
type SettingStatus struct {
	LastAppliedGeneration int64 `json:"lastAppliedGeneration"`
	// operatorStates is a map of operator ID to operator state evaluations.
	// Any operator which consumes this kind SHOULD add its state evaluation information to this field.
	OperatorStates map[string]SettingstatusOperatorState `json:"operatorStates,omitempty"`
	// additionalFields is reserved for future use
	AdditionalFields map[string]interface{} `json:"additionalFields,omitempty"`
}

// NewSettingStatus creates a new SettingStatus object.
func NewSettingStatus() *SettingStatus {
	return &SettingStatus{}
}

// +k8s:openapi-gen=true
type SettingStatusOperatorStateState string

const (
	SettingStatusOperatorStateStateSuccess    SettingStatusOperatorStateState = "success"
	SettingStatusOperatorStateStateInProgress SettingStatusOperatorStateState = "in_progress"
	SettingStatusOperatorStateStateFailed     SettingStatusOperatorStateState = "failed"
)
