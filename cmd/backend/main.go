package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/chuckha/downloadkubernetes/backend"
	"github.com/chuckha/downloadkubernetes/bakers"
	"github.com/chuckha/downloadkubernetes/events"
	"github.com/chuckha/downloadkubernetes/logging"
	"github.com/chuckha/downloadkubernetes/stores/sqlite"
)

const (
	dbname = "downloadkubernetes"
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

	db, err := sqlite.NewStore(dbname)
	if err != nil {
		fmt.Printf("%+v\n", err)
		return
	}
	httpLogger := logging.NewLog("http-logger")
	eventLogger := logging.NewLog("event-logger")
	proxyLogger := logging.NewLog("proxy")
	// TODO understand fmt.Prinltn from a goroutine
	p := events.NewProxy(proxyLogger)

	c := backend.NewCache()
	saveLinkCopyHandler := &events.SaveLinkCopyHandler{
		Log:   eventLogger,
		Store: db,
	}
	saveUserIDCreateHandler := &events.SaveUserIDCreateHandler{
		Log:   eventLogger,
		Store: db,
	}
	// Register handlers
	// handler for the cache to watch for copy events
	p.RegisterCopyEventListener(c)
	// handler for saver to write copy events to disk
	p.RegisterCopyEventListener(saveLinkCopyHandler)
	// handler for saver to write user ids to disk
	p.RegisterUserIDEventListeners(saveUserIDCreateHandler)
	// handler for cache to expire user ids
	p.RegisterUserIDEventListeners(c)
	go p.StartListeners()

	mymux := http.NewServeMux()

	s := backend.NewServer(
		backend.WithListenAddress(args.addr, args.port),
		backend.WithMux(mymux),
		backend.WithStore(db),
		backend.WithBaker(bakers.NewDumbBaker(10, 5)),
		backend.WithDev(args.dev),
		backend.WithLogger(httpLogger),
		backend.WithProxy(p),
		backend.WithRecentGetter(c),
	)

	// TODO: These belong here.
	mymux.HandleFunc("/cookie", s.LogRequest(s.EnableCORS(s.Cookie)))
	mymux.HandleFunc("/link-copied", s.LogRequest(s.EnableCORS(s.CopyLinkEvent)))
	mymux.HandleFunc("/recent-downloads", s.LogRequest(s.EnableCORS(s.CookieRequired(s.Recent))))
	fmt.Println("Listening on", s.Addr)
	mode := "prod"
	if args.dev {
		mode = "dev"
	}
	fmt.Printf("Running in %s mode\n", mode)
	panic(s.ListenAndServe())
}
