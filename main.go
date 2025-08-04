package main

import (
	"fmt"
	"os"

	"github.com/allexandrecardos/dck/cmd"
)

// var version = "dev"

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
