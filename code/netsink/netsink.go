// Listen on a port, write the output to the syslog.
// Used for testing network service.

package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"syscall"

	"go-utils/httpsrv"
	"go-utils/process"
	"go-utils/status"
)

var port = flag.Int("p", 8088, "Listen on `port`")

func main() {
	status.Start("jobanalyzer/netsink")
	flag.Parse()
	http.HandleFunc("/netsink", func(w http.ResponseWriter, r *http.Request) {
		payload := make([]byte, r.ContentLength)
		_, err := io.ReadFull(r.Body, payload)
		if err != nil {
			w.WriteHeader(400)
			fmt.Fprintf(w, "Could not read content up to ContentLength")
			return
		}
		status.Info(string(payload))
		// Will return 200 OK
	})
	s := httpsrv.New(false, *port, nil)
	go s.Start()
	defer s.Stop()
	process.WaitForSignal(syscall.SIGHUP, syscall.SIGTERM)
}
