package main

import (
	"flag"
	"fmt"
	"os"

	"apple-conquest/internal/terminal"
)

func main() {
	noColor := flag.Bool("no-color", false, "disable ANSI colors")
	flag.Parse()
	if err := terminal.Run(*noColor); err != nil {
		fmt.Fprintln(os.Stderr, "Apple Conquest:", err)
		os.Exit(1)
	}
}
