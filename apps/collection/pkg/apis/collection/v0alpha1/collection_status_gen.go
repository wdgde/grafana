// Code generated - EDITING IS FUTILE. DO NOT EDIT.

package v0alpha1

// +k8s:openapi-gen=true
type CollectionstatusOperatorState struct {
	// lastEvaluation is the ResourceVersion last evaluated
	LastEvaluation string `json:"lastEvaluation"`
	// state describes the state of the lastEvaluation.
	// It is limited to three possible states for machine evaluation.
	State CollectionStatusOperatorStateState `json:"state"`
	// descriptiveState is an optional more descriptive state field which has no requirements on format
	DescriptiveState *string `json:"descriptiveState,omitempty"`
	// details contains any extra information that is operator-specific
	Details map[string]interface{} `json:"details,omitempty"`
}

// NewCollectionstatusOperatorState creates a new CollectionstatusOperatorState object.
func NewCollectionstatusOperatorState() *CollectionstatusOperatorState {
	return &CollectionstatusOperatorState{}
}

// +k8s:openapi-gen=true
type CollectionStatus struct {
	// operatorStates is a map of operator ID to operator state evaluations.
	// Any operator which consumes this kind SHOULD add its state evaluation information to this field.
	OperatorStates map[string]CollectionstatusOperatorState `json:"operatorStates,omitempty"`
	// additionalFields is reserved for future use
	AdditionalFields map[string]interface{} `json:"additionalFields,omitempty"`
}

// NewCollectionStatus creates a new CollectionStatus object.
func NewCollectionStatus() *CollectionStatus {
	return &CollectionStatus{}
}

// +k8s:openapi-gen=true
type CollectionStatusOperatorStateState string

const (
	CollectionStatusOperatorStateStateSuccess    CollectionStatusOperatorStateState = "success"
	CollectionStatusOperatorStateStateInProgress CollectionStatusOperatorStateState = "in_progress"
	CollectionStatusOperatorStateStateFailed     CollectionStatusOperatorStateState = "failed"
)
