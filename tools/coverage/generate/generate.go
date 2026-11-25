package generate

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"mime"
	"net/mail"
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

// RightsManaged returns message, a multipart/mixed message with a boundary, with a
// Content-Class: rpmsg.message header added at the end of its header and payload attached, base64
// encoded, as message.rpmsg before its closing delimiter. It keeps the message's line endings.
func RightsManaged(message, payload []byte) ([]byte, error) {
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

// Mbox returns an mbox of 1,000 messages, messages in turn. A message that starts with a From
// line keeps it; any other gets a From MAILER-DAEMON line with now in UTC. A line in a message
// body that starts with From is escaped as >From.
func Mbox(messages [][]byte, now time.Time) []byte {
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
