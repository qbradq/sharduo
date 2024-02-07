package game

import "strings"

// postCreationEvent represents the event and optional string parameter to
// execute after the creation of an object.
type postCreationEvent struct {
	EventName string // Name of the event to fire
	Argument  string // String argument
}

// parsePostCreationEvent parses a creation event from a string.
func parsePostCreationEvent(es string) *postCreationEvent {
	e := &postCreationEvent{}
	parts := strings.SplitN(es, "|", 2)
	e.EventName = parts[0]
	if len(parts) == 2 {
		e.Argument = parts[1]
	}
	return e
}

// Execute executes the post creation event returning the status bool.
func (e *postCreationEvent) Execute(r any) bool {
	return ExecuteEventHandler(e.EventName, r, nil, e.Argument)
}
