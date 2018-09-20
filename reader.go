package logger

import (
	"bufio"
	"os"
	"sync"
	"time"

	"github.com/Hatch1fy/errors"
)

// NewReader will return a new reader
func NewReader(filename string) (rp *Reader, err error) {
	var r Reader
	if r.f, err = os.Open(filename); err != nil {
		return
	}

	rp = &r
	return
}

// Reader will read log files
type Reader struct {
	mu sync.Mutex

	f *os.File
}

// ForEach will iterate through each log line
func (r *Reader) ForEach(fn func(ts time.Time, log []byte) error) (err error) {
	// Acquire reader lock
	r.mu.Lock()
	// Defer the release of the reader lock
	defer r.mu.Unlock()

	// If our file is nil, this reader has been closed
	if r.f == nil {
		// Reader is closed, return
		return errors.ErrIsClosed
	}

	// Ensure we are looking at the beginning of the file,
	// this will ensure this is safe for re-use
	if _, err = r.f.Seek(0, 0); err != nil {
		return
	}

	// Create a new scanner
	s := bufio.NewScanner(r.f)

	for s.Scan() {
		var (
			ts  time.Time
			log []byte
		)

		// Parse timestamp and log bytes from line
		if ts, log, err = parseLine(s.Bytes()); err != nil {
			return
		}

		// Pass timestamp and log bytes to provided iterating func
		fn(ts, log)
	}

	return
}

// Close will close a reader
func (r *Reader) Close() (err error) {
	// Acquire reader lock
	r.mu.Lock()
	// Defer the release of the reader lock
	defer r.mu.Unlock()

	// If our file is nil, this reader has already been closed
	if r.f == nil {
		// Reader is closed, return
		return errors.ErrIsClosed
	}

	// Close underlying file
	if err = r.f.Close(); err != nil {
		return
	}

	// Set file reference to nil
	r.f = nil
	return
}
