package format

import (
	"bytes"
	_ "embed"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"git.iscode.ca/msantos/mdg/internal/pkg/fdpair"
	"git.iscode.ca/msantos/mdg/pkg/config"
	"git.iscode.ca/msantos/mdg/pkg/markdown"

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

	flag.Usage = func() { usage() }

	flag.Parse()

	args := []string{"-"}
	if flag.NArg() > 0 {
		args = flag.Args()
	}

	o := &Opt{
		md:        markdown.New(),
		diff:      *diff,
		verbose:   *verbose,
		isChanged: func(_, _ []byte) bool { return true },
	}

	for _, v := range args {
		if err := o.run(v); err != nil {
			log.Fatalln(err)
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
		log.Println("Formatting:", r.Name())
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
			log.Println(w.Name(), err)
		}
	}()

	err = os.Rename(w.Name(), r.Name())
	if err != nil {
		return fmt.Errorf("%s: %w", w.Name(), err)
	}

	if err := w.Close(); err != nil {
		log.Println(w.Name(), err)
	}

	return nil
}

func (o *Opt) format(r *fdpair.Opt) error {
	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	var buf bytes.Buffer

	if err := o.md.Format(bytes.NewBuffer(b), &buf); err != nil {
		return fmt.Errorf("%s: %w", r.Name(), err)
	}

	if o.diff {
		d := gitdiff.CompareBytes(
			b, r.Name(),
			buf.Bytes(), fmt.Sprintf("%s (formatted)", r.Name()),
		)

		fmt.Println(string(d.ToCombinedFormat()))

		return nil
	}

	if !o.isChanged(b, buf.Bytes()) {
		return nil
	}

	w, err := r.OpenOutput(r.File)
	if err != nil {
		return fmt.Errorf("%s: %w", r.Name(), err)
	}

	if _, err := w.Write(buf.Bytes()); err != nil {
		return fmt.Errorf("%s: %w", w.Name(), err)
	}

	if err := r.CloseOutput(r.File, w); err != nil {
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
			log.Println(r.Name(), err)
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
