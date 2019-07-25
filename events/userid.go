package events

import (
	"fmt"
	"time"
)

const (
	Created = "created"
	Expired = "expired"
)

// UserIDC is an event that happens when a user id gets created
type UserID struct {
	*Event

	// UserID is the created user id
	UserID string
	// Action is what happened to the User ID
	Action string
}

func NewUserID(userid, action string) *UserID {
	return &UserID{
		Event: &Event{
			When: time.Now(),
		},
		UserID: userid,
		Action: action,
	}
}

func (u *UserID) InsertQueryName() string {
	return "user-id-event-insert"
}

func (l *UserID) CreateTableIfNotExistsQueries(flavor string) string {
	switch flavor {
	case "sqlite3":
		return `CREATE TABLE IF NOT EXISTS user_id_events (
happened text,
user_id text,
action text
)`
	default:
		return fmt.Sprintf("unknown flavor %s", flavor)
	}
}

func (l *UserID) InsertIntoPreparedStatements(flavor string) string {
	switch flavor {
	case "sqlite3":
		return `INSERT INTO user_id_events (happened, user_id, action) VALUES (?, ?, ?)`
	default:
		return fmt.Sprintf("unknown flavor %s", flavor)
	}
}
