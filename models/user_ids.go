package models

import (
	"fmt"
	"time"
)

// User identifies a single user. A human can have many, many IDs.
type User struct {
	ID         string
	CreateTime time.Time
	ExpireTime time.Time
}

func (u *User) InsertQueryName() string {
	return "user-id-insert"
}

func (u *User) CreateTableIfNotExistsQueries(flavor string) string {
	switch flavor {
	case "sqlite3":
		return `CREATE TABLE IF NOT EXISTS user_ids
(
	id text PRIMARY KEY,
	create_time integer,
	expire_time integer
);`
	default:
		return fmt.Sprintf("Unknown flavor %q", flavor)
	}
}

func (u *User) InsertIntoPreparedStatements(flavor string) string {
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

func (u *User) ExpireUserPreparedStatement(flavor string) string {
	switch flavor {
	case "sqlite3":
		return `UPDATE user_ids
SET expire_time = ?
WHERE id = ?`
	default:
		return fmt.Sprintf("Unknown flavor %q", flavor)
	}
}

func (u *User) FetchActiveUsersClicksStatement(flavor string) string {
	switch flavor {
	case "sqlite3":
		return `SELECT u.id, c.url, c.created
FROM user_ids AS u JOIN copy_link_events as c
ON u.id = c.user
WHERE u.expire_time > ?`
	default:
		return fmt.Sprintf("Unknown flavor %q", flavor)
	}
}
