package events

import "fmt"

// object that watches all event types
// can register another object?

type CopyEventListener interface {
	HandleCopyLinkEvent(*LinkCopy) error
	ID() string
}
type UserIDEventListener interface {
	HandleUserIDEvent(*UserID) error
	ID() string
}

type proxyLogger interface {
	Debugf(string, ...interface{})
	Error(error)
}

// TODO Rename to broker
type Proxy struct {
	Log                  proxyLogger
	CopyEvents           chan *LinkCopy
	CopyEventListeners   map[string]CopyEventListener
	UserIDEvents         chan *UserID
	UserIDEventListeners map[string]UserIDEventListener
}

func NewProxy(log proxyLogger) *Proxy {
	return &Proxy{
		Log:                  log,
		CopyEvents:           make(chan *LinkCopy),
		CopyEventListeners:   map[string]CopyEventListener{},
		UserIDEvents:         make(chan *UserID),
		UserIDEventListeners: map[string]UserIDEventListener{},
	}
}

func (p *Proxy) StartListeners() {
	for {
		select {
		case copyEvent := <-p.CopyEvents:
			for _, listener := range p.CopyEventListeners {
				if err := listener.HandleCopyLinkEvent(copyEvent); err != nil {
					fmt.Println("I Bet you will never see this.")
					p.Log.Error(err)
				}
			}
		case userIDEvent := <-p.UserIDEvents:
			for _, listener := range p.UserIDEventListeners {
				if err := listener.HandleUserIDEvent(userIDEvent); err != nil {
					p.Log.Error(err)
				}
			}
		}
	}
}

func (p *Proxy) WriteCopyEvent(copy *LinkCopy) {
	p.Log.Debugf("writing a copy event for user: %q", copy.UserID)
	p.CopyEvents <- copy
}

func (p *Proxy) WriteUserIDEvent(userIDEvent *UserID) {
	p.Log.Debugf("writing a user event for user: %q", userIDEvent.User.ID)
	p.UserIDEvents <- userIDEvent
}

func (p *Proxy) RegisterCopyEventListener(listener CopyEventListener) {
	p.Log.Debugf("registering a new copy event listener: %q", listener.ID())
	p.CopyEventListeners[listener.ID()] = listener
}

func (p *Proxy) RegisterUserIDEventListeners(listener UserIDEventListener) {
	p.Log.Debugf("registering a new user id event listener: %q", listener.ID())
	p.UserIDEventListeners[listener.ID()] = listener
}
