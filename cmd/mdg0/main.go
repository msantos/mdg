package main

import (
	"bytes"
	_ "embed"
	"flag"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"git.iscode.ca/msantos/mdg0/format"
	"git.iscode.ca/msantos/mdg0/markdown"
)

//go:embed index_tmpl.html
var indexHTML string

//go:embed default.css
var css string

type State struct {
	verbose bool

	md *markdown.Opt
	t  *template.Template
}

func main() {
	verbose := flag.Bool("verbose", false, "Enable debug messages")
	flag.Parse()

	dir := "."
	if flag.NArg() > 0 {
		dir = flag.Arg(0)
	}

	t, err := template.New("index").Parse(indexHTML)
	if err != nil {
		log.Fatalf("template: %v\n", err)
	}

	s := &State{
		md:      markdown.New(),
		verbose: *verbose,
		t:       t,
	}

	if err := s.run(dir); err != nil {
		log.Fatalln(err)
	}
}

type Metadata struct {
	Author     string
	Title      string
	Version    string
	Date       string
	Footer     string
	Styles     []string
	DefaultCSS string
	Body       string
}

func (s *State) run(dir string) error {
	return filepath.WalkDir(dir, s.convert)
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

	md, err := format.Parse(file, p)
	if err != nil {
		return err
	}

	var body bytes.Buffer

	if err := s.md.Convert(md.Content, &body); err != nil {
		return err
	}

	metadata := &Metadata{
		Author:     format.Field("author", md.FrontMatter),
		Title:      format.Field("title", md.FrontMatter),
		Version:    format.Field("version", md.FrontMatter),
		Date:       format.Field("date", md.FrontMatter),
		DefaultCSS: css,
		Body:       body.String(),
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

	return s.t.Execute(w, metadata)
}
