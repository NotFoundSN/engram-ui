// engram-ui is a web viewer for engram persistent memory.
//
// Run `engram-ui help` for usage.
package main

import (
	"os"

	"github.com/Gentleman-Programming/engram-ui/internal/cli"
)

func main() {
	os.Exit(cli.Dispatch(os.Args[1:]))
}
