package backend

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/chuckha/downloadkubernetes/events"
	"github.com/chuckha/downloadkubernetes/models"
)

type handlerStore interface {
	SaveUser(*models.User) error
	SaveCopyLinkEvent(*events.LinkCopy) error
	SaveUserIDEvent(*events.UserID) error
	ExpireUser(string) error
}

type storehandlerLog interface {
	Debugf(string, ...interface{})
}

type StoreHandler struct {
	Store handlerStore
	Log   storehandlerLog
}

func (s *StoreHandler) ID() string {
	return "store-handler"
}

func (s *StoreHandler) HandleCopyLinkEvent(l *events.LinkCopy) error {
	return errors.WithStack(s.Store.SaveCopyLinkEvent(l))
}

func (s *StoreHandler) HandleUserIDEvent(u *events.UserID) error {
	switch u.Action {
	case events.Expired:
		s.Log.Debugf("expiring user %q", u.User.ID)
		return errors.WithStack(s.Store.ExpireUser(u.User.ID))
	case events.Created:
		if err := s.Store.SaveUser(u.User); err != nil {
			return errors.WithStack(err)
		}
		return errors.WithStack(s.Store.SaveUserIDEvent(u))
	default:
		return fmt.Errorf("Unknown event type")
	}
}
