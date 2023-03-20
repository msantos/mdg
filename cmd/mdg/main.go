package main

import (
	"flag"
	"fmt"
	"os"
	"path"

	"git.iscode.ca/msantos/mdg/cmd/mdg/convert"
	"git.iscode.ca/msantos/mdg/config"
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
	fmt.Fprintf(os.Stderr, "Options:\n\n")
	f.PrintDefaults()
}

func main() {
	f.Usage = func() { usage() }
	_ = f.Parse(os.Args[1:])

	oargs := f.Args()

	command := "help"

	if len(os.Args) > 1 {
		command = oargs[0]
	}

	var args []string
	if len(oargs) > 1 {
		args = oargs[1:]
	}

	os.Args = append(os.Args[:1], args...)

	switch command {
	case "convert":
		convert.Run()
	case "format":
		fallthrough
	case "fmt":
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
