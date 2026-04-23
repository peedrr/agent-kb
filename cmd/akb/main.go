package main

import (
	"fmt"
	"os"
)

var version = "0.6.0"

func main() {
	if err := Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
