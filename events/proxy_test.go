package events_test

import (
	"sync"
	"testing"

	"github.com/chuckha/downloadkubernetes/events"
)

type mytesthandler struct {
	sync.Mutex
	called int
}

func (m *mytesthandler) Handle(*events.LinkCopy) {
	m.called++
	m.Unlock()
}
func (m *mytesthandler) ID() string {
	return "abcd"
}

func TestStartListeners(t *testing.T) {
	p := events.NewProxy()
	go p.StartListeners()
	p.WriteCopyEvent(&events.LinkCopy{})
	m := &mytesthandler{}
	if m.called != 0 {
		t.Fatal("should not have called the test handler yet")
	}
	p.RegisterCopyEventListener(m)
	m.Lock()
	p.WriteCopyEvent(&events.LinkCopy{})
	m.Lock()
	if m.called != 1 {
		t.Fatal("should have called the handler once but got", m.called)
	}
}
