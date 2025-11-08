package convert

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type fsobj struct {
	*Opt

	r *os.File
	w *os.File
}

var ErrSkipMD = errors.New("skip markdown file")

func (rw *fsobj) Open() error {
	html := strings.TrimSuffix(rw.r.Name(), filepath.Ext(rw.r.Name())) + ".html"

	if !rw.compare(rw.r.Name(), html) {
		return ErrSkipMD
	}

	if rw.verbose {
		fmt.Fprintln(os.Stderr, "Converting:", rw.r.Name(), " -> ", html)
	}

	w, err := os.OpenFile(html, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("%s: %w", html, err)
	}

	rw.w = w

	return nil
}

func (rw *fsobj) Close() error {
	return rw.w.Close()
}

func (rw *fsobj) In() io.Reader {
	return rw.r
}

func (rw *fsobj) Out() io.Writer {
	return rw.w
}
