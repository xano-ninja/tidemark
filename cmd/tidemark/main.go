package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "-h" || os.Args[1] == "--help") {
		fmt.Println("usage: tidemark <stations|today|week|windows|ics>")
		return
	}
	fmt.Fprintln(os.Stderr, "tidemark: no commands implemented yet")
	os.Exit(1)
}
