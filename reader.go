package logger

import (
	"bufio"
	"bytes"
	"os"
	"strconv"
	"time"

	"github.com/Hatch1fy/errors"
)

// NewReader will return a new reader
func NewReader(filename string) (rp *Reader, err error) {
	var r Reader
	if r.f, err = os.Open(filename); err != nil {
		return
	}

	r.s = bufio.NewScanner(r.f)
	rp = &r
	return
}

// Reader will read log files
type Reader struct {
	f *os.File
	s *bufio.Scanner
}

// ForEach will iterate through each log line
func (r *Reader) ForEach(fn func(ts time.Time, log []byte) error) (err error) {
	// Ensure we are looking at the beginning of the file (this will ensure this is safe for re-use)
	if _, err = r.f.Seek(0, 0); err != nil {
		return
	}

	s := bufio.NewScanner(r.f)

	for s.Scan() {
		lineBytes := s.Bytes()
		separator := bytes.IndexByte(lineBytes, '@')
		tsBytes := lineBytes[:separator]
		logBytes := lineBytes[separator+1:]

		var ts int64
		if ts, err = strconv.ParseInt(string(tsBytes), 10, 64); err != nil {
			return
		}

		fn(time.Unix(0, ts), logBytes)
	}

	return
}

// Close will close a reader
func (r *Reader) Close() (err error) {
	if r.f == nil {
		return errors.ErrIsClosed
	}

	if err = r.f.Close(); err != nil {
		return
	}

	r.f = nil
	return
}
