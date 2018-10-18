package logger

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestReader(t *testing.T) {
	var (
		l *Logger
		v *Viewer
		r *Reader

		logKey string

		logCount  int
		lineCount int

		err error
	)

	if err = os.MkdirAll(testDir, 0744); err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(testDir)

	if l, err = New(testDir, testName); err != nil {
		t.Fatal(err)
	}

	l.SetNumLines(5)

	if err = testLogs(l, 5); err != nil {
		t.Fatal(err)
	}

	if v, err = NewViewer(testDir, testName); err != nil {
		t.Fatal(err)
	}

	if err = v.ForEach(func(key string) (err error) {
		logKey = key
		logCount++
		return
	}); err != nil {
		t.Fatal(err)
	}

	if logCount != 1 {
		t.Fatalf("invalid number of logs, expected %d and received %d", 1, logCount)
	}

	if r, err = NewReader(logKey); err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	if err = r.ForEach(0, func(ts time.Time, log []byte) (err error) {
		lineCount++
		expected := fmt.Sprintf("#%d", lineCount)
		if strLog := string(log); strLog != expected {
			return fmt.Errorf("invalid log, expected \"%s\" and received \"%s\" ", expected, strLog)
		}
		return
	}); err != nil {
		t.Fatal(err)
	}

	if lineCount != 5 {
		t.Fatalf("invalid line count, expected %d and received %d", 5, lineCount)
	}

}
