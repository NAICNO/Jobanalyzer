// This package contains code for a simple HTTP/HTTPS server (built on existing Go library code) and
// some utilities for that.
//
// NOTE, this is being tested as part of the `infiltrate` server, see ../../tests/transport

package httpsrv

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"go-utils/status"
)

const (
	serverShutdownTimeoutSec = 10
)

// This is not thread-safe.  The server will panic if started more than once, and Start and Stop
// should be called from the same goroutine.

type Server struct {
	verbose bool
	port    int
	failed  func(error)
	stop    chan bool
	server  atomic.Pointer[http.Server]
	tlsKey  string
	tlsCert string
}

// Create a server that will be listening on `port`.  It will call `failed` if the server returns a
// failure code.  The server is not started by this.

func New(verbose bool, port int, failed func(error)) *Server {
	return &Server{
		verbose: verbose,
		port:    port,
		failed:  failed,
		stop:    make(chan bool),
	}
}

// Ditto, but with TLS.

func NewTLS(verbose bool, port int, tlsKey, tlsCert string, failed func(error)) *Server {
	return &Server{
		verbose: verbose,
		port:    port,
		failed:  failed,
		tlsKey:  tlsKey,
		tlsCert: tlsCert,
		stop:    make(chan bool),
	}
}

// Start the server.  This blocks the current goroutine until the server exits, so typical usage
// would be `go s.Start()`.  To force the server to shut down, call s.Stop().  When the server
// exits, it will call s.failed if there was an error.

func (s *Server) Start() {
	if s.verbose {
		status.Info(fmt.Sprintf("Listening on port %d", s.port))
	}
	var err error
	if s.tlsKey != "" {
		var hn string
		hn, err = os.Hostname()
		if err == nil {
			if !s.server.CompareAndSwap(nil, &http.Server{Addr: fmt.Sprintf("%s:%d", hn, s.port)}) {
				panic("Start server only once")
			}
			err = s.server.Load().ListenAndServeTLS(s.tlsCert, s.tlsKey)
		}
	} else {
		if !s.server.CompareAndSwap(nil, &http.Server{Addr: fmt.Sprintf(":%d", s.port)}) {
			panic("Start server only once")
		}
		err = s.server.Load().ListenAndServe()
	}
	if err != nil {
		if err != http.ErrServerClosed {
			status.Error(err.Error())
			status.Error("SERVER NOT RUNNING")
			if s.failed != nil {
				s.failed(err)
			}
		} else {
			status.Info(err.Error())
		}
	}
	s.stop <- true
}

// Cause the server to shut down and stop.

func (s *Server) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), serverShutdownTimeoutSec*time.Second)
	defer cancel()
	if err := s.server.Load().Shutdown(ctx); err != nil {
		status.Warning(err.Error())
	}
	<-s.stop
}
