package main

import (
	"log"

	"apple-conquest/internal/desktop"
)

func main() {
	if err := desktop.Run(); err != nil {
		log.Fatal(err)
	}
}
