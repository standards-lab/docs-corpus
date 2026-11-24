package main

import (
	"fmt"
	"os"
	"time"

	"github.com/standards-lab/docs-corpus/tools/coverage/generate"
)

func main() {
	if err := generate.Run(os.Args[1:], time.Now()); err != nil {
		fmt.Fprintln(os.Stderr, "generate:", err)
		os.Exit(1)
	}
}
