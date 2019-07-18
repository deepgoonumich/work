package models

import (
	"time"
)

// EventVersion specifies the event schema version.
type EventVersion string

// Valid event versions.
var (
	EventV2 EventVersion = "2.0"
)

// EventType specifies the kind of event.
type EventType string

// Valid event types.
var (
	EventDocCreated EventType = "Created"
	EventDocUpdated EventType = "Updated"
	EventDocRemoved EventType = "Removed"
)

// Event represents the event schema.
type Event struct {
	// ID is an opaque event ID.
	ID string

	// CreatedAt is the time this event was created.
	CreatedAt *time.Time

	// Version is the version of the event.
	Version EventVersion

	// Type is the type of the event.
	Type EventType

	// Quiet specifies if the event should be ignored by real-time clients, i.e.
	// Benzinga Pro or TCP.
	Quiet bool

	// Document is the document being acted on. In some cases, this may be a
	// partially filled document.
	Document Document
}
