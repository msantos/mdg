// Package format formats a markdown document reproducibly and
// consistently.
package format

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/Kunde21/markdownfmt/v2/markdown"
	"github.com/bwplotka/mdox/pkg/gitdiff"
	"github.com/gohugoio/hugo/parser/pageparser"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
)

var ErrNotFormatted = errors.New("markdown not formatted")

type Markdown struct {
	FrontMatter map[string]any
	Content     []byte
	source      []byte
	name        string
}

// Parse returns the frontmatter and markdown content.
func Parse(name string, source []byte) (*Markdown, error) {
	md := &Markdown{
		Content: source,
		source:  source,
		name:    name,
	}

	fm, err := pageparser.ParseFrontMatterAndContent(bytes.NewReader(source))
	if err != nil {
		return md, err
	}

	if len(fm.FrontMatter) > 0 {
		md.Content = fm.Content
		md.FrontMatter = fm.FrontMatter
	}

	return md, nil
}

// WriteFrontMatter converts the parsed metadata into YAML.
func (md *Markdown) WriteFrontMatter(w io.Writer) error {
	if len(md.FrontMatter) == 0 {
		return nil
	}

	b, err := FormatFrontMatter(md.FrontMatter)
	if err != nil {
		return err
	}

	if _, err := w.Write(b); err != nil {
		return err
	}

	return nil
}

type Formatter struct {
	style Style
}

type Option func(*Formatter)

// WithStyle sets the method of formatting the document.
func WithStyle(s Style) Option {
	return func(f *Formatter) {
		f.style = s
	}
}

// New configures the formatter.
func New(opt ...Option) *Formatter {
	f := &Formatter{}

	for _, fn := range opt {
		fn(f)
	}

	return f
}

// Format formats and writes a parsed markdown document to the provided
// writer.
func (f *Formatter) Format(w io.Writer, md *Markdown) error {
	if f.style == StyleNone {
		_, err := w.Write(md.source)
		return err
	}

	if err := md.WriteFrontMatter(w); err != nil {
		return err
	}

	renderer := markdown.NewRenderer()
	if f.style == StyleWrap {
		renderer.AddMarkdownOptions(markdown.WithSoftWraps())
	}

	return goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
		),
		goldmark.WithParserOptions(
			parser.WithAttribute(), /* Enable # headers {#custom-ids} */
			parser.WithHeadingAttribute(),
		),
		goldmark.WithRenderer(renderer),
	).Convert(md.Content, w)
}

// Diff verifies if the document has been formatted. If the document
// is unformatted, Diff returns the differences and the ErrNotFormatted
// error. Otherwise diff will be an empty string.
func (f *Formatter) Diff(md *Markdown) (string, error) {
	if f.style == StyleNone {
		return "", nil
	}

	b := bytes.Buffer{}

	if err := f.Format(&b, md); err != nil {
		return "", err
	}

	if bytes.Equal(md.source, b.Bytes()) {
		return "", nil
	}

	d := gitdiff.CompareBytes(
		md.source, md.name,
		b.Bytes(), md.name+" (formatted)",
	)

	return string(d.ToCombinedFormat()), ErrNotFormatted
}

func Field(key string, fm map[string]any) string {
	val, ok := fm[key]
	if !ok {
		return ""
	}
	switch v := val.(type) {
	case string:
		return fmt.Sprintf("%s", v)
	case []interface{}:
		a := make([]string, 0, len(v))
		for _, x := range v {
			a = append(a, fmt.Sprintf("%s", x.(string)))
		}
		return fmt.Sprintf("%s", strings.Join(a, ", "))
	}
	return ""
}

func Map(key string, fm map[string]any) map[string]string {
	val, ok := fm[key]
	if !ok {
		return nil
	}
	mi, ok := val.(map[string]interface{})
	if !ok {
		return nil
	}
	m := make(map[string]string)
	for k, v := range mi {
		s, ok := v.(string)
		if ok {
			m[k] = s
		}
	}
	return m
}
