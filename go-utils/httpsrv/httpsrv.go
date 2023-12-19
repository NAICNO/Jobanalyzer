// A simple HTTP server.

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

func New(verbose bool, port int, failed func(error)) *Server {
	return &Server{
		verbose: verbose,
		port:    port,
		failed:  failed,
		stop:    make(chan bool),
	}
}

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

func (s *Server) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), serverShutdownTimeoutSec*time.Second)
	defer cancel()
	if err := s.server.Shutdown(ctx); err != nil {
		status.Warning(err.Error())
	}
	<-s.stop
}
