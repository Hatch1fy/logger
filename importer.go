package logger

import (
	"io"
	"os"
	"path"
	"sync"
	"time"

	"github.com/gdbu/snapshotter"
	"github.com/hatchify/cron"
	"github.com/hatchify/errors"

	"github.com/hatchify/atoms"
	"github.com/hatchify/mum"
	"github.com/hatchify/scribe"
)

// NewImporter will return a new importer
func NewImporter(dir, name string, be snapshotter.Backend, importInterval time.Duration, fn Handler) (ip *Importer, err error) {
	var i Importer
	i.be = be
	i.out = scribe.New("Importer :: " + name)
	i.dir = dir
	i.name = name

	i.h = fn

	if err = i.init(); err != nil {
		return
	}

	cron.New(i.importJob).Every(importInterval)
	ip = &i
	return
}

// Importer manages the importing process
type Importer struct {
	mu sync.RWMutex
	// Data backend
	be snapshotter.Backend
	// Output logger
	out *scribe.Scribe

	// Target directory
	dir string
	// Name of service
	name string
	// Last loaded file
	last string

	h Handler

	// Key used for the current file
	currentKey string

	current    *os.File
	currentRdr *Reader

	marker    *os.File
	markerEnc *mum.Encoder

	// Closed state
	closed atoms.Bool
}

func (i *Importer) init() (err error) {
	currentKey := path.Join(i.dir, i.name+".current")
	if i.current, err = os.OpenFile(currentKey, os.O_CREATE|os.O_RDWR, 0644); err != nil {
		return
	}

	i.currentRdr = newReader(i.current)

	markerKey := path.Join(i.dir, i.name+".marker")
	if i.marker, err = os.OpenFile(markerKey, os.O_CREATE|os.O_RDWR, 0644); err != nil {
		return
	}

	i.markerEnc = mum.NewEncoder(i.marker)

	var stat os.FileInfo
	if stat, err = i.marker.Stat(); err != nil {
		return
	}

	if stat.Size() == 0 {
		return
	}

	return
}

func (i *Importer) importJob() {
	err := i.Import()
	switch err {
	case nil:
	case io.EOF:

	default:
		i.out.Errorf("Error importing: %v", err)
	}
}

func (i *Importer) importKey(key string) (err error) {
	if err = i.save(key); err != nil {
		return
	}

	return i.importCurrent(key, 0)
}

func (i *Importer) importCurrent(key string, lines int64) (err error) {
	var count int64
	r := newReader(i.current)
	// Iterate through each log line
	err = r.forEach(lines, func(ts time.Time, line []byte) (err error) {
		// Call handler with timestamp and line bytes
		if err = i.h(ts, line); err != nil {
			return
		}
		// Increment line count
		count++
		// Set marker file
		return i.setMarker(key, count)
	})

	switch err {
	case nil:
	case Break:
		return i.setMarker(key, count+1)

	default:
		return
	}

	// Set marker as complete
	if err = i.setMarker(key, -1); err != nil {
		return
	}

	i.last = key
	return
}

func (i *Importer) getMarker() (key string, lines int64, err error) {
	if _, err = i.marker.Seek(0, 0); err != nil {
		return
	}

	dec := mum.NewDecoder(i.marker)
	if key, err = dec.String(); err != nil {
		return
	}

	if lines, err = dec.Int64(); err != nil {
		return
	}

	return
}

func (i *Importer) setMarker(key string, lines int64) (err error) {
	if _, err = i.marker.Seek(0, 0); err != nil {
		return
	}

	if err = i.marker.Truncate(0); err != nil {
		return
	}

	if err = i.markerEnc.String(key); err != nil {
		return
	}

	if err = i.markerEnc.Int64(lines); err != nil {
		return
	}

	return
}

func (i *Importer) save(key string) (err error) {
	if _, err = i.current.Seek(0, 0); err != nil {
		return
	}

	if err = i.current.Truncate(0); err != nil {
		return
	}

	if err = i.be.ReadFrom(key, func(r io.Reader) (err error) {
		_, err = io.Copy(i.current, r)
		return
	}); err != nil {
		return
	}

	if _, err = i.current.Seek(0, 0); err != nil {
		return
	}

	return i.setMarker(key, 0)
}

// Import will import
func (i *Importer) Import() (err error) {
	i.mu.Lock()
	defer i.mu.Unlock()
	var (
		key   string
		lines int64
	)

	key, lines, err = i.getMarker()
	switch err {
	case io.EOF:
	case nil:
		if lines > -1 {
			return i.importCurrent(key, lines)
		}

	default:
		return
	}

	var nextKey string
	// Get the next key
	if nextKey, err = i.be.Next(i.name, i.last); err != nil {
		return
	}

	return i.importKey(nextKey)
}

// Close will attempt to close an instance of importer
func (i *Importer) Close() (err error) {
	// Attempt to set closed state to true
	if !i.closed.Set(true) {
		// Importer has already been closed, return
		return errors.ErrIsClosed
	}

	var errs errors.ErrorList
	errs.Push(i.marker.Close())
	errs.Push(i.current.Close())
	return errs.Err()
}
