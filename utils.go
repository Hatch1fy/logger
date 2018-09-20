package logger

import (
	"strconv"
	"time"
)

// getTimestamp will get a unix timestamp as a string
func getTimestamp() (ts string) {
	// Current unix timestamp
	now := time.Now().UnixNano()
	// Format timestamp to a string and return
	return strconv.FormatInt(now, 10)
}

// getTimestampBytes will get a unix timestamp as a byteslice
func getTimestampBytes() (ts []byte) {
	// Convert string timestamp to byteslice and return
	return []byte(getTimestamp())
}
