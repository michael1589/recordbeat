package main

import (
	"os"

	"github.com/michael1589/recordbeat/cmd"

	_ "github.com/michael1589/recordbeat/include"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
