package main

import (
	"fmt"
	"os"

	"github.com/standards-lab/docs-corpus/tools/coverage/internal/app"
)

func main() {
	if err := app.Dispose(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "dispose:", err)
		os.Exit(1)
	}
}
