// The AppendStore is an abstraction over a cluster's data directory that manages and optimizes the
// appending of records to a set of files within the store.  The records may be for various hosts
// within the cluster and for various dates, this allows a proxy on a cluster to cache and forward
// data records for multiple nodes within the cluster, for example.
//
// This is tied to the LogStore obviously but given the present structure of sonalyze (it's a
// one-operation-per-invocation program) they are two different entities.  If sonalyze were a daemon
// or db management process, it would be very different.

package sonarlog

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path"
	"time"
)

const (
	dirPermissions  = 0755
	filePermissions = 0644
	newline         = 10
)

var (
	BadTimestampErr = errors.New("Bad timestamp")
)

type writer struct {
	file *os.File
	buf  *bufio.Writer
}

type AppendStore struct {
	dataDir string
	writers map[string]writer
}

func OpenDirForAppend(dir string) (*AppendStore, error) {
	// Currently infallible but I think it would be sensible to check that the directory exists.
	return &AppendStore{
		dataDir: dir,
		writers: make(map[string]writer),
	}, nil

}

func (as *AppendStore) Close() error {
	return as.flush()
}

func (as *AppendStore) flush() error {
	var err error
	for _, w := range as.writers {
		err = errors.Join(err, w.buf.Flush())
		err = errors.Join(err, w.file.Close())
	}
	// TODO: In Go 1.21, we can use clear()
	as.writers = make(map[string]writer)
	return err
}

// Append the data to the file (adding termination or other formatting as necessary by the file
// format).
//
// This caches the open file and its buffered writer, because the normal situation is that there
// will be many records per node and date when a batch of records arrives.
//
// The `format` must be a format string with exactly one %s parameter, it should produce a full file
// name with extension from a node name.
//
// When BadTimestampErr is returned, the record can in principle be dropped silently.  All other
// errors are basically I/O errors.

func (as *AppendStore) Write(host, timestamp, format string, payload []byte) error {
	if len(payload) == 0 {
		return nil
	}

	w, err := as.getWriter(host, timestamp, format)
	if err != nil {
		return err
	}

	_, err = w.buf.Write(payload)
	if err != nil {
		return fmt.Errorf("Failed to append to file (%v)", err)
	}

	if payload[len(payload)-1] != newline {
		err := w.buf.WriteByte(newline)
		if err != nil {
			return fmt.Errorf("Failed to append to file (%v)", err)
		}
	}

	return nil
}

func (as *AppendStore) WriteString(host, timestamp, format string, payload string) error {
	if len(payload) == 0 {
		return nil
	}

	w, err := as.getWriter(host, timestamp, format)
	if err != nil {
		return err
	}

	_, err = w.buf.WriteString(payload)
	if err != nil {
		return fmt.Errorf("Failed to append to file (%v)", err)
	}

	if payload[len(payload)-1] != newline {
		err := w.buf.WriteByte(newline)
		if err != nil {
			return fmt.Errorf("Failed to append to file (%v)", err)
		}
	}

	return nil
}

func (as *AppendStore) getWriter(host, timestamp, format string) (writer, error) {
	// The path will be (below as.dataDir) yyyy/mm/dd/$FILENAME
	tval, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return writer{}, BadTimestampErr
	}

	// Mostly, dirname (indeed the timestamp) will be the same for many records in the same bundle.
	// We may want to optimize that by caching the timestamp and avoiding the parsing and mkdir.
	dirname := fmt.Sprintf("%04d/%02d/%02d", tval.Year(), tval.Month(), tval.Day())
	err = os.MkdirAll(path.Join(as.dataDir, dirname), dirPermissions)
	if err != nil {
		return writer{}, fmt.Errorf("Failed to create path (%v)", err)
	}

	filename := path.Join(as.dataDir, dirname, fmt.Sprintf(format, host))
	if probe, ok := as.writers[filename]; ok {
		return probe, nil
	}

	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, filePermissions)
	if err != nil {
		// Could be disk full, fs went away, file is directory, wrong permissions
		//
		// Could also be too many open files, in which case we really want to close all open
		// files and retry.
		return writer{}, fmt.Errorf("Failed to open/create file (%v)", err)
	}
	w := writer{file: f, buf: bufio.NewWriter(f)}
	as.writers[filename] = w
	return w, nil
}
