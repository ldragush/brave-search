package ui

import (
	"fmt"
	"io"
)

func PrintBanner(w io.Writer, codename, version string) {
	fmt.Fprintln(w, `
   __                                              __ 
  / /  _______ __  _____ _______ ___ ___ _________/ / 
 / _ \/ __/ _  / |/ / -_)___(_-</ -_) _  / __/ __/ _ \
/_.__/_/  \_,_/|___/\__/   /___/\__/\_,_/_/  \__/_//_/

haltman.io (https://github.com/haltman-io)

[codename: `+codename+`] - [release: `+version+`]
`)
}
