package main

import (
	"fmt"
	"os"
)

var version = "dev"

func main() {
	if err := Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
