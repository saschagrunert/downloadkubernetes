package models

import (
	"fmt"
	"time"
)

// UserID identifies a single user. A human can have many, many IDs.
type UserID struct {
	ID         string
	CreateTime time.Time
	ExpireTime time.Time
}

func (u *UserID) InsertQueryName() string {
	return "user-id-insert"
}

func (u *UserID) CreateTableIfNotExistsQueries(flavor string) string {
	switch flavor {
	case "sqlite3":
		return `CREATE TABLE IF NOT EXISTS user_ids
(
	id text PRIMARY KEY,
	create_time text,
	expire_time text
);`
	default:
		return fmt.Sprintf("Unknown flavor %q", flavor)
	}
}

func (u *UserID) InsertIntoPreparedStatements(flavor string) string {
	switch flavor {
	case "sqlite3":
		return `INSERT INTO user_ids (
	id, create_time, expire_time
)
VALUES (
	?, ?, ?
)`
	default:
		return fmt.Sprintf("Unknown flavor %q", flavor)
	}
}
