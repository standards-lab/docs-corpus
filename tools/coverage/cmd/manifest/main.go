package main

import (
	"fmt"
	"os"

	"github.com/standards-lab/docs-corpus/tools/coverage/internal/app"
)

func main() {
	if err := app.Manifest(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "manifest:", err)
		os.Exit(1)
	}
}
