package bakers

import (
	"math/rand"
)

const (
	charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

// DumbIdentifier randomly generates ID
type DumbIdentifier struct {
	Random   *rand.Rand
	IDLength int
}

// NewDumbIdentifier returns an Identifier with a very simple algorithm
func NewDumbIdentifier(idLength, ttl int, seed int64) *DumbIdentifier {
	return &DumbIdentifier{
		Random:   rand.New(rand.NewSource(seed)),
		IDLength: idLength,
	}
}

// DumbBaker randomly generates an ID for a user
func (d *DumbIdentifier) Identify() string {
	id := make([]byte, d.IDLength)
	for i := range id {
		id[i] = charset[d.Random.Intn(len(charset))]
	}
	return string(id)
}
