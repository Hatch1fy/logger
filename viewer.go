package logger

import (
	"os"
	"path/filepath"
	"strings"
)

// NewViewer will return a new Viewer
func NewViewer(dir, name string) (vp *Viewer, err error) {
	var v Viewer
	v.dir = dir
	v.name = name
	vp = &v
	return
}

// Viewer will view the files for a particular directory and name
type Viewer struct {
	dir  string
	name string
}

// ForEach will iterate through all the available logs
func (v *Viewer) ForEach(fn func(key string) (err error)) (err error) {
	// Walk through each file in the set directory
	err = filepath.Walk(v.dir, func(filepath string, info os.FileInfo, ierr error) (err error) {
		// If we're looking at a directory, we're definitely not looking at a log file
		if info.IsDir() {
			// This is a directory, return
			return
		}

		// Ensure Iterating log is a target log
		if strings.Index(filepath, v.name) == -1 {
			// Not a target log, return
			return
		}

		// Pass filepath provided iterating function
		return fn(filepath)
	})

	return
}
