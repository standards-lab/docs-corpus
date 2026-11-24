package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/standards-lab/docs-corpus/tools/coverage/internal/app"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	err := app.PlaceTraits(ctx, os.Args[1:], os.Stdout)
	stop()
	if err != nil {
		fmt.Fprintln(os.Stderr, "place-traits:", err)
		os.Exit(1)
	}
}
