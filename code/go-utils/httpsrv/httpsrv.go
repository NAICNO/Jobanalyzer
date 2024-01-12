// This package contains code for a simple HTTP server (built on existing Go library code) and some
// utilities for that.
//
// NOTE, this is being tested as part of the `infiltrate` server, see ../../tests/transport

package httpsrv

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go-utils/status"
)

const (
	serverShutdownTimeoutSec = 10
)

type Server struct {
	verbose bool
	port    int
	failed  func(error)
	stop    chan bool
	server  *http.Server
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

// Start the server.  This blocks the current goroutine until the server exits, so typical usage
// would be `go s.Start()`.  To force the server to shut down, call s.Stop().  When the server
// exits, it will call s.failed if there was an error.

func (s *Server) Start() {
	if s.verbose {
		status.Info(fmt.Sprintf("Listening on port %d", s.port))
	}
	s.server = &http.Server{Addr: fmt.Sprintf(":%d", s.port)}
	err := s.server.ListenAndServe()
	if err != nil {
		if err != http.ErrServerClosed {
			status.Error(err.Error())
			status.Error("SERVER NOT RUNNING")
			s.failed(err)
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
	if err := s.server.Shutdown(ctx); err != nil {
		status.Warning(err.Error())
	}
	<-s.stop
}
