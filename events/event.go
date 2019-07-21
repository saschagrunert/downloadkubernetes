package events

import "time"

// Event is the base event that gets embeded in all events
type Event struct {
	// When is when the event happened
	When time.Time
}
