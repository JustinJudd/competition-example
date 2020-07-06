package event

import (
	"fmt"
)

type (
	// The Event type represents a Server Sent Event message.
	Event struct {
		Data []byte
		Type string
		Id   int
	}
)

// New creates a new event with the given type and data.
func New(t string, data []byte) *Event {
	return &Event{
		Type: t,
		Data: data,
	}
}

func (e *Event) String() string {
	var s string
	if e.Id > 0 {
		s = fmt.Sprintf("event: %s\ndata: %s\nid: %d\n\n", e.Type, e.Data, e.Id)
	} else {
		s = fmt.Sprintf("event: %s\ndata: %s\n\n", e.Type, e.Data)
	}
	return s
}
