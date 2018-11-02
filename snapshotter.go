package logger

import (
	"io"
	"os"
	"path/filepath"

	"github.com/Hatch1fy/snapshotter"
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
	// Backend to snapshot to
	be snapshotter.Backend
	// Whether or not to delete a file after it's been successfully
	deleteOnSnapshot bool
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
