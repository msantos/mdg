# SYNOPSIS

mdg [*options*] [fmt|convert] [-|*directory*|*file*] [...]

# DESCRIPTION

Generate formatted markdown or HTML from markdown input.

By default, mdg reads from standard input and writes to standard output.

Arguments may be:
* `-`: read markdown from stdin (the default)
* file: path to markdown file
* directory: walk the specified path for any files ending with the
  `.md` extension

# BUILDING

```
go install git.iscode.ca/msantos/mgd/cmd/mgd@latest
```

# EXAMPLES

```bash
# current directory
mdg convert

# any markdown in the $HOME/docs directory
mdg convert ~/docs

# format from stdin
mdg fmt -
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
