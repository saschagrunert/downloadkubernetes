package events

import (
	"github.com/chuckha/downloadkubernetes/models"
	"github.com/pkg/errors"
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

func (s *SaveLinkCopyHandler) HandleCopyLinkEvent(l *LinkCopy) error {
	return errors.WithStack(s.Store.SaveCopyLinkEvent(l))
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

func (s *SaveUserIDCreateHandler) HandleUserIDEvent(u *UserID) error {
	// Filter out non create actions
	if u.Action != Created {
		return nil
	}
	return errors.WithStack(s.Store.SaveUserIDEvent(u))
}
func (h *SaveUserIDCreateHandler) ID() string {
	return "save-user-id-event-handler"
}
