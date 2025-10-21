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

type Opt struct {
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

	if flag.NArg() < 1 {
		usage()
		os.Exit(2)
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

	o := &Opt{
		md:      markdown.New(markdown.WithTemplate(t), markdown.WithCSS(cssContent)),
		verbose: *verbose,
	}

	for _, v := range flag.Args() {
		if err := o.run(v); err != nil {
			log.Fatalln(err)
		}
	}
}

func (o *Opt) run(dir string) error {
	if dir == "-" {
		return o.stdin()
	}

	return filepath.WalkDir(dir, o.convert)
}

func (o *Opt) stdin() error {
	b, err := io.ReadAll(os.Stdin)
	if err != nil {
		return err
	}

	return o.md.Convert(b, os.Stdout)
}

func (o *Opt) newer(md, html string) bool {
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

func (o *Opt) convert(file string, d fs.DirEntry, err error) error {
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

	if !o.newer(file, html) {
		return nil
	}

	if o.verbose {
		log.Println("Converting:", file, " -> ", html)
	}

	b, err := os.ReadFile(file)
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

	return o.md.Convert(b, w)
}
