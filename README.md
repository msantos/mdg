[![Go Reference](https://pkg.go.dev/badge/go.iscode.ca/mdg.svg)](https://pkg.go.dev/go.iscode.ca/mdg)

# SYNOPSIS

mdg [*options*] [fmt|convert] [-|*directory*|*file*] [...]

# DESCRIPTION

Generate formatted markdown or HTML from markdown input.

By default, mdg reads from standard input and writes to standard output.

Arguments may be:
* `-`: read markdown from stdin (the default)
* file: path to markdown file
* directory: walk the specified path for any files ending with the
  `.md` or `.markdown` extensions

# BUILDING

```
go install go.iscode.ca/mdg/cmd/mdg@latest
```

## Source

```
CGO_ENABLED=0 go build -trimpath -ldflags "-w" ./cmd/mdg/
```

# EXAMPLES

## fmt

* format markdown input from stdin and output formatted markdown

```
mdg fmt
```

* format in place markdown files

```
mdg fmt test1.md doc/test2.md
```

* format in place markdown files ending with .md or .markdown in the current
  directory

```
mdg fmt .
```

## convert

* convert markdown input from stdin and output HTML

```
mdg convert
```

* convert markdown files ending in .md or .markdown to HTML in the current
  directory

```
mdg convert .
```

* convert markdown files ending in .md or .markdown to HTML in $HOME/docs
  directory and current working directory

```
mdg convert ~/docs .
```

# ENVIRONMENT VARIABLES

None.

# COMMANDS

## convert

Convert markdown documents to HTML.

### OPTIONS

css *string*
: CSS file

template *string*
: HTML template

verbose
: Enable debug messages

## format

Format markdown documents.

### OPTIONS

diff
: Display formatting changes as diff

verbose
: Enable debug messages
