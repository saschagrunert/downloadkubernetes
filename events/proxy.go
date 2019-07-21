package events

// object that watches all event types
// can register another object?

type CopyEventListener interface {
	Handle(*LinkCopy) error
	ID() string
}
type UserIDEventListener interface {
	Handle(*UserID) error
	ID() string
}

type Proxy struct {
	Log                  Logger
	CopyEvents           chan *LinkCopy
	CopyEventListeners   map[string]CopyEventListener
	UserIDEvents         chan *UserID
	UserIDEventListeners map[string]UserIDEventListener
}

func NewProxy(log Logger) *Proxy {
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
				if err := listener.Handle(copyEvent); err != nil {
					p.Log.Error(err)
				}
			}
		case userIDEvent := <-p.UserIDEvents:
			for _, listener := range p.UserIDEventListeners {
				if err := listener.Handle(userIDEvent); err != nil {
					p.Log.Error(err)
				}
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
