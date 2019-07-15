package models

import (
	"net/url"
	"time"
)

// Download represents a single instance of someone downloading something
type Download struct {
	User       string
	Downloaded time.Time
	FilterSet  int
	Binary     string
	Version    string
	URL        url.URL
}
