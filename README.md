# SYNOPSIS

iscode [*options*] [*address*]:*port*

# DESCRIPTION

iscode is the https://iscode.ca web service.

# BUILDING

```
go install git.iscode.ca/iscode.ca/iscode/cmd/iscode@latest
```

# EXAMPLES

```
iscode :8080

iscode --lib=d2 :8080
```

# ENVIRONMENT VARIABLES

RESUME_FILE
: the default resume.json input

RESUME_USER
: the API user

RESUME_CACHEDIR
: set the cache directory (default `$HOME/.cache`)

# OPTIONS

lib *string*
: Default diagram library: mermaid, d2 (default `mermaid`)

proxyproto
: Enable proxy protocol

resume string
: path to resume.json file (default `resume.json`)

timeout duration
: Read timeout (default `5s`)

verbose
: Enable debug messages
