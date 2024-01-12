// Simple HTTP client with a retry queue.
//
// NOTE, HttpClient is not thread-safe: you can't keep posting data on one thread and processing
// resends on another thread, for example.  This is fixable.
//
// NOTE, this is being tested as part of the `exfiltrate` client, see ../../tests/transport

package httpclient

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"go-utils/status"
)

// Create a client with NewClient().  Post data to it using PostDataByHttp().  Process resends by
// calling ProcessResends(), which will block until all resends are done.  See note above about
// thread safety.

type HttpClient struct {
	target                         *url.URL
	authUser, authPass             string
	maxAttempts, resendIntervalMin uint
	verbose                        bool
	retries                        []retry
}

type retry struct {
	prevAttempts uint // number of attempts that have been performed
	path         string
	buf          []byte // the content
}

// Here, `authUser` can be empty; if not, `authPass` must also be provided and the pair are used for
// HTTP basic authentication.  `maxAttempts` should be 0 if no retries are desired.  Otherwise,
// `resendIntervalMin` is the resend interval in minutes.

func NewClient(
	target *url.URL,
	authUser, authPass string,
	maxAttempts, resendIntervalMin uint,
	verbose bool,
) *HttpClient {
	return &HttpClient{
		target:            target,
		authUser:          authUser,
		authPass:          authPass,
		maxAttempts:       maxAttempts,
		resendIntervalMin: resendIntervalMin,
		verbose:           verbose,
		retries:           make([]retry, 0),
	}
}

// The "path" is appended to the URL and should start with a "/".
//
// If the URL is "http://host:port/service" and path is "/path" then the full URL will be
// "http://host:port/service/path".

func (c *HttpClient) PostDataByHttp(path string, buf []byte) {
	c.postDataByHttp(0, path, buf)
}

// Try sending pending packets.  This blocks until the retry queue is empty (all packets have been
// sent or have timed out).

func (c *HttpClient) ProcessRetries() {
	for len(c.retries) > 0 {
		time.Sleep(time.Duration(c.resendIntervalMin) * time.Minute)
		rs := c.retries
		c.retries = make([]retry, 0)
		for _, r := range rs {
			c.postDataByHttp(r.prevAttempts, r.path, r.buf)
		}
	}
}

func (c *HttpClient) postDataByHttp(prevAttempts uint, path string, buf []byte) {
	if c.verbose {
		fmt.Printf("Trying to send %s\n", string(buf))
	}

	// Go down a level from http.Post() in order to be able to set authentication header.
	req, err := http.NewRequest("POST", c.target.String()+path, bytes.NewReader(buf))
	if err != nil {
		status.Infof("Failed to post: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	if c.authUser != "" {
		req.SetBasicAuth(c.authUser, c.authPass)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// There doesn't seem to be any good way to determine that a host is currently unreachable
		// vs all sorts of other errors that can happen along the way.  So when a sending error
		// occurs, always retry.
		if prevAttempts+1 <= c.maxAttempts {
			c.addRetry(prevAttempts+1, path, buf)
		} else {
			status.Infof("Failed to post to %s after max retries: %v", c.target, err)
		}
		return
	}

	if c.verbose {
		fmt.Printf("Response %s\n", resp.Status)
	}

	// Codes in the 200 range indicate everything is OK, for now.
	// Really we should expect
	//  202 (StatusAccepted) for when a new record is created
	//  208 (StatusAlreadyReported) for when the record is a dup
	//
	// TODO: Possibly for codes in the 500 range we should retry?
	if resp.StatusCode >= 300 {
		status.Infof("Failed to post: HTTP status=%d", resp.StatusCode)
		// Fall through: must read response body
	}

	// API requires that we read and close the body
	_, _ = io.ReadAll(resp.Body)
	resp.Body.Close()
}

func (c *HttpClient) addRetry(prevAttempts uint, path string, buf []byte) {
	c.retries = append(c.retries, retry{prevAttempts, path, buf})
}
