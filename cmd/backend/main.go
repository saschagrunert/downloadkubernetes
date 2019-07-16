package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/chuckha/downloadkubernetes/bakers"
	"github.com/chuckha/downloadkubernetes/logging"
	"github.com/chuckha/downloadkubernetes/models"
	"github.com/chuckha/downloadkubernetes/models/stores/sqlite3"
)

const (
	cookieName       = "downloadkubernetes"
	cookieExpiryDays = 30
)

// serverConfig holds the command line arguments to set various options on the server.
type serverConfig struct {
	addr string
	port string
}

func main() {
	args := &serverConfig{}
	fs := flag.NewFlagSet("config", flag.ExitOnError)
	fs.StringVar(&args.port, "port", "9999", "The port to listen on")
	fs.StringVar(&args.addr, "address", "127.0.0.1", "TCP address to listen on")

	mymux := http.NewServeMux()
	s := &Server{
		&http.Server{
			Addr:    fmt.Sprintf("%s:%s", args.addr, args.port),
			Handler: mymux,
		},
		&sqlite3.Store{},
		bakers.NewDumbBaker(10, 5),
		args.dev,
		&logging.Log{},
	}

	mymux.HandleFunc("/cookie", s.LogRequest(s.EnableCORS(s.Cookie)))
	mymux.HandleFunc("/save-download", s.LogRequest(s.EnableCORS(s.SaveDownload)))
	mymux.HandleFunc("/recent-downloads", s.LogRequest(s.EnableCORS(s.Recent)))
	fmt.Println("Listening on", s.Addr)
	panic(s.ListenAndServe())
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

// Store are the functions used on the store object.
// This is entirely for interacting with some storage backend.
type Store interface {
	SaveDownload(*models.Download) error
}

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

// The server itself is an http.Server and then some.
type Server struct {
	*http.Server
	Store Store
	Baker Baker
	dev   bool
	Log   Logger
}

// Recent returns the 5 most recent downloads a user has made
func (s *Server) Recent(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie(cookieName)
	if err != nil {
		if err == http.ErrNoCookie {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// make sure the cookie is not expired
	if c.Expires.Before(time.Now()) {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

}

// Cookie sets the cookie if there is none or refreshes the cookie
func (s *Server) Cookie(w http.ResponseWriter, r *http.Request) {
	// get the cookie
	c, err := r.Cookie(cookieName)
	if err != nil {
		if err == http.ErrNoCookie {
			http.SetCookie(w, s.Baker.NewCookieForRequest(r))
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if c.Expires.Before(time.Now()) {
		c.Expires = time.Now().Add(cookieExpiryDays * 24 * time.Hour)
		http.SetCookie(w, c)
		return
	}
}

// SaveDownload is the endpoint that saves an instance of a user downloading something
func (s *Server) SaveDownload(w http.ResponseWriter, r *http.Request) {
	// get the cookie
	c, err := r.Cookie(cookieName)
	if err != nil {
		if err == http.ErrNoCookie {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// make sure the cookie is not expired
	if c.Expires.Before(time.Now()) {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	// read the request body and deserialize to a Download object
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	download := &models.Download{}
	if err := json.Unmarshal(body, download); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	download.User = c.Value
	if err := s.Store.SaveDownload(download); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

// LogRequest wraps handlers and logs them.
func (s *Server) LogRequest(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rw := &responseWriter{w, 200}
		fn(rw, r)
		s.Log.Infof("[%d] %s %s %s\n", rw.status, r.Method, r.URL.String(), w.Header().Get("Set-Cookie"))
	}
}

// EnableCORS will eneable cors if the server is running in dev mode.
func (s *Server) EnableCORS(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.dev {
			w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3333")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		fn(w, r)
	}
}
