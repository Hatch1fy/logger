package logger

import (
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/Hatch1fy/errors"
	"github.com/Hatch1fy/snapshotter"
	"github.com/PathDNA/atoms"
)

// NewSnapshotter will return a new instance of Snapshotter
func NewSnapshotter(be snapshotter.Backend, deleteOnSnapshot bool) (ss *Snapshotter) {
	var s Snapshotter
	s.be = be
	s.deleteOnSnapshot = deleteOnSnapshot
	ss = &s
	return
}

// Snapshotter manages logger snapshots
type Snapshotter struct {
	mu sync.Mutex

	// Backend to snapshot to
	be snapshotter.Backend
	// Whether or not to delete a file after it's been successfully
	deleteOnSnapshot bool

	closed atoms.Bool
}

// snapshot will perform a snapshot of the given file
func (s *Snapshotter) snapshot(filename string) (err error) {
	var f *os.File
	// Open file with given filename
	if f, err = os.Open(filename); err != nil {
		return
	}
	// Defer the closing of the file
	defer f.Close()

	// Get name of file from filepath
	name := filepath.Base(filename)

	// Write to snapshotter backend
	return s.be.WriteTo(name, func(w io.Writer) (err error) {
		_, err = io.Copy(w, f)
		return
	})
}

// Snapshot will perform a snapshot of the given file
func (s *Snapshotter) Snapshot(filename string) (err error) {
	// Acquire snapshot until snapshot has completed
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed.Get() {
		return errors.ErrIsClosed
	}

	// Perform snapshot
	if err = s.snapshot(filename); err != nil {
		return
	}

	if !s.deleteOnSnapshot {
		// Delete on snapshot is set to false, return
		return
	}

	// Remove file
	return os.Remove(filename)
}

// Close will close an instance of snapshotter
func (s *Snapshotter) Close() (err error) {
	if !s.closed.Set(true) {
		return errors.ErrIsClosed
	}

	// Acquire lock to ensure any current snapshot finished
	s.mu.Lock()
	s.mu.Unlock()
	return
}
