package backend_test

import (
	"testing"

	"github.com/chuckha/downloadkubernetes/backend"
	"github.com/chuckha/downloadkubernetes/events"
)

func TestRecents(t *testing.T) {
	c := backend.NewCache()
	c.HandleCopyLinkEvent(&events.LinkCopy{
		UserID: "test",
		URL:    "some url",
	})
	out := c.Recents("test")
	if len(out) != 1 {
		t.Fatalf("Expected exactly 1 recent download")
	}
	c.HandleCopyLinkEvent(&events.LinkCopy{
		UserID: "test",
		URL:    "some url",
	})
	c.HandleCopyLinkEvent(&events.LinkCopy{
		UserID: "test",
		URL:    "some url",
	})
	out = c.Recents("test")
	if len(out) != 3 {
		t.Fatal("Expected exactly 3 recent downloads, got", len(out))
	}
}

func TestExpired(t *testing.T) {
	c := backend.NewCache()
	c.HandleCopyLinkEvent(&events.LinkCopy{
		UserID: "test",
		URL:    "a value",
	})
	if len(c.Recents("test")) != 1 {
		t.Fatalf("needed one but got %d", len(c.Recents("test")))
	}
	c.HandleUserIDEvent(&events.UserID{
		UserID: "test",
		Action: events.Expired,
	})
	if len(c.Recents("test")) != 0 {
		t.Fatalf("wanted 0 but got %d", len(c.Recents("test")))
	}
}
