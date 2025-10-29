package format

import (
	"bytes"
	_ "embed"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"codeberg.org/msantos/mdg/internal/pkg/fdpair"
	"codeberg.org/msantos/mdg/pkg/config"
	"codeberg.org/msantos/mdg/pkg/markdown"

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
		return o.format(fdpair.New())
	}

	o.isChanged = func(in, out []byte) bool {
		return !bytes.Equal(in, out)
	}

	return filepath.WalkDir(dir, o.walkdir)
}

func (o *Opt) openOutput(r *os.File) (*os.File, error) {
	if o.verbose {
		fmt.Fprintln(os.Stderr, "Formatting:", r.Name())
	}

	w, err := os.CreateTemp(filepath.Dir(r.Name()), filepath.Base(r.Name()))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", r.Name(), err)
	}

	return w, nil
}

func (o *Opt) closeOutput(r, w *os.File) error {
	err := w.Sync()
	if err != nil {
		return fmt.Errorf("%s: %w", w.Name(), err)
	}

	defer func() {
		if err == nil {
			return
		}

		if err := os.Remove(w.Name()); err != nil {
			fmt.Fprintln(os.Stderr, w.Name(), err)
		}
	}()

	err = os.Rename(w.Name(), r.Name())
	if err != nil {
		return fmt.Errorf("%s: %w", w.Name(), err)
	}

	if err := w.Close(); err != nil {
		fmt.Fprintln(os.Stderr, w.Name(), err)
	}

	return nil
}

func (o *Opt) format(r *fdpair.Opt) error {
	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	var formatted bytes.Buffer

	unformatted := bytes.NewBuffer(b)

	if err := o.md.Format(unformatted, &formatted); err != nil {
		return fmt.Errorf("%s: %w", r.Name(), err)
	}

	if o.diff {
		d := gitdiff.CompareBytes(
			unformatted.Bytes(), r.Name(),
			formatted.Bytes(), fmt.Sprintf("%s (formatted)", r.Name()),
		)

		fmt.Println(string(d.ToCombinedFormat()))

		return nil
	}

	if !o.isChanged(formatted.Bytes(), unformatted.Bytes()) {
		return nil
	}

	w, err := r.OpenOutput(r.File)
	if err != nil {
		return fmt.Errorf("%s: %w", r.Name(), err)
	}

	defer func() {
		if rerr := r.CloseOutput(r.File, w); rerr != nil {
			if err == nil {
				err = fmt.Errorf("%s: %w", w.Name(), rerr)
			}
		}
	}()

	if _, err := w.Write(formatted.Bytes()); err != nil {
		return fmt.Errorf("%s: %w", w.Name(), err)
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
		if err := r.Close(); err != nil {
			fmt.Fprintln(os.Stderr, r.Name(), err)
		}
	}()

	fd := fdpair.New()

	fd.File = r
	fd.OpenOutput = o.openOutput
	fd.CloseOutput = o.closeOutput

	if err := o.format(fd); err != nil {
		return fmt.Errorf("%s: %w", file, err)
	}

	return nil
}
