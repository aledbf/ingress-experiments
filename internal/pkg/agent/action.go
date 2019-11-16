// Package agent provides functions related to ingress controller clients
package agent

// Operation defines
type Operation string

const (
	// AddOperation defines an action when an agent is added
	AddOperation Operation = "add"
	// RemoveOperation defines an action when an agent is removed
	RemoveOperation Operation = "remove"
)

// Event describes a change in an agent
type Event struct {
	// Operation defines the action related to the event
	Op Operation `json:"op"`
	// Agent contains information about the pod where the agent is running
	Agent string `json:"agent"`
}

// NewEvent creates a new Event, triggered by a change in an agent
func NewEvent(op Operation, agent string) *Event {
	return &Event{
		Op:    op,
		Agent: agent,
	}
}
