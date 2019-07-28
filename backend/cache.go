package backend

import (
	"container/ring"
	"sort"

	"github.com/chuckha/downloadkubernetes/events"
	"github.com/pkg/errors"
)

type cacheLog interface {
	Debugf(string, ...interface{})
}

// Cache is responsible for temporary state.
// Temporary state is any state that goes away between server reloads/deployments.
type Cache struct {
	recents map[string]*ring.Ring
	log     cacheLog
}

type store interface {
	FetchClicksForUnexpiredUsers() ([]*events.LinkCopy, error)
}

// NewCache creates a cache
func NewCache(store store, log cacheLog) (*Cache, error) {
	c := &Cache{
		recents: map[string]*ring.Ring{},
		log:     log,
	}
	// Hydrate the cache
	links, err := store.FetchClicksForUnexpiredUsers()
	c.log.Debugf("hydrating cache with %d entries", len(links))
	if err != nil {
		return nil, err
	}
	for _, lc := range links {
		if err := c.HandleCopyLinkEvent(lc); err != nil {
			return nil, errors.WithStack(err)
		}
	}
	return c, nil
}

// HandleCopyLinkEvent responds to any LinkCopy events.
// The cache will build up a datastructure of link copy events as time goes on.
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

// HandleUserIDEvent will clear out any expired user cache.
// This is an effort to reduce memory footprint.
func (c *Cache) HandleUserIDEvent(id *events.UserID) error {
	if id.Action != events.Expired {
		return nil
	}

	delete(c.recents, id.User.ID)
	return nil
}

// ID identifies what service this is.
func (c *Cache) ID() string {
	return "backend-cache"
}

// Recents returns the most recently clicked links found in the cache.
func (c *Cache) Recents(uid string) []string {
	lcs := []events.LinkCopy{}
	dupes := map[string]struct{}{}
	set := map[events.LinkCopy]struct{}{}
	c.recents[uid].Do(func(item interface{}) {
		if item == nil {
			return
		}
		linkCopy := item.(events.LinkCopy)
		if _, ok := dupes[linkCopy.URL]; ok {
			return
		}
		dupes[linkCopy.URL] = struct{}{}
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
	return l[i].Created < l[j].Created
}
func (l linkcopies) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}
