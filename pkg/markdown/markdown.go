package markdown

import (
	"bytes"
	_ "embed"
	"io"
	"text/template"

	"go.iscode.ca/mdg/pkg/config"
	"go.iscode.ca/mdg/pkg/format"
	d2 "github.com/FurqanSoftware/goldmark-d2"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"go.abhg.dev/goldmark/anchor"
	mermaid "go.abhg.dev/goldmark/mermaid"
	"go.abhg.dev/goldmark/toc"
	"oss.terrastruct.com/d2/d2layouts/d2elklayout"
	"oss.terrastruct.com/d2/d2themes/d2themescatalog"
)

//go:embed default_tmpl.html
var defaultHTML string
var templateHTML *template.Template

//go:embed default.css
var defaultCSS string

func init() {
	t, err := template.New("html").Parse(defaultHTML)
	if err != nil {
		panic(err)
	}
	templateHTML = t
}

type Opt struct {
	goldmark.Markdown
	f        *format.Formatter
	linewrap bool
	css      string
	t        *template.Template
}

type Option func(*Opt)

// WithCSS sets the CSS for the markdown template.
func WithCSS(s string) Option {
	return func(o *Opt) {
		if s != "" {
			o.css = s
		}
	}
}

// WithLineWrap enables or disables wrapping long lines.
func WithLineWrap(t bool) Option {
	return func(o *Opt) {
		o.linewrap = t
	}
}

// WithTemplate sets the markdown template.
func WithTemplate(t *template.Template) Option {
	return func(o *Opt) {
		if t != nil {
			o.t = t
		}
	}
}

func New(opt ...Option) *Opt {
	o := &Opt{
		Markdown: goldmark.New(
			goldmark.WithParserOptions(parser.WithAutoHeadingID()),
			goldmark.WithExtensions(
				extension.GFM,
				meta.Meta,
				&mermaid.Extender{
					Theme: "neutral",
				},
				highlighting.Highlighting,
				&toc.Extender{},
				&anchor.Extender{},
				&d2.Extender{
					Layout:  d2elklayout.DefaultLayout,
					ThemeID: &d2themescatalog.TerminalGrayscale.ID,
				},
			),
		),
		t:        templateHTML,
		css:      defaultCSS,
		linewrap: true,
	}

	for _, fn := range opt {
		fn(o)
	}

	o.f = format.New(format.WithLineWrap(o.linewrap))

	return o
}

type Metadata struct {
	Author     string
	Title      string
	Creator    string
	Version    string
	VCS        string
	Date       string
	Footer     map[string]string
	Styles     []string
	DefaultCSS string
	Body       string
}

func metadata(key string, fm map[string]any, def string) string {
	s := format.String(key, fm)
	if s == "" {
		return def
	}

	return s
}

func (o *Opt) Convert(r io.Reader, w io.Writer) error {
	md, err := format.Parse(r)
	if err != nil {
		return err
	}

	var body bytes.Buffer

	if err := o.Markdown.Convert(md.Content, &body); err != nil {
		return err
	}

	metadata := &Metadata{
		Author:     format.String("author", md.FrontMatter),
		Title:      format.String("title", md.FrontMatter),
		Creator:    metadata("creator", md.FrontMatter, config.Name()),
		Version:    metadata("version", md.FrontMatter, config.Version()),
		VCS:        metadata("vcs", md.FrontMatter, config.Repo()),
		Date:       format.String("date", md.FrontMatter),
		Footer:     format.Map("footer", md.FrontMatter),
		DefaultCSS: o.css,
		Body:       body.String(),
	}

	return o.t.Execute(w, metadata)
}

func (o *Opt) Format(r io.Reader, w io.Writer) error {
	md, err := format.Parse(r)
	if err != nil {
		return err
	}

	return o.f.Format(w, md)
}
