package logger

import (
	"bytes"
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

// parseLine will parse a log line and return it's timestamp and log bytes
func parseLine(lineBytes []byte) (ts time.Time, log []byte, err error) {
	separator := bytes.IndexByte(lineBytes, '@')
	tsStr := string(lineBytes[:separator])

	var tsInt int64
	if tsInt, err = strconv.ParseInt(tsStr, 10, 64); err != nil {
		return
	}

	ts = time.Unix(0, tsInt)
	log = lineBytes[separator+1:]
	return
}

// RotateFn is called during rotations
type RotateFn func(filename string)
