package events

import (
	"github.com/chuckha/downloadkubernetes/models"
)

type Logger interface {
	Info(string)
	Infof(string, ...interface{})
	Error(error)
}

type Store interface {
	SaveDownload(*models.Download) error
	SaveUserID(*models.UserID) error
}

type SaveCopyLink interface {
	SaveCopyLinkEvent(*LinkCopy) error
}
type SaveLinkCopyHandler struct {
	Log   Logger
	Store SaveCopyLink
}

func (s *SaveLinkCopyHandler) Handle(l *LinkCopy) {
	s.Log.Infof("HELLO WORLD????????? %v", l)
	if err := s.Store.SaveCopyLinkEvent(l); err != nil {
		s.Log.Error(err)
	}
}
func (h *SaveLinkCopyHandler) ID() string {
	return "save-link-copy-handler"
}

type SaveUserIDEvent interface {
	SaveUserIDEvent(*UserID) error
}
type SaveUserIDCreateHandler struct {
	Log   Logger
	Store SaveUserIDEvent
}

func (s *SaveUserIDCreateHandler) Handle(u *UserID) {
	// Filter out non create actions
	if u.Action != Created {
		return
	}
	s.Log.Infof("%#v", u)
	if err := s.Store.SaveUserIDEvent(u); err != nil {
		s.Log.Error(err)
	}
}
func (h *SaveUserIDCreateHandler) ID() string {
	return "save-user-id-event-handler"
}
