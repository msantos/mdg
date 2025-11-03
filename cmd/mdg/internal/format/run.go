package format

import (
	"bytes"
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"go.iscode.ca/mdg/internal/pkg/fdpair"
	"go.iscode.ca/mdg/pkg/config"
	"go.iscode.ca/mdg/pkg/markdown"

	"github.com/bwplotka/mdox/pkg/gitdiff"
)

type Opt struct {
	diff      bool
	verbose   bool
	md        *markdown.Opt
	isChanged func(_, _ []byte) bool
}

func usage() {
	fmt.Fprintf(os.Stderr, `%s %s
Usage: %s format [<option>] [-|<path>]

Format markdown documents.

`, path.Base(os.Args[0]), config.Version(), os.Args[0])
	fmt.Fprintf(os.Stderr, "Options:\n\n")
	flag.PrintDefaults()
}

func Run() {
	diff := flag.Bool("diff", false, "Display formatting changes as diff")
	verbose := flag.Bool("verbose", false, "Enable debug messages")
	noLineWrap := flag.Bool("no-linewrap", false, "Disable wrapping of long lines")

	flag.Usage = func() { usage() }

	flag.Parse()

	args := []string{"-"}
	if flag.NArg() > 0 {
		args = flag.Args()
	}

	o := &Opt{
		md:        markdown.New(markdown.WithLineWrap(!*noLineWrap)),
		diff:      *diff,
		verbose:   *verbose,
		isChanged: func(_, _ []byte) bool { return true },
	}

	for _, v := range args {
		if err := o.run(v); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}

func (o *Opt) run(dir string) error {
	if dir == "-" {
		return o.format(&stdio{
			r:   os.Stdin,
			Opt: o,
		})
	}

	o.isChanged = func(in, out []byte) bool {
		return !bytes.Equal(in, out)
	}

	return filepath.WalkDir(dir, o.walkdir)
}

func (o *Opt) format(rw fdpair.FD) error {
	b, err := io.ReadAll(rw.In())
	if err != nil {
		return err
	}

	var formatted bytes.Buffer

	unformatted := bytes.NewBuffer(b)

	var in string

	if f, ok := rw.In().(*os.File); ok {
		in = f.Name()
	}

	if err := o.md.Format(unformatted, &formatted); err != nil {
		return fmt.Errorf("%s: %w", in, err)
	}

	if o.diff {
		d := gitdiff.CompareBytes(
			unformatted.Bytes(), in,
			formatted.Bytes(), fmt.Sprintf("%s (formatted)", in),
		)

		fmt.Println(string(d.ToCombinedFormat()))

		return nil
	}

	if !o.isChanged(formatted.Bytes(), unformatted.Bytes()) {
		return nil
	}

	if err := rw.Open(); err != nil {
		return fmt.Errorf("%s: %w", in, err)
	}

	var out string

	if f, ok := rw.Out().(*os.File); ok {
		out = f.Name()
	}

	defer func() {
		err = errors.Join(err, rw.Close())
	}()

	if _, err := rw.Out().Write(formatted.Bytes()); err != nil {
		return fmt.Errorf("%s: %w", out, err)
	}

	return nil
}

func (o *Opt) walkdir(file string, d fs.DirEntry, err error) error {
	if err != nil {
		return err
	}

	if d.Type() != 0 {
		return nil
	}

	if strings.HasPrefix(file, ".") || strings.HasPrefix(file, "_") {
		return nil
	}

	switch filepath.Ext(file) {
	case ".md", ".markdown":
	default:
		return nil
	}

	r, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("%s: %w", file, err)
	}

	defer func() {
		err = errors.Join(err, r.Close())
	}()

	rw := &fsobj{
		r:   r,
		Opt: o,
	}

	if err := o.format(rw); err != nil {
		return fmt.Errorf("%s: %w", file, err)
	}

	return nil
}

type stdio struct {
	*Opt

	r *os.File
	w *os.File
}

func (rw *stdio) Open() error {
	rw.w = os.Stdout
	return nil
}

func (rw *stdio) Close() error {
	return nil
}

func (rw *stdio) In() io.Reader {
	return rw.r
}

func (rw *stdio) Out() io.Writer {
	return rw.w
}

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
