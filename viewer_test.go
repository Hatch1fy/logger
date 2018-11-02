package logger

import (
	"os"
	"testing"
)

func TestViewer(t *testing.T) {
	var (
		l *Logger
		v *Viewer

		logCount int

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

	if err = testLogs(l, 5); err != nil {
		t.Fatal(err)
	}

	if v, err = NewViewer(testDir, testName); err != nil {
		t.Fatal(err)
	}

	if err = v.ForEach(func(key string) (err error) {
		logCount++
		return
	}); err != nil {
		t.Fatal(err)
	}

	if logCount != 5 {
		t.Fatalf("invalid number of logs, expected %d and received %d", 5, logCount)
	}
}
