package logger

import (
	"fmt"
	"os"
	"testing"
	"time"
)

const (
	testDir  = "test_data"
	testName = "testing"
)

func TestLogger(t *testing.T) {
	var (
		l   *Logger
		err error
	)

	if err = os.MkdirAll(testDir, 0744); err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(testDir)

	if l, err = New(testDir, testName); err != nil {
		t.Fatal(err)
	}

	l.SetNumLines(1)

	if err = testLogs(l, 3); err != nil {
		t.Fatal(err)
	}
}

func TestTimeRotation(t *testing.T) {
	var (
		l *Logger
		v *Viewer

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

	l.SetRotateInterval(time.Second)

	for i := 0; i < 6; i++ {
		log := fmt.Sprintf("#%d", i+1)
		if err = l.LogString(log); err != nil {
			return
		}
		time.Sleep(time.Millisecond * 500)
	}

	if err = l.Close(); err != nil {
		t.Fatal(err)
	}

	if v, err = NewViewer(testDir, testName); err != nil {
		t.Fatal(err)
	}

	if err = v.ForEach(func(key string) (err error) {
		var r *Reader
		if r, err = NewReader(key); err != nil {
			return
		}
		defer r.Close()

		if err = r.ForEach(0, func(ts time.Time, log []byte) (err error) {
			lineCount++
			expected := fmt.Sprintf("#%d", lineCount)
			if strLog := string(log); strLog != expected {
				return fmt.Errorf("invalid log, expected \"%s\" and received \"%s\"", expected, strLog)
			}

			return
		}); err != nil {
			return
		}

		logCount++
		return
	}); err != nil {
		t.Fatal(err)
	}

	if logCount != 3 {
		t.Fatalf("invalid number of logs, expected %d and received %d", 3, logCount)
	}
}

func testLogs(l *Logger, n int) (err error) {
	for i := 0; i < n; i++ {
		log := fmt.Sprintf("#%d", i+1)
		if err = l.LogString(log); err != nil {
			return
		}

	}

	return l.Close()
}
