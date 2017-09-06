package main

import (
	"os"

	"github.com/gravewrought/cryptocompbeat/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
