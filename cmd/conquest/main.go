package main

import (
	"fmt"
	"os"

	"apple-conquest/internal/game"
	"apple-conquest/internal/terminal"
)

func main() {
	if err := terminal.Run(os.Stdin, os.Stdout, game.NewGame()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
