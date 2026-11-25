package app

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/standards-lab/docs-corpus/tools/coverage/generate"
)

// Generate makes one candidate from the command line's arguments, args without the program's
// name, writing its output file: rights-managed <message.eml> <payload> <out.eml>, or mbox
// <out.mbox> <message.eml> [<message.eml> ...], made at now. Paths are relative to the working
// directory.
func Generate(args []string, now time.Time) error {
	if len(args) == 0 {
		return errors.New("usage: generate rights-managed <message.eml> <payload> <out.eml> | mbox <out.mbox> <message.eml> [<message.eml> ...]")
	}
	switch args[0] {
	case "rights-managed":
		if len(args) != 4 {
			return errors.New("usage: generate rights-managed <message.eml> <payload> <out.eml>")
		}
		message, err := os.ReadFile(args[1])
		if err != nil {
			return err
		}
		payload, err := os.ReadFile(args[2])
		if err != nil {
			return err
		}
		out, err := generate.RightsManaged(message, payload)
		if err != nil {
			return fmt.Errorf("%s: %w", args[1], err)
		}
		return os.WriteFile(args[3], out, 0o644)
	case "mbox":
		if len(args) < 3 {
			return errors.New("usage: generate mbox <out.mbox> <message.eml> [<message.eml> ...]")
		}
		var messages [][]byte
		for _, path := range args[2:] {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			messages = append(messages, data)
		}
		return os.WriteFile(args[1], generate.Mbox(messages, now), 0o644)
	}
	return fmt.Errorf("unknown candidate %q", args[0])
}
