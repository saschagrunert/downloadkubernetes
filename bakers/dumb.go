package bakers

import (
	"math/rand"
	"net/http"
	"time"
)

const (
	charset    = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	cookieName = "downloadkubernetes"
)

// DumbBaker is a baker that is not smart.
type DumbBaker struct {
	Random   *rand.Rand
	IDLength int
	// TTL Is how long the cookie will live for in days
	TTL int
}

// NewDumbBaker bakes cookies by randomly assigning an ID
func NewDumbBaker(idLength, ttl int) *DumbBaker {
	return &DumbBaker{
		Random:   rand.New(rand.NewSource(1)),
		IDLength: idLength,
		TTL:      ttl,
	}
}

// DumbBaker randomly generates an ID for a user
func (b *DumbBaker) calculateNewID() string {
	id := make([]byte, b.IDLength)
	for i := range id {
		id[i] = charset[b.Random.Intn(len(charset))]
	}
	return string(id)
}

// NewCookieForRequest ignores the request because this baker is dumb
func (b *DumbBaker) NewCookieForRequest(_ *http.Request) *http.Cookie {
	return &http.Cookie{
		Name:     cookieName,
		Value:    b.calculateNewID(),
		Expires:  time.Now().Add(time.Duration(b.TTL) * 24 * time.Hour),
		SameSite: http.SameSiteStrictMode,
	}
}
