package agent

type Action string

const (
	// AddAction defines an action when an agent is added
	AddAction Action = "add"
	// RemoveAction defines an action when an agent is removed
	RemoveAction Action = "remove"
)

// Event describes a change in an agent
type Event struct {
	// Operation defines the action related to the event
	Operation Action `json:"op"`
	// Agent contains information about the pod where the agent is running
	Agent string `json:"agent"`
}

// NewEvent creates a new Event, triggered by a change in an agent
func NewEvent(op Action, agent string) *Event {
	return &Event{
		Operation: op,
		Agent:     agent,
	}
}
