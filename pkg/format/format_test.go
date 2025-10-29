package format_test

import (
	"bytes"
	"testing"

	"codeberg.org/msantos/mdg/pkg/format"
)

const (
	mdUnformatted = `Test
====

Subtest
-------

This is a
line with
breaks.

option
: list
`

	mdFormatted = `# Test

## Subtest

This is a
line with
breaks.

option
: list
`

	mdFrontMatterUnformatted = `---
author: Firstname Lastname
title: The Title Goes Here
date: "2022-10-27"
version: 1.0.0
status: proposal
---

Test
====

Subtest
-------

This is a
line with
breaks.

option
: list
`

	mdFrontMatterFormatted = `---
version: 1.0.0
title: The Title Goes Here
status: proposal
date: "2022-10-27"
author: Firstname Lastname
---

# Test

## Subtest

This is a
line with
breaks.

option
: list
`
)

func TestFormatMarkdown(t *testing.T) {
	md, err := format.Parse("test", bytes.NewBufferString(mdUnformatted))
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	f := format.New(format.WithLineWrap(true))

	b := &bytes.Buffer{}

	if err := f.Format(b, md); err != nil {
		t.Errorf("%v", err)
		return
	}

	if !bytes.Equal([]byte(mdFormatted), b.Bytes()) {
		t.Errorf("markdown unformatted")
		return
	}
}

func TestFormatMarkdownFrontmatter(t *testing.T) {
	md, err := format.Parse("test", bytes.NewBufferString(mdFrontMatterUnformatted))
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	f := format.New(format.WithLineWrap(true))

	b := &bytes.Buffer{}

	if err := f.Format(b, md); err != nil {
		t.Errorf("%v", err)
		return
	}

	if !bytes.Equal([]byte(mdFrontMatterFormatted), b.Bytes()) {
		t.Errorf("markdown unformatted")
		return
	}
}

func TestDiffMarkdown(t *testing.T) {
	md, err := format.Parse("test", bytes.NewBufferString(mdFrontMatterUnformatted))
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	f := format.New(format.WithLineWrap(true))

	diff, err := f.Diff(md)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	if diff == "" {
		t.Errorf("markdown: no difference found")
		return
	}
}

func TestNoDiffMarkdown(t *testing.T) {
	md, err := format.Parse("test", bytes.NewBufferString(mdFrontMatterFormatted))
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	f := format.New(format.WithLineWrap(true))

	diff, err := f.Diff(md)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	if diff != "" {
		t.Errorf("markdown: differences found: %s", diff)
		return
	}
}
