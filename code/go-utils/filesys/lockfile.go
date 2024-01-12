package filesys

type Lockfile struct {
}

// Returns a new Lockfile if it was possible to create it, otherwise nil.
// Calling Unlock() on the lockfile deletes it.

func NewLockfile(lockdir, pattern string) *Lockfile {
	// TODO: This needs to be an atomic create-and-open
	// Once we have it, write something to it.
	return &Lockfile{}
}

func (l *Lockfile) Unlock() {
	// TODO: delete the file
}
