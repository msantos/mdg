package format

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type fsobj struct {
	*Opt

	r *os.File
	w *os.File
}

func (rw *fsobj) Open() error {
	if rw.verbose {
		fmt.Fprintln(os.Stderr, "Formatting:", rw.r.Name())
	}

	w, err := os.CreateTemp(filepath.Dir(rw.r.Name()), filepath.Base(rw.r.Name()))
	if err != nil {
		return fmt.Errorf("%s: %w", rw.r.Name(), err)
	}

	rw.w = w

	return nil
}

func (rw *fsobj) Close() error {
	err := rw.w.Sync()
	if err != nil {
		return fmt.Errorf("%s: %w", rw.w.Name(), err)
	}

	defer func() {
		err = errors.Join(err, os.Remove(rw.w.Name()))
	}()

	err = os.Rename(rw.w.Name(), rw.r.Name())
	if err != nil {
		return fmt.Errorf("%s: %w", rw.w.Name(), err)
	}

	if err := rw.w.Close(); err != nil {
		return fmt.Errorf("%s: %w", rw.w.Name(), err)
	}

	return nil
}

func (rw *fsobj) In() io.Reader {
	return rw.r
}

func (rw *fsobj) Out() io.Writer {
	return rw.w
}
