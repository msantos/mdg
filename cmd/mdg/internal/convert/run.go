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

	"go.iscode.ca/mdg/internal/pkg/fdpair"
	"go.iscode.ca/mdg/pkg/config"
	"go.iscode.ca/mdg/pkg/markdown"
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

		return o.convert(&stdio{
			r:   os.Stdin,
			Opt: o,
		})
	}

	return filepath.WalkDir(dir, o.walkdir)
}

func (o *Opt) convert(rw fdpair.FD) error {
	var in string

	if f, ok := rw.In().(*os.File); ok {
		in = f.Name()
	}

	err := rw.Open()

	switch {
	case err == nil:
	case errors.Is(err, ErrSkipMD):
		return nil
	default:
		return fmt.Errorf("%s: %w", in, err)
	}

	var out string

	if f, ok := rw.Out().(*os.File); ok {
		out = f.Name()
	}

	defer func() {
		if rerr := rw.Close(); rerr != nil {
			if err == nil {
				err = fmt.Errorf("%s: %w", out, rerr)
			}
		}
	}()

	if err := o.md.Convert(rw.In(), rw.Out()); err != nil {
		return fmt.Errorf("%s: %w", out, err)
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

	rw := &fsobj{
		r:   r,
		Opt: o,
	}

	if err := o.convert(rw); err != nil {
		return fmt.Errorf("%s: %w", file, err)
	}

	return nil
}
