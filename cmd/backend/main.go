package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/chuckha/downloadkubernetes/bakers"
	"github.com/chuckha/downloadkubernetes/logging"
	"github.com/chuckha/downloadkubernetes/models"
	"github.com/chuckha/downloadkubernetes/models/stores/sqlite"
)

const (
	cookieName       = "downloadkubernetes"
	cookieExpiryDays = 30
	dbname           = "downloadkubernetes"
)

// serverConfig holds the command line arguments to set various options on the server.
type serverConfig struct {
	addr string
	port string
	dev  bool
}

func main() {
	args := &serverConfig{}
	fs := flag.NewFlagSet("config", flag.ExitOnError)
	fs.StringVar(&args.port, "port", "9999", "The port to listen on")
	fs.StringVar(&args.addr, "address", "127.0.0.1", "TCP address to listen on")
	fs.BoolVar(&args.dev, "dev", false, "enable the development server")
	fs.Parse(os.Args[1:])

	mymux := http.NewServeMux()
	db, err := sqlite.NewStore(dbname)
	if err != nil {
		fmt.Println(err)
		return
	}
	s := &Server{
		&http.Server{
			Addr:    fmt.Sprintf("%s:%s", args.addr, args.port),
			Handler: mymux,
		},
		db,
		bakers.NewDumbBaker(10, 5),
		args.dev,
		&logging.Log{},
	}

	mymux.HandleFunc("/cookie", s.LogRequest(s.EnableCORS(s.Cookie)))
	mymux.HandleFunc("/save-download", s.LogRequest(s.EnableCORS(s.CookieRequired(s.SaveDownload))))
	mymux.HandleFunc("/recent-downloads", s.LogRequest(s.EnableCORS(s.CookieRequired(s.Recent))))
	fmt.Println("Listening on", s.Addr)
	mode := "prod"
	if args.dev {
		mode = "dev"
	}
	fmt.Printf("Running in %s mode\n", mode)
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
	GetRecentDownloads(*models.UserID) ([]*models.Download, error)
	SaveDownload(*models.Download) error
	SaveUserID(*models.UserID) error
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
func (s *Server) Recent(w http.ResponseWriter, r *http.Request, c *http.Cookie) {
	downloads, err := s.Store.GetRecentDownloads(&models.UserID{
		ID: c.Value,
	})
	if err != nil {
		s.Log.Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	out, err := json.Marshal(downloads)
	if err != nil {
		s.Log.Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
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
			userID := &models.UserID{
				ID:         cookie.Value,
				CreateTime: time.Now(),
				ExpireTime: cookie.Expires,
			}
			s.Store.SaveUserID(userID)
			http.SetCookie(w, cookie)
			return
		}
		s.Log.Error(err)
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
func (s *Server) SaveDownload(w http.ResponseWriter, r *http.Request, c *http.Cookie) {
	// read the request body and deserialize to a Download object
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.Log.Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	download := &models.Download{}
	if err := json.Unmarshal(body, download); err != nil {
		s.Log.Error(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	download.User = c.Value
	if err := s.Store.SaveDownload(download); err != nil {
		s.Log.Error(err)
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

type cookieRequiredHandler func(w http.ResponseWriter, r *http.Request, cookie *http.Cookie)

func (s *Server) CookieRequired(fn cookieRequiredHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// get the cookie
		c, err := r.Cookie(cookieName)
		if err != nil {
			if err == http.ErrNoCookie {
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}
			s.Log.Error(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// make sure the cookie is not expired
		if c.Expires.Before(time.Now()) {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		fn(w, r, c)
	}
}
