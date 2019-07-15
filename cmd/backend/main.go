package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/chuckha/downloadkubernetes/bakers"
	"github.com/chuckha/downloadkubernetes/models"
	"github.com/chuckha/downloadkubernetes/models/stores/sqlite3"
)

const (
	cookieName       = "downloadkubernetes"
	cookieExpiryDays = 30
)

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
	}

	mymux.HandleFunc("/cookie", s.Cookie)
	mymux.HandleFunc("/save-download", s.SaveDownload)
	mymux.HandleFunc("/recent-downloads", s.Recent)
	fmt.Println("Listening on", s.Addr)
	panic(s.ListenAndServe())
}

// Features
// * Recently downloaded
// * needs a place to store data
// What data? user, date/time, filterset, binary, version

type Store interface {
	SaveDownload(*models.Download) error
}
type Baker interface {
	NewCookieForRequest(*http.Request) *http.Cookie
}

type Server struct {
	*http.Server
	Store Store
	Baker Baker
}

// a button that says "remember me"
// once clicked it changes to "forget me"
// a cookie is issued
// activity is tracked
// once clicked it changes to "remember me"
// cookie is logged as inactive

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

// http server
//   all functions
//      get request
//      extract user from cookie, ensure valid
//      return data associated with user

// What data? Recently downloaded.
//

// endpoints
// recently downloaded (3 unique)
// saved filterset
