package markdown

import (
	_ "embed"

	"git.iscode.ca/msantos/goldmark-d2"
	"git.iscode.ca/msantos/goldmark-mermaid"
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

type Opt struct {
	goldmark.Markdown
}

type Option func(*Opt)

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
					ThemeID: d2themescatalog.TerminalGrayscale.ID,
					Sketch:  true,
				},
			),
		),
	}
	for _, fn := range opt {
		fn(o)
	}

	return o
}
