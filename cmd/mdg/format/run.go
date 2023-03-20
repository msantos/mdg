package format

import (
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
)

type State struct {
	verbose bool
	md      *markdown.Opt
}

func usage() {
	fmt.Fprintf(os.Stderr, `%s %s
Usage: %s format [<option>] [-|<path>]

Convert markdown to HTML.

`, path.Base(os.Args[0]), config.Version(), os.Args[0])
	fmt.Fprintf(os.Stderr, "Options:\n\n")
	flag.PrintDefaults()
}

func Run() {
	verbose := flag.Bool("verbose", false, "Enable debug messages")

	flag.Usage = func() { usage() }

	flag.Parse()

	dir := "."
	if flag.NArg() > 0 {
		dir = flag.Arg(0)
	}

	s := &State{
		md:      markdown.New(),
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

	log.Println("Formatting:", file)

	p, err := os.ReadFile(file)
	if err != nil {
		return err
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

	if err := s.md.Format(p, w); err != nil {
		return err
	}

	if err := w.Sync(); err != nil {
		return err
	}

	return os.Rename(w.Name(), file)
}
