package mail

import (
	"bytes"

	"github.com/standards-lab/docs-corpus/tools/coverage/idset"
)

// MSGBodies returns the body properties an Outlook .msg file's data names in its streams, of
// 1000 (plain text), 1009 (RTF), and 1013 (HTML), in that order.
func MSGBodies(data []byte) []string {
	prefix := []byte("_\x00_\x00s\x00u\x00b\x00s\x00t\x00g\x001\x00.\x000\x00_\x00")
	found := idset.Set{}
	for rest := data; ; {
		i := bytes.Index(rest, prefix)
		if i < 0 {
			break
		}
		rest = rest[i+len(prefix):]
		if len(rest) < 8 {
			continue
		}
		var id []byte
		for k := range 4 {
			c := rest[2*k]
			if rest[2*k+1] != 0 || (c < '0' || c > '9') && (c < 'A' || c > 'F') {
				id = nil
				break
			}
			id = append(id, c)
		}
		if id != nil {
			found.Add(string(id))
			rest = rest[8:]
		}
	}
	bodies := []string{}
	for _, id := range []string{"1000", "1009", "1013"} {
		if found[id] {
			bodies = append(bodies, id)
		}
	}
	return bodies
}
