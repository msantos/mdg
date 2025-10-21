# SYNOPSIS

mdg [*options*] [*command*] [-|*directory*]

# DESCRIPTION

Generate formatted markdown or HTML from markdown input.

mdg requires an argument:

* path: mdg walks the specified path for any files ending with the
  `.md` extension
* `-`: read markdown from stdin

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
