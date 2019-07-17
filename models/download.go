package models

import (
	"fmt"
	"time"
)

// Download represents a single instance of someone downloading something
type Download struct {
	User            string
	Downloaded      time.Time
	FilterSet       int
	OperatingSystem string
	Architecture    string
	Version         string
	Binary          string
	URL             string
}

func (d *Download) CreateTableIfNotExistsQueries(flavor string) string {
	switch flavor {
	case "sqlite3":
		return `CREATE TABLE IF NOT EXISTS downloads
(
	user text PRIMARY KEY,
	downloaded text,
	filterset integer,
	operating_system text,
	architecture text,
	version text,
	binary text,
	url text
);`
	default:
		return fmt.Sprintf("Unknown flavor %q", flavor)
	}
}

func (d *Download) SelectRecentDownloads(flavor string) string {
	switch flavor {
	case "sqlite3":
		return `SELECT operating_system, architecture, version, binary FROM downloads WHERE user = ? LIMIT ?`
	default:
		return fmt.Sprintf("Unknown flavor %q", flavor)
	}
}

func (d *Download) InsertIntoPreparedStatements(flavor string) string {
	switch flavor {
	case "sqlite3":
		return `INSERT INTO downloads (
	user, downloaded, filterset, operating_system, architecture, version, binary, url
)
VALUES (
	?, ?, ?, ?, ?, ?, ?, ?
)`
	default:
		return fmt.Sprintf("Unknown flavor %q", flavor)
	}
}

// UserID identifies a single user. A human can have many, many IDs.
type UserID struct {
	ID         string
	CreateTime time.Time
	ExpireTime time.Time
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
