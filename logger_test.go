package logger

import (
	"fmt"
	"os"
	"testing"
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

func testLogs(l *Logger, n int) (err error) {
	for i := 0; i < n; i++ {
		log := fmt.Sprintf("#%d", i+1)
		if err = l.LogString(log); err != nil {
			return
		}

	}

	return l.Close()
}
