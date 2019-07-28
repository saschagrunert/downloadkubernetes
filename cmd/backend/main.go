package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

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
	addr  string
	port  string
	dev   bool
	debug bool
}

func main() {
	args := &serverConfig{}
	fs := flag.NewFlagSet("config", flag.ExitOnError)
	fs.StringVar(&args.port, "port", "9999", "The port to listen on")
	fs.StringVar(&args.addr, "address", "127.0.0.1", "TCP address to listen on")
	fs.BoolVar(&args.dev, "dev", false, "enable the development server")
	fs.BoolVar(&args.debug, "debug", false, "show debug logs")
	fs.Parse(os.Args[1:])

	httpLogger := logging.NewLog("http-logger", args.debug)
	proxyLogger := logging.NewLog("proxy", args.debug)
	storeLogger := logging.NewLog("store", args.debug)
	cacheLogger := logging.NewLog("cache", args.debug)
	storeHandlerLogger := logging.NewLog("store-handler", args.debug)

	db, err := sqlite.NewStore(dbname, storeLogger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		return
	}

	c, err := backend.NewCache(db, cacheLogger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
	}
	saver := &backend.StoreHandler{db, storeHandlerLogger}

	p := events.NewProxy(proxyLogger)
	// Register handlers

	// handler for the cache to watch for copy events
	p.RegisterCopyEventListener(c)

	// handler for saver to write copy events to disk
	p.RegisterCopyEventListener(saver)

	// handler for saver to write user ids to disk
	p.RegisterUserIDEventListeners(saver)

	// handler for cache to expire user ids
	p.RegisterUserIDEventListeners(c)

	go p.StartListeners()

	mymux := http.NewServeMux()

	s := backend.NewServer(
		backend.WithListenAddress(args.addr, args.port),
		backend.WithMux(mymux),
		backend.WithIdentifier(bakers.NewDumbIdentifier(30, 5, time.Now().Unix())),
		backend.WithDev(args.dev),
		backend.WithLogger(httpLogger),
		backend.WithProxy(p),
		backend.WithRecentGetter(c),
	)

	// TODO: These belong here.
	mymux.HandleFunc("/cookie", s.LogRequest(s.EnableCORS(s.Cookie)))
	mymux.HandleFunc("/link-copied", s.LogRequest(s.EnableCORS(s.CopyLinkEvent)))
	mymux.HandleFunc("/recent-downloads", s.LogRequest(s.EnableCORS(s.CookieRequired(s.Recent))))
	mymux.HandleFunc("/forget", s.LogRequest(s.EnableCORS(s.CookieRequired(s.Forget))))
	fmt.Println("Listening on", s.Addr)
	mode := "prod"
	if args.dev {
		mode = "dev"
	}
	fmt.Printf("Running in %s mode\n", mode)
	panic(s.ListenAndServe())
}
