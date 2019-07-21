package events

// object that watches all event types
// can register another object?

type CopyEventListener interface {
	Handle(*LinkCopy) // what should handle th errors?
	ID() string
}
type UserIDEventListener interface {
	Handle(*UserID) // what should handle the errors?
	ID() string
}

type Proxy struct {
	CopyEvents           chan *LinkCopy
	CopyEventListeners   map[string]CopyEventListener
	UserIDEvents         chan *UserID
	UserIDEventListeners map[string]UserIDEventListener
}

func NewProxy() *Proxy {
	return &Proxy{
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
				listener.Handle(copyEvent)
			}
		case userIDEvent := <-p.UserIDEvents:
			for _, listener := range p.UserIDEventListeners {
				listener.Handle(userIDEvent)
			}
		}
	}
}

func (p *Proxy) WriteCopyEvent(copy *LinkCopy) {
	p.CopyEvents <- copy
}
func (p *Proxy) WriteUserIDEvent(userIDEvent *UserID) {
	p.UserIDEvents <- userIDEvent
}

func (p *Proxy) RegisterCopyEventListener(listener CopyEventListener) {
	p.CopyEventListeners[listener.ID()] = listener
}

func (p *Proxy) RegisterUserIDEventListeners(listener UserIDEventListener) {
	p.UserIDEventListeners[listener.ID()] = listener
}
