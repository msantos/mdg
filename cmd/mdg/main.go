package main

import (
	"flag"
	"fmt"
	"os"
	"path"

	"go.iscode.ca/mdg/cmd/mdg/internal/convert"
	"go.iscode.ca/mdg/cmd/mdg/internal/format"
	"go.iscode.ca/mdg/pkg/config"
)

var f = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

func usage() {
	fmt.Fprintf(os.Stderr, `%s %s
Usage: %s <command> [<option>]

Convert and format markdown.

Commands:

      convert  - convert markdown to HTML
      fmt      - format markdown
      version  - display version

`, path.Base(os.Args[0]), config.Version(), os.Args[0])
	f.PrintDefaults()
}

func main() {
	f.Usage = func() { usage() }
	_ = f.Parse(os.Args[1:])

	command := "help"

	if len(os.Args) > 1 {
		command = f.Args()[0]
	}

	var args []string
	if f.NArg() > 1 {
		args = f.Args()[1:]
	}

	os.Args = append(os.Args[:1], args...)

	switch command {
	case "convert":
		convert.Run()
	case "fmt", "format":
		format.Run()
	case "help":
		usage()
	case "version":
		fmt.Println(config.Version())
	default:
		fmt.Println("command not found:", command)
		os.Exit(127)
	}

	os.Exit(0)
}
