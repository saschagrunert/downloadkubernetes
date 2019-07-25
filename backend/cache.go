package backend

import (
	"container/ring"
	"sort"

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

func (c *Cache) HandleCopyLinkEvent(lc *events.LinkCopy) error {
	r, ok := c.recents[lc.UserID]
	if !ok {
		r = ring.New(5)
		c.recents[lc.UserID] = r
	}
	r.Value = *lc
	c.recents[lc.UserID] = r.Next()
	return nil
}

func (c *Cache) HandleUserIDEvent(id *events.UserID) error {
	if id.Action != events.Expired {
		return nil
	}

	delete(c.recents, id.UserID)
	return nil
}

func (c *Cache) ID() string {
	return "backend-cache"
}

func (c *Cache) Recents(uid string) []string {
	lcs := []events.LinkCopy{}
	set := map[events.LinkCopy]struct{}{}
	c.recents[uid].Do(func(item interface{}) {
		if item == nil {
			return
		}
		linkCopy := item.(events.LinkCopy)
		set[linkCopy] = struct{}{}
	})
	for k := range set {
		lcs = append(lcs, k)
	}
	sort.Sort(linkcopies(lcs))
	out := []string{}
	for _, lc := range lcs {
		out = append(out, lc.URL)
	}
	return out
}

type linkcopies []events.LinkCopy

func (l linkcopies) Len() int {
	return len(l)
}
func (l linkcopies) Less(i, j int) bool {
	return l[i].When.Before(l[j].When)
}
func (l linkcopies) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}
