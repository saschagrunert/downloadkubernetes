package backend

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/chuckha/downloadkubernetes/events"
)

const (
	cookieName       = "downloadkubernetes"
	cookieExpiryDays = 30
)

// Baker will give us cookies.
type Baker interface {
	NewCookieForRequest(*http.Request) *http.Cookie
}

// Logger are all the logger functions the server needs
type Logger interface {
	Info(string)
	Infof(string, ...interface{})
	Error(error)
}

type EventHandler interface {
	Handle(interface{}) error
}

type Proxy interface {
	WriteCopyEvent(*events.LinkCopy)
	WriteUserIDEvent(*events.UserID)
}

type RecentGetter interface {
	Recents(string) []string
}

// The server itself is an http.Server and then some.
type Server struct {
	*http.Server
	Baker        Baker
	dev          bool
	Log          Logger
	Proxy        Proxy
	RecentGetter RecentGetter
}

type Option func(*Server)

func NewServer(options ...Option) *Server {
	s := &Server{
		Server: &http.Server{
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
		},
	}
	for _, option := range options {
		option(s)
	}
	return s
}

func WithDev(dev bool) Option {
	return func(s *Server) {
		s.dev = dev
	}
}
func WithLogger(l Logger) Option {
	return func(s *Server) {
		s.Log = l
	}
}
func WithBaker(b Baker) Option {
	return func(s *Server) {
		s.Baker = b
	}
}
func WithListenAddress(host, port string) Option {
	return func(s *Server) {
		s.Server.Addr = fmt.Sprintf("%s:%s", host, port)
	}
}
func WithMux(mux *http.ServeMux) Option {
	return func(s *Server) {
		s.Server.Handler = mux
	}
}
func WithProxy(p Proxy) Option {
	return func(s *Server) {
		s.Proxy = p
	}
}
func WithRecentGetter(r RecentGetter) Option {
	return func(s *Server) {
		s.RecentGetter = r
	}
}

// Recent returns the 5 most recent downloads a user has made
func (s *Server) Recent(w http.ResponseWriter, r *http.Request, c *http.Cookie) {
	recentURLs := s.RecentGetter.Recents(c.Value)
	s.Log.Infof("Recent URLS: %v", recentURLs)

	out, err := json.Marshal(recentURLs)
	if err != nil {
		s.Log.Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	fmt.Println(string(out))
	w.Header().Add("Content-Type", "application/json")
	if _, err := w.Write(out); err != nil {
		s.Log.Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

// Cookie sets the cookie if there is none or refreshes the cookie
func (s *Server) Cookie(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie(cookieName)
	if err != nil {
		if err == http.ErrNoCookie {
			cookie := s.Baker.NewCookieForRequest(r)
			go s.Proxy.WriteUserIDEvent(events.NewUserID(cookie.Value, events.Created))
			http.SetCookie(w, cookie)
			return
		}
		s.Log.Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	c.Expires = time.Now().Add(cookieExpiryDays * 24 * time.Hour)
	http.SetCookie(w, c)
}

type copyLinkRequest struct {
	URL string
}

// CopyLinkEvent is the endpoint that saves an instance of a user downloading something
func (s *Server) CopyLinkEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	c, err := r.Cookie(cookieName)
	if err != nil {
		if err != http.ErrNoCookie {
			s.Log.Error(err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
	}
	userID := ""
	if c != nil {
		userID = c.Value
	}

	// read the request body and deserialize
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.Log.Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	clr := &copyLinkRequest{}
	if err := json.Unmarshal(body, clr); err != nil {
		s.Log.Error(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// validate potential user input
	if _, err := url.Parse(clr.URL); err != nil {
		s.Log.Error(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	go s.Proxy.WriteCopyEvent(events.NewLinkCopyEvent(userID, clr.URL))
}

// responseWriter exists to let us track the status that gets written to a
// response. Used for logging.
type responseWriter struct {
	http.ResponseWriter
	status int
}

// WriteHeader wraps the http.ResponseWriter.WriteHeader function.
func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

// LogRequest wraps handlers and logs them
func (s *Server) LogRequest(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rw := &responseWriter{w, 200}
		fn(rw, r)
		s.Log.Infof("[%d] %s %s %s", rw.status, r.Method, r.URL.String(), w.Header().Get("Set-Cookie"))
	}
}

// EnableCORS sets CORS headers if the server is running in dev mode.
func (s *Server) EnableCORS(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.dev {
			w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3333")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		fn(w, r)
	}
}

type cookieRequiredHandler func(w http.ResponseWriter, r *http.Request, cookie *http.Cookie)

// CookieRequired wraps a cookieRequiredHandler.
// It return a 403 if the endpoint is accessed without a cookie.
func (s *Server) CookieRequired(fn cookieRequiredHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// get the cookie
		c, err := r.Cookie(cookieName)
		if err != nil {
			if err == http.ErrNoCookie {
				s.Log.Info("there is no cookie in the request")
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}
			s.Log.Error(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		fn(w, r, c)
	}
}
