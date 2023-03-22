package convert

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
	"text/template"

	"git.iscode.ca/msantos/mdg/config"
	"git.iscode.ca/msantos/mdg/markdown"
)

type State struct {
	verbose bool
	md      *markdown.Opt
}

func usage() {
	fmt.Fprintf(os.Stderr, `%s %s
Usage: %s convert [<option>] [-|<path>]

Convert markdown documents to HTML.

`, path.Base(os.Args[0]), config.Version(), os.Args[0])
	fmt.Fprintf(os.Stderr, "Options:\n\n")
	flag.PrintDefaults()
}

func Run() {
	css := flag.String("css", "", "CSS file")
	tmpl := flag.String("template", "", "HTML template")
	verbose := flag.Bool("verbose", false, "Enable debug messages")

	flag.Usage = func() { usage() }

	flag.Parse()

	dir := "."
	if flag.NArg() > 0 {
		dir = flag.Arg(0)
	}

	cssContent := ""
	if *css != "" {
		b, err := os.ReadFile(*css)
		if err != nil {
			log.Fatalf("css: %v\n", err)
		}
		cssContent = string(b)
	}

	var t *template.Template

	if *tmpl != "" {
		b, err := os.ReadFile(*tmpl)
		if err != nil {
			log.Fatalf("template: %v\n", err)
		}

		t, err = template.New("index").Parse(string(b))
		if err != nil {
			log.Fatalf("template: %v\n", err)
		}
	}

	s := &State{
		md:      markdown.New(markdown.WithTemplate(t), markdown.WithCSS(cssContent)),
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
	return filepath.WalkDir(dir, s.convert)
}

func (s *State) stdin() error {
	p, err := io.ReadAll(os.Stdin)
	if err != nil {
		return err
	}
	return s.md.Convert(p, os.Stdout)
}

func (s *State) newer(md, html string) bool {
	stmd, err := os.Stat(md)
	if err != nil {
		return false
	}

	sthtml, err := os.Stat(html)
	if err != nil {
		return true
	}

	return stmd.ModTime().After(sthtml.ModTime())
}

func (s *State) convert(file string, d fs.DirEntry, err error) error {
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

	html := strings.TrimSuffix(file, filepath.Ext(file)) + ".html"

	if !s.newer(file, html) {
		return nil
	}

	log.Println("Converting:", file, " -> ", html)

	p, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	w, err := os.OpenFile(html, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	defer func() {
		if err := w.Close(); err != nil {
			log.Println(err)
		}
	}()

	return s.md.Convert(p, w)
}
