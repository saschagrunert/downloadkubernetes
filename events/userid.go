package events

import (
	"fmt"
	"time"

	"github.com/chuckha/downloadkubernetes/models"
)

const (
	Created = "created"
	Expired = "expired"
)

// UserID is an event that happens when a user id gets created
type UserID struct {
	*Event

	// User is the user being modified
	User *models.User
	// Action is what happened to the User ID
	Action string
}

func NewUserID(user *models.User, action string) *UserID {
	if user == nil {
		user = &models.User{}
	}
	return &UserID{
		Event: &Event{
			When: time.Now(),
		},
		User:   user,
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
