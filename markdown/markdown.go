package markdown

import (
	"bytes"
	_ "embed"
	"io"
	"text/template"

	"git.iscode.ca/msantos/goldmark-d2"
	"git.iscode.ca/msantos/goldmark-mermaid"
	"git.iscode.ca/msantos/mdg/format"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark-highlighting/v2"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"go.abhg.dev/goldmark/anchor"
	"go.abhg.dev/goldmark/toc"
	"oss.terrastruct.com/d2/d2layouts/d2elklayout"
	"oss.terrastruct.com/d2/d2themes/d2themescatalog"
)

//go:embed default_tmpl.html
var defaultHTML string

//go:embed default.css
var defaultCSS string

type Opt struct {
	goldmark.Markdown
	css string
	t   *template.Template
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

// WithTemplate sets the markdown template.
func WithTemplate(t *template.Template) Option {
	return func(o *Opt) {
		if t != nil {
			o.t = t
		}
	}
}

func New(opt ...Option) *Opt {
	t, err := template.New("html").Parse(defaultHTML)
	if err != nil {
		panic(err)
	}

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
					ThemeID: d2themescatalog.TerminalGrayscale.ID,
					Sketch:  true,
				},
			),
		),
		t:   t,
		css: defaultCSS,
	}

	for _, fn := range opt {
		fn(o)
	}

	return o
}

type Metadata struct {
	Author     string
	Title      string
	Version    string
	Date       string
	Footer     map[string]string
	Styles     []string
	DefaultCSS string
	Body       string
}

func (o *Opt) Convert(content []byte, w io.Writer) error {
	md, err := format.Parse("", content)
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
		Version:    format.String("version", md.FrontMatter),
		Date:       format.String("date", md.FrontMatter),
		Footer:     format.Map("footer", md.FrontMatter),
		DefaultCSS: o.css,
		Body:       body.String(),
	}

	return o.t.Execute(w, metadata)
}
