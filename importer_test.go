package logger

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Hatch1fy/errors"
	"github.com/Hatch1fy/snapshotter/backends"
)

const errTestFull = errors.Error("disk is full")

func TestImporterBreak(t *testing.T) {
	var (
		ss    *Snapshotter
		log   *Logger
		imp   *Importer
		lines int64
		err   error
	)

	if err = initTestImporterDirs(); err != nil {
		return
	}
	defer removeTestImporterDirs()

	be := backends.NewFile("./test_back_end")
	ss = NewSnapshotter(be, true)
	breakHandler := func(ts time.Time, line []byte) (err error) {
		if lines++; lines == 12 {
			return Break
		}

		return
	}

	if imp, err = NewImporter("./test_data", "tester", be, time.Hour, breakHandler); err != nil {
		t.Fatal(err)
	}

	if log, err = New("./test_logs", "tester"); err != nil {
		t.Fatal(err)
	}

	log.SetNumLines(5)
	log.SetRotateFn(func(filename string) {
		if err = ss.Snapshot(filename); err != nil {
			t.Fatal(err)
		}
	})

	if err = testPopulateLogs(log, 15); err != nil {
		t.Fatal(err)
	}

	if err = imp.Import(); err != nil {
		t.Fatal(err)
	}

	if err = imp.Import(); err != nil {
		t.Fatal(err)
	}

	if err = imp.Import(); err != nil {
		t.Fatal(err)
	}

	if err = imp.Close(); err != nil {
		return
	}

	if imp, err = NewImporter("./test_data", "tester", be, time.Hour, breakHandler); err != nil {
		t.Fatal(err)
	}
	defer imp.Close()

	if err = imp.Import(); err != nil {
		t.Fatal(err)
	}

	// Our number of lines should match because we ended our loop with a break instead of an error
	if lines != 15 {
		t.Fatalf("Invalid number of lines, expected %d and received %d", 15, lines)
	}

	return
}

func TestImporterError(t *testing.T) {
	var (
		ss    *Snapshotter
		log   *Logger
		imp   *Importer
		lines int64
		err   error
	)

	if err = initTestImporterDirs(); err != nil {
		return
	}
	defer removeTestImporterDirs()

	be := backends.NewFile("./test_back_end")
	ss = NewSnapshotter(be, true)
	errorHandler := func(ts time.Time, line []byte) (err error) {
		if lines++; lines == 12 {
			return errTestFull
		}

		return
	}

	if imp, err = NewImporter("./test_data", "tester", be, time.Hour, errorHandler); err != nil {
		t.Fatal(err)
	}

	if log, err = New("./test_logs", "tester"); err != nil {
		t.Fatal(err)
	}

	log.SetNumLines(5)
	log.SetRotateFn(func(filename string) {
		if err = ss.Snapshot(filename); err != nil {
			t.Fatal(err)
		}
	})

	if err = testPopulateLogs(log, 15); err != nil {
		t.Fatal(err)
	}

	if err = imp.Import(); err != nil {
		t.Fatal(err)
	}

	if err = imp.Import(); err != nil {
		t.Fatal(err)
	}

	if err = imp.Import(); err != errTestFull {
		t.Fatalf("invalid error, expected \"%v\" and received \"%v\"", errTestFull, err)
	}

	if err = imp.Close(); err != nil {
		return
	}

	if imp, err = NewImporter("./test_data", "tester", be, time.Hour, errorHandler); err != nil {
		t.Fatal(err)
	}
	defer imp.Close()

	if err = imp.Import(); err != nil {
		t.Fatal(err)
	}

	// Our number of lines should be one higher than our true number of lines due to
	// the error iteration (which causes that line to be replayed next time)
	if lines != 16 {
		t.Fatalf("Invalid number of lines, expected %d and received %d", 16, lines)
	}

	return
}

func testPopulateLogs(l *Logger, count int) (err error) {
	for i := 0; i < count; i++ {
		msg := fmt.Sprintf("#%d", i+1)
		if err = l.Log([]byte(msg)); err != nil {
			return
		}
	}

	return
}

func initTestImporterDirs() (err error) {
	if err = os.MkdirAll("./test_back_end", 0744); err != nil {
		return
	}

	if err = os.MkdirAll("./test_data", 0744); err != nil {
		return
	}

	if err = os.MkdirAll("./test_logs", 0744); err != nil {
		return
	}

	return
}

func removeTestImporterDirs() {
	os.RemoveAll("./test_data")
	os.RemoveAll("./test_back_end")
	os.RemoveAll("./test_logs")
	return
}
