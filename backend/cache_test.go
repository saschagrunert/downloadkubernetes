package backend_test

import (
	"testing"

	"github.com/chuckha/downloadkubernetes/backend"
	"github.com/chuckha/downloadkubernetes/events"
)

func TestRecents(t *testing.T) {
	c := backend.NewCache()
	c.Handle(&events.LinkCopy{
		UserID: "test",
		URL:    "some url",
	})
	out := c.Recents("test")
	if len(out) != 1 {
		t.Fatalf("Expected exactly 1 recent download")
	}
	c.Handle(&events.LinkCopy{
		UserID: "test",
		URL:    "some url",
	})
	c.Handle(&events.LinkCopy{
		UserID: "test",
		URL:    "some url",
	})
	out = c.Recents("test")
	if len(out) != 3 {
		t.Fatal("Expected exactly 3 recent downloads, got", len(out))
	}
}
