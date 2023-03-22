# SYNOPSIS

mdg [*options*] [*command*] [-|*directory*]

# DESCRIPTION

mdg generates other formats from markdown. By default, mdg walks the
current directory for any files ending with the `.md` extension.

Specify `-` to read markdown from stdin or a path to read from another
directory.

# BUILDING

```
go install git.iscode.ca/msantos/mgd/cmd/mgd@latest
```

# EXAMPLES

```
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
