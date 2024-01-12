package main

import (
	_ "fmt"

	"go-utils/sysinfo"
)

// Possibly the ProcFSAPI should not be forced into the sysinfo package but defined here instead?
// Although, by being in the sysinfo package, that package can be tested?

func getJobs() []*jobObject {
	//fs := sysinfo.NewProcFS()

	pidsAndUids, err := sysinfo.EnumeratePids()
	_ = err

	um := sysinfo.NewUserMap()
	_ = um

	for _, x := range pidsAndUids {
		_ = x
	}

	return make([]*jobObject, 0)
}

/*
type ProcFS struct {
}

func NewProcFS() *ProcFS {
	return &ProcFS{}
}

func (_ *ProcFS) Open(name string) (io.ReadCloser, error) {
	return os.Open(name)
}


// For testing purposes we want to be able to virtualize /proc (to have stable testing data).  So
// all access to /proc goes via this shim.

type ProcFSAPI interface {
	Open(filename string) (io.ReadCloser, error)
	UserByUid(uid uint) (userName string, found bool)
	EnumeratePids() ([]PidAndUid, error)
	Getpagesize() uint
}
*/

/*
// MockFS allows an Open function to be registered with each individual filename.  The filename
// should be the full path in /proc.

type MockFS struct {
	handlers map[string]func(string) (io.ReadCloser, error)
	procs    []PidAndUid
}

func (m *MockFS) AddFile(fn string, handler func(string) (io.ReadCloser, error)) {
	m.handlers[fn] = handler
}

func (m *MocFS) AddProc(pid, uid uint) {
}

func (m *MockFS) Open(name string) (io.ReadCloser, error) {
	if handler, found := m.handlers[name]; found {
		return handler(name)
	}
	return nil, fmt.Errorf("File not found: %s", name)
}

// MockFile returns a handler function (an `Open`) that will deliver a buffer that delivers the data
// content.

type byteFile struct {
	*bytes.Reader
}

func (b *byteFile) Close() error {
	return nil
}

func MockFile(data []byte) func(fn string) (io.ReadCloser, error) {
	return func(_ string) (io.ReadCloser, error) {
		return &byteFile{bytes.NewReader(data)}, nil
	}
}
*/
