package backend_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chuckha/downloadkubernetes/backend"
	"github.com/chuckha/downloadkubernetes/events"
)

type myproxy struct {
	lc *events.LinkCopy
}

func (m *myproxy) WriteCopyEvent(lc *events.LinkCopy) {
	m.lc = lc
}
func (m *myproxy) WriteUserIDEvent(*events.UserID) {}

func TestCopyLinkEvent(t *testing.T) {
	myProxy := &myproxy{}
	server := backend.NewServer(
		backend.WithProxy(myProxy),
	)
	rec := httptest.NewRecorder()
	var buf bytes.Buffer
	buf.WriteString(`{"url": "test"}`)
	server.CopyLinkEvent(rec, &http.Request{
		Method: http.MethodPost,
		Body:   mycloser{&buf},
	})
	if rec.Code != 200 {
		t.Fatalf("did not get a 200, got a %d", rec.Code)
	}
	if myProxy.lc == nil {
		t.Fatal("myproxy should have been called")
	}
}

type mycloser struct {
	io.Reader
}

func (mycloser) Close() error {
	return nil
}
