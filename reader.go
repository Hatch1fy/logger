package logger

import (
	"bufio"
	"os"
	"sync"
	"time"

	"github.com/hatchify/errors"
)

// NewReader will return a new reader
func NewReader(filename string) (rp *Reader, err error) {
	var f *os.File
	if f, err = os.Open(filename); err != nil {
		return
	}

	rp = newReader(f)
	return
}

// newReader will return a new reader
func newReader(f *os.File) (rp *Reader) {
	var r Reader
	r.f = f
	return &r
}

// Reader will read log files
type Reader struct {
	mu sync.Mutex

	f *os.File
}

func (r *Reader) forEach(offset int64, fn Handler) (err error) {
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

	var cnt int64
	for s.Scan() {
		var (
			ts  time.Time
			log []byte
		)

		// Parse timestamp and log bytes from line
		if ts, log, err = parseLine(s.Bytes()); err != nil {
			return
		}

		if cnt++; cnt <= offset {
			continue
		}

		// Pass timestamp and log bytes to provided iterating func
		if err = fn(ts, log); err != nil {
			break
		}
	}

	return
}

// ForEach will iterate through each log line
func (r *Reader) ForEach(offset int64, fn Handler) (err error) {
	// Acquire reader lock
	r.mu.Lock()
	// Defer the release of the reader lock
	defer r.mu.Unlock()

	if err = r.forEach(offset, fn); err == Break {
		err = nil
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
