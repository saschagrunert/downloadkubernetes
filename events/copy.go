package events

import (
	"fmt"
	"strings"
	"time"
)

// LinkCopy is an even that takes place when a link is copied on the front end.
type LinkCopy struct {
	*Event

	// An emtpy UserID indicates a non-cookied user
	UserID string
	// URL is the copied link
	URL string
}

func NewLinkCopyEvent(userID, url string) *LinkCopy {
	return &LinkCopy{
		Event: &Event{
			When: time.Now(),
		},
		UserID: userID,
		URL:    url,
	}
}

func (l *LinkCopy) InsertQueryName() string {
	return "link-copy-event-insert"
}

func (l *LinkCopy) CreateTableIfNotExistsQueries(flavor string) string {
	switch flavor {
	case "sqlite3":
		return `CREATE TABLE IF NOT EXISTS copy_link_events
(
	happened text,
	user text PRIMARY KEY,
	url text
);`
	default:
		return fmt.Sprintf("unknown flavor %s", flavor)
	}
}

func (l *LinkCopy) InsertIntoPreparedStatements(flavor string) string {
	switch flavor {
	case "sqlite3":
		return `INSERT INTO copy_link_events (
happened, user, url
)
VALUES (
	?, ?, ?
)`
	default:
		return fmt.Sprintf("unknown flavor %s", flavor)
	}
}

// RipApart returns version, os, arch, bin which may all be empty
func (l *LinkCopy) ripApart() (string, string, string, string) {
	if l.URL == "" {
		return "", "", "", ""
	}
	interesting := strings.TrimPrefix(l.URL, "https://storage.googleapis.com/kubernetes-release/release/")
	parts := strings.Split(interesting, "/")
	if len(parts) != 5 {
		return "", "", "", ""
	}
	version, os, arch, bin := parts[0], parts[2], parts[3], parts[4]
	return version, os, arch, bin
}
