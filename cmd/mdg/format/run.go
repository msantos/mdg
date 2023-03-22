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

type State struct {
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

	s := &State{
		md:      markdown.New(),
		diff:    *diff,
		verbose: *verbose,
	}

	if err := s.run(dir); err != nil {
		log.Fatalln(err)
	}
}

func (s *State) run(dir string) error {
	if dir == "-" {
		return s.stdin()
	}
	return filepath.WalkDir(dir, s.format)
}

func (s *State) stdin() error {
	p, err := io.ReadAll(os.Stdin)
	if err != nil {
		return err
	}
	if s.diff {
		var buf bytes.Buffer

		if err := s.md.Format(p, &buf); err != nil {
			return err
		}

		if bytes.Equal(p, buf.Bytes()) {
			return nil
		}

		d := gitdiff.CompareBytes(
			p, "stdin",
			buf.Bytes(), "stdin (formatted)",
		)
		fmt.Println(string(d.ToCombinedFormat()))
		return nil
	}
	return s.md.Format(p, os.Stdout)
}

func (s *State) format(file string, d fs.DirEntry, err error) error {
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

	p, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	var buf bytes.Buffer

	if err := s.md.Format(p, &buf); err != nil {
		return err
	}

	if bytes.Equal(p, buf.Bytes()) {
		return nil
	}

	if s.diff {
		d := gitdiff.CompareBytes(
			p, file,
			buf.Bytes(), file+" (formatted)",
		)

		fmt.Println(string(d.ToCombinedFormat()))
		return nil
	}

	log.Println("Formatting:", file)

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
