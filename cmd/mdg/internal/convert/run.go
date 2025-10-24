package convert

import (
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"git.iscode.ca/msantos/mdg/internal/pkg/fdpair"
	"git.iscode.ca/msantos/mdg/pkg/config"
	"git.iscode.ca/msantos/mdg/pkg/markdown"
)

type Opt struct {
	verbose bool
	check   string
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
	check := flag.String("check", "newer", "Compare markdown files to HTML before conversion: newer, disable")
	verbose := flag.Bool("verbose", false, "Enable debug messages")

	flag.Usage = func() { usage() }

	flag.Parse()

	args := []string{"-"}
	if flag.NArg() > 0 {
		args = flag.Args()
	}

	cssContent := ""
	if *css != "" {
		b, err := os.ReadFile(*css)
		if err != nil {
			fmt.Fprintf(os.Stderr, "css: %v\n", err)
			os.Exit(1)
		}
		cssContent = string(b)
	}

	var t *template.Template

	if *tmpl != "" {
		b, err := os.ReadFile(*tmpl)
		if err != nil {
			fmt.Fprintf(os.Stderr, "template: %v\n", err)
			os.Exit(1)
		}

		t, err = template.New("index").Parse(string(b))
		if err != nil {
			fmt.Fprintf(os.Stderr, "template: %v\n", err)
			os.Exit(1)
		}
	}

	o := &Opt{
		md:      markdown.New(markdown.WithTemplate(t), markdown.WithCSS(cssContent)),
		check:   *check,
		verbose: *verbose,
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
		o.check = "disable"

		return o.convert(fdpair.New())
	}

	return filepath.WalkDir(dir, o.walkdir)
}

func (o *Opt) convert(r *fdpair.Opt) error {
	w, err := r.OpenOutput(r.File)

	switch {
	case err == nil:
	case errors.Is(err, ErrSkipMD):
		return nil
	default:
		return fmt.Errorf("%s: %w", r.Name(), err)
	}

	defer func() {
		if rerr := r.CloseOutput(r.File, w); rerr != nil {
			if err == nil {
				err = fmt.Errorf("%s: %w", w.Name(), rerr)
			}
		}
	}()

	if err := o.md.Convert(r, w); err != nil {
		return fmt.Errorf("%s: %w", w.Name(), err)
	}

	return nil
}

func (o *Opt) compare(md, html string) bool {
	switch o.check {
	case "", "disable":
		return true
	case "newer":
	default:
		fmt.Fprintln(os.Stderr, "invalid check:", o.check)
		return false
	}

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

var ErrSkipMD = errors.New("skip markdown file")

func (o *Opt) openOutput(r *os.File) (*os.File, error) {
	html := strings.TrimSuffix(r.Name(), filepath.Ext(r.Name())) + ".html"

	if !o.compare(r.Name(), html) {
		return nil, ErrSkipMD
	}

	if o.verbose {
		fmt.Fprintln(os.Stderr, "Converting:", r.Name(), " -> ", html)
	}

	w, err := os.OpenFile(html, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", html, err)
	}

	return w, nil
}

func (o *Opt) closeOutput(_, w *os.File) error {
	return w.Close()
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

	fd := fdpair.New()

	fd.File = r
	fd.OpenOutput = o.openOutput
	fd.CloseOutput = o.closeOutput

	if err := o.convert(fd); err != nil {
		return fmt.Errorf("%s: %w", file, err)
	}

	return nil
}
