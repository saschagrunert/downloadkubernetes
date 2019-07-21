package events

// Filter event is an event that is fired when a user clicks on a filter
type Filter struct {
	*Event

	// An emtpy UserID indicates a non-cookied user
	UserID string
	// Name is name of the filter (os, arch, etc)
	Name string
	// Value is the value (amd64, darwin, 1.14, etc)
	Value string
	// Enabled indicates if the filter was turned on or off
	Enabled bool
}
