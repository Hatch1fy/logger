package logger

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"sync"
	"time"

	"github.com/Hatch1fy/errors"
	"github.com/PathDNA/atoms"
)

const (
	// ErrMessageContainsNewline is returned when a message contains a newline
	ErrMessageContainsNewline = errors.Error("message contains newline, which is not a valid character")
	// ErrInvalidRotationInterval is returned when a rotation interval is set to zero
	ErrInvalidRotationInterval = errors.Error("rotation interval cannot be zero")

	// Break will break a ForEach loop early and still yield a nil error
	Break = errors.Error("break")
)

const (
	// loggerFlag is the os file flags used for log files
	loggerFlag = os.O_RDWR | os.O_APPEND | os.O_CREATE
)

var (
	// newline as a byteslice
	newline = []byte("\n")
)

// New will return a new instance of Logger
func New(dir, name string) (lp *Logger, err error) {
	var l Logger
	l.dir = dir
	l.name = name

	// Set initial logger file
	if err = l.setFile(); err != nil {
		return
	}

	// Assign lp as a pointer to our created logger
	lp = &l
	return
}

// Logger will manage system logs
type Logger struct {
	mu sync.Mutex
	f  *os.File
	w  *bufio.Writer

	// Log directory
	dir string
	// Log name
	name string

	// Number of lines before rotation (defaults to unlimited)
	numLines int
	// Duration before rotation (defaults to unlimited)
	rotateInterval time.Duration

	onRotate RotateFn

	// Current line count
	count int

	// Closed state
	closed atoms.Bool
}

// isClosed will return the current closed state
// Note: This function is atomic
func (l *Logger) isClosed() (ok bool) {
	return l.closed.Get()
}

// setFile will set the underlying logger file
// Note: This will close the currently opened file
func (l *Logger) setFile() (err error) {
	// Close existing file (if it exists)
	if err = l.closeFile(); err != nil {
		return
	}

	// Open a file with our directory, name, and current timestamp
	if l.f, err = os.OpenFile(l.getFilename(), loggerFlag, 0644); err != nil {
		return
	}

	// Set writer
	l.w = bufio.NewWriter(l.f)
	// Reset count to zero
	l.count = 0
	return
}

// closeFile will close the underlying logger file
// Note: This will flush the buffer and file before closing
func (l *Logger) closeFile() (err error) {
	if l.f == nil {
		// File does not exist - no need to close, return
		return
	}

	// Get current file's name, we need this for post-close actions
	name := l.f.Name()

	// Flush contents
	if err = l.flush(); err != nil {
		return
	}

	// Close file
	if err = l.f.Close(); err != nil {
		return
	}

	// Set file to nil
	l.f = nil
	// Set buffer to nil
	l.w = nil

	if l.count == 0 {
		// File has no contents, remove file within a gorotuine
		go os.Remove(name)
	} else if l.onRotate != nil {
		// File has been rotated & onRotate func is set, call on on rotate func within a gorotuine
		go l.onRotate(name)
	}

	return
}

// flush will flush the contents of the buffer and sync the underlying file
func (l *Logger) flush() (err error) {
	// Flush buffer
	if err = l.w.Flush(); err != nil {
		return
	}

	// Flush file
	return l.f.Sync()
}

// rotationLoop will manage a rotation loop to call rotate on a set interval
func (l *Logger) rotationLoop() {
	var err error
	for {
		// Sleep for rotation interval
		time.Sleep(l.rotateInterval)
		// Attempt to rotate underlying log file
		err = l.rotate()

		switch err {
		case nil:
			// Err is nil, nothing to see here
		case errors.ErrIsClosed:
			// Instance of logger is closed, we can bail out completely
			return

		default:
			// We encountered an unexpected error, print to stdout
			fmt.Printf("logger :: %s :: error rotating file: %v", l.name, err)
		}
	}
}

func (l *Logger) rotate() (err error) {
	// Acquire lock
	l.mu.Lock()
	// Defer the release of our lock
	defer l.mu.Unlock()

	// Ensure the logger has not been closed
	if l.isClosed() {
		// Instance of logger has been closed, return
		return errors.ErrIsClosed
	}

	// Don't set new file if count is zero
	if l.count == 0 {
		return
	}

	// Set a new underlying log file
	return l.setFile()
}

// getFilename will get the current full filename for the log
// Note: This function is time-sensitive (seconds)
func (l *Logger) getFilename() (filename string) {
	// Get current unix timestamp
	now := time.Now().UnixNano()
	// Create a filename by:
	//	- Concatinate directory and name
	//	- Append unix timestamp
	//	- Append log extension
	return fmt.Sprintf("%s.%d.log", path.Join(l.dir, l.name), now)
}

// logMessage will log the full message (prefix, message, suffix)
func (l *Logger) logMessage(msg []byte) (err error) {
	// Write timestamp
	if _, err = l.w.Write(getTimestampBytes()); err != nil {
		return
	}

	// Write '@', which separates timestamp and the message
	if err = l.w.WriteByte('@'); err != nil {
		return
	}

	// Write message
	if _, err = l.w.Write(msg); err != nil {
		return
	}

	// Write newline to follow message
	return l.w.WriteByte('\n')
}

// incrementCount will increment the current line count
// Note: If the line count exceeds the line limit, a new file will be set
func (l *Logger) incrementCount() (err error) {
	// Increment count, then ensure new count does not equal our number of lines limit
	if l.count++; l.numLines == 0 || l.count < l.numLines {
		// Line number limit unset OR count is less than our lines, return
		return
	}

	// Count equals our number of lines limit, set file
	return l.setFile()
}

// Log will log a message
func (l *Logger) Log(msg []byte) (err error) {
	// Ensure the message is valid before acquiring lock
	if bytes.Index(msg, newline) > -1 {
		// Log message contains a newline, return
		return ErrMessageContainsNewline
	}

	// Acquire lock
	l.mu.Lock()
	// Defer the release of our lock
	defer l.mu.Unlock()

	// Ensure the logger has not been closed
	if l.isClosed() {
		// Instance of logger has been closed, return
		return errors.ErrIsClosed
	}

	// Log message
	if err = l.logMessage(msg); err != nil {
		return
	}

	// Increment line count
	return l.incrementCount()
}

// LogString will log a string message
func (l *Logger) LogString(msg string) (err error) {
	// Convert message to bytes and pass to l.Log
	return l.Log([]byte(msg))
}

// LogJSON will log a generic value as a JSON message
func (l *Logger) LogJSON(value interface{}) (err error) {
	var msg []byte
	if msg, err = json.Marshal(value); err != nil {
		return
	}

	// Convert message to bytes and pass to l.Log
	return l.Log(msg)
}

// Flush will manually flush the buffer bytes to disk
// Note: This is not typically needed, only needed in rare and/or debugging situations
func (l *Logger) Flush() (err error) {
	// Acquire lock
	l.mu.Lock()
	// Defer the release of our lock
	defer l.mu.Unlock()

	// Ensure the logger has not been closed
	if l.isClosed() {
		// Instance of logger has been closed, return
		return errors.ErrIsClosed
	}

	// Flush contents
	return l.flush()
}

// SetNumLines will set the maximum number of lines per log file
func (l *Logger) SetNumLines(n int) {
	// Acquire lock
	l.mu.Lock()
	// Defer the release of our lock
	defer l.mu.Unlock()
	// Set line number limit
	l.numLines = n
}

// SetRotateInterval will set the rotation interval timing of a log file
func (l *Logger) SetRotateInterval(duration time.Duration) (err error) {
	var wasUnset bool
	// Ensure duration isn't set to zero
	if duration == 0 {
		// We do not like rotation intervals of zero, return
		err = ErrInvalidRotationInterval
		return
	}

	// Acquire lock
	l.mu.Lock()
	// Defer the release of our lock
	defer l.mu.Unlock()

	// Ensure the logger has not been closed
	if l.isClosed() {
		// Instance of logger has been closed, return
		return errors.ErrIsClosed
	}

	// Set unset to true if rotate interval is currently zero
	wasUnset = l.rotateInterval == 0
	// Set rotate interval to the provided duration
	l.rotateInterval = duration

	if wasUnset {
		// Rotate interval was previously unset, initialize rotation loop
		go l.rotationLoop()
	}

	return
}

// SetRotateFn will set the function to be called on rotations
func (l *Logger) SetRotateFn(fn RotateFn) {
	// Acquire lock
	l.mu.Lock()
	// Defer the release of our lock
	defer l.mu.Unlock()
	// Set onRotation value as provided fn
	l.onRotate = fn
}

// Close will attempt to close an instance of logger
func (l *Logger) Close() (err error) {
	if !l.closed.Set(true) {
		return errors.ErrIsClosed
	}

	// Acquire lock to ensure all writers have completed
	l.mu.Lock()
	// Defer the release of our lock
	defer l.mu.Unlock()

	// Close underlying logger file
	return l.closeFile()
}
