package main

import (
	"fmt"
	"os"

	"github.com/standards-lab/docs-corpus/tools/coverage/disposition"
)

func main() {
	if err := disposition.Run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "dispose:", err)
		os.Exit(1)
	}
}
