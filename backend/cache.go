package backend

import (
	"container/ring"

	"github.com/chuckha/downloadkubernetes/events"
)

type Cache struct {
	recents map[string]*ring.Ring
}

func NewCache() *Cache {
	return &Cache{
		recents: map[string]*ring.Ring{},
	}
}

func (c *Cache) Handle(lc *events.LinkCopy) {
	r, ok := c.recents[lc.UserID]
	if !ok {
		c.recents[lc.UserID] = ring.New(5)
		r = c.recents[lc.UserID]
	}
	r.Value = *lc
	c.recents[lc.UserID] = r.Next()
}

func (c *Cache) ID() string {
	return "backend-cache"
}

func (c *Cache) Recents(uid string) []string {
	out := []string{}
	c.recents[uid].Do(func(item interface{}) {
		if item == nil {
			return
		}
		linkCopy := item.(events.LinkCopy)
		out = append(out, linkCopy.URL)
	})
	return out
}
