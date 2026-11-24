// Command generate writes the corpus's generated candidates from adopted samples. Run it from
// the repository root.
//
//	go run ./scripts/generate rights-managed <message.eml> <payload> <out.eml>
//	go run ./scripts/generate mbox <out.mbox> <message.eml>...
//
// rights-managed adds a Content-Class: rpmsg.message header to a multipart/mixed message and
// attaches the payload as message.rpmsg. mbox writes 1,000 messages, the given ones in turn.
//
// See docs/coverage.md
package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"mime"
	"net/mail"
	"os"
	"time"
)

const (
	contentClass   = "rpmsg.message"
	rpmsgName      = "message.rpmsg"
	rpmsgType      = "application/x-microsoft-rpmsg-message"
	base64LineSize = 76
	mboxMessages   = 1000
)

func newlineOf(data []byte) []byte {
	if i := bytes.IndexByte(data, '\n'); i > 0 && data[i-1] == '\r' {
		return []byte("\r\n")
	}
	return []byte("\n")
}

func rightsManaged(message, payload []byte) ([]byte, error) {
	msg, err := mail.ReadMessage(bytes.NewReader(message))
	if err != nil {
		return nil, err
	}
	mediaType, params, err := mime.ParseMediaType(msg.Header.Get("Content-Type"))
	if err != nil {
		return nil, err
	}
	if mediaType != "multipart/mixed" || params["boundary"] == "" {
		return nil, fmt.Errorf("the message is %s, not multipart/mixed with a boundary", mediaType)
	}
	nl := newlineOf(message)
	headerEnd := bytes.Index(message, append(append([]byte{}, nl...), nl...))
	if headerEnd < 0 {
		return nil, errors.New("the message has no blank line after its header")
	}
	headerEnd += len(nl)
	delimiter := []byte("--" + params["boundary"])
	closing := bytes.LastIndex(message, append(append(append([]byte{}, nl...), delimiter...), "--"...))
	if closing < headerEnd {
		return nil, fmt.Errorf("the message has no closing %s-- line", delimiter)
	}
	closing += len(nl)

	var out bytes.Buffer
	line := func(s string) {
		out.WriteString(s)
		out.Write(nl)
	}
	out.Write(message[:headerEnd])
	line("Content-Class: " + contentClass)
	out.Write(message[headerEnd:closing])
	line(string(delimiter))
	line("Content-Type: " + rpmsgType)
	line("Content-Transfer-Encoding: base64")
	line(`Content-Disposition: attachment; filename="` + rpmsgName + `"`)
	line("MIME-Version: 1.0")
	out.Write(nl)
	encoded := base64.StdEncoding.EncodeToString(payload)
	for len(encoded) > 0 {
		n := min(base64LineSize, len(encoded))
		line(encoded[:n])
		encoded = encoded[n:]
	}
	out.Write(message[closing:])
	return out.Bytes(), nil
}

func mbox(messages [][]byte, now time.Time) []byte {
	var out bytes.Buffer
	for i := range mboxMessages {
		msg := messages[i%len(messages)]
		from := []byte("From MAILER-DAEMON " + now.UTC().Format("Mon Jan _2 15:04:05 2006"))
		if bytes.HasPrefix(msg, []byte("From ")) {
			if nl := bytes.IndexByte(msg, '\n'); nl >= 0 {
				from, msg = msg[:nl], msg[nl+1:]
			} else {
				from, msg = msg, nil
			}
		}
		out.Write(from)
		out.WriteByte('\n')
		msg = bytes.ReplaceAll(msg, []byte("\nFrom "), []byte("\n>From "))
		out.Write(msg)
		if !bytes.HasSuffix(msg, []byte("\n")) {
			out.WriteByte('\n')
		}
		out.WriteByte('\n')
	}
	return out.Bytes()
}

func generate(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: generate rights-managed <message.eml> <payload> <out.eml> | mbox <out.mbox> <message.eml>...")
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
		out, err := rightsManaged(message, payload)
		if err != nil {
			return fmt.Errorf("%s: %w", args[1], err)
		}
		return os.WriteFile(args[3], out, 0o644)
	case "mbox":
		if len(args) < 3 {
			return errors.New("usage: generate mbox <out.mbox> <message.eml>...")
		}
		var messages [][]byte
		for _, path := range args[2:] {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			messages = append(messages, data)
		}
		return os.WriteFile(args[1], mbox(messages, time.Now()), 0o644)
	}
	return fmt.Errorf("unknown candidate %q", args[0])
}

func main() {
	if err := generate(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "generate:", err)
		os.Exit(1)
	}
}
