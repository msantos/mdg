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

	"git.iscode.ca/msantos/mdg/config"
	"git.iscode.ca/msantos/mdg/markdown"
	"github.com/bwplotka/mdox/pkg/gitdiff"
)

type Opt struct {
	diff    bool
	verbose bool
	md      *markdown.Opt
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

	dir := "."
	if flag.NArg() > 0 {
		dir = flag.Arg(0)
	}

	o := &Opt{
		md:      markdown.New(),
		diff:    *diff,
		verbose: *verbose,
	}

	if err := o.run(dir); err != nil {
		log.Fatalln(err)
	}
}

func (o *Opt) run(dir string) error {
	if dir == "-" {
		return o.stdin()
	}

	return filepath.WalkDir(dir, o.format)
}

func (o *Opt) stdin() error {
	b, err := io.ReadAll(os.Stdin)
	if err != nil {
		return err
	}

	if o.diff {
		var buf bytes.Buffer

		if err := o.md.Format(b, &buf); err != nil {
			return err
		}

		if bytes.Equal(b, buf.Bytes()) {
			return nil
		}

		d := gitdiff.CompareBytes(
			b, "stdin",
			buf.Bytes(), "stdin (formatted)",
		)

		fmt.Println(string(d.ToCombinedFormat()))

		return nil
	}

	return o.md.Format(b, os.Stdout)
}

func (o *Opt) format(file string, d fs.DirEntry, err error) error {
	if err != nil {
		return err
	}

	if d.Type() != 0 {
		return nil
	}

	if strings.HasPrefix(file, ".") || strings.HasPrefix(file, "_") {
		return nil
	}

	if filepath.Ext(file) != ".md" {
		return nil
	}

	b, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	var buf bytes.Buffer

	if err := o.md.Format(b, &buf); err != nil {
		return err
	}

	if bytes.Equal(b, buf.Bytes()) {
		return nil
	}

	if o.diff {
		d := gitdiff.CompareBytes(
			b, file,
			buf.Bytes(), file+" (formatted)",
		)

		fmt.Println(string(d.ToCombinedFormat()))

		return nil
	}

	if o.verbose {
		log.Println("Formatting:", file)
	}

	w, err := os.CreateTemp(filepath.Dir(file), filepath.Base(file))
	if err != nil {
		return err
	}

	defer func() {
		if err := w.Close(); err != nil {
			log.Println(err)
		}

		if err != nil {
			if err := os.Remove(w.Name()); err != nil {
				log.Println(err)
			}
		}
	}()

	if _, err := w.Write(buf.Bytes()); err != nil {
		return err
	}

	if err := w.Sync(); err != nil {
		return err
	}

	return os.Rename(w.Name(), file)
}
