package backend_test

import (
	"testing"
	"time"

	"github.com/chuckha/downloadkubernetes/backend"
	"github.com/chuckha/downloadkubernetes/events"
)

func TestRecentsAreInOrder(t *testing.T) {
	c := backend.NewCache()
	c.HandleCopyLinkEvent(&events.LinkCopy{
		Event: &events.Event{
			When: time.Now(),
		},
		UserID: "test",
		URL:    "some url",
	})
	c.HandleCopyLinkEvent(&events.LinkCopy{
		Event: &events.Event{
			When: time.Now().Add(1 * time.Second),
		},
		UserID: "test",
		URL:    "some url 2",
	})
	c.HandleCopyLinkEvent(&events.LinkCopy{
		Event: &events.Event{
			When: time.Now().Add(2 * time.Second),
		},
		UserID: "test",
		URL:    "some url 4",
	})
	actual := c.Recents("test")
	expected := []string{"some url", "some url 2", "some url 4"}
	for i, e := range expected {
		if actual[i] != e {
			t.Fatalf("expected %v but got %v", e, actual[i])
		}
	}

}

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
