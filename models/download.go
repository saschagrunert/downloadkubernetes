package models

import (
	"fmt"
	"time"
)

// Download represents a single instance of someone downloading something
type Download struct {
	// set in backend
	User string
	// set in backend
	Downloaded time.Time
	// URL is the download URL
	URL string
	// attempt at turning this into useful query params
	OperatingSystem string
	Architecture    string
	Version         string
	Binary          string
}

func (d *Download) InsertQueryName() string {
	return "download-insert"
}

func (d *Download) CreateTableIfNotExistsQueries(flavor string) string {
	switch flavor {
	case "sqlite3":
		return `CREATE TABLE IF NOT EXISTS downloads
(
	user text PRIMARY KEY,
	downloaded text,
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
		return `SELECT operating_system, architecture, version, binary FROM downloads WHERE user = ? ORDER BY downloaded LIMIT ?`
	default:
		return fmt.Sprintf("Unknown flavor %q", flavor)
	}
}

func (d *Download) InsertIntoPreparedStatements(flavor string) string {
	switch flavor {
	case "sqlite3":
		return `INSERT INTO downloads (
	user, downloaded, operating_system, architecture, version, binary, url
)
VALUES (
	?, ?, ?, ?, ?, ?, ?
)`
	default:
		return fmt.Sprintf("Unknown flavor %q", flavor)
	}
}
