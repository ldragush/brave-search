package main

import (
	"os"

	"github.com/haltman-io/brave-search/internal/app"
)

const (
	Codename = "brave-search"
	Version  = "v1.0.0-stable"
)

func main() {
	os.Exit(app.Run(Codename, Version, os.Args[1:]))
}
