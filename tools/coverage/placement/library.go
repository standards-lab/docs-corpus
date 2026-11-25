package placement

import (
	"github.com/standards-lab/docs-corpus/tools/coverage/idset"
	"github.com/standards-lab/docs-corpus/tools/coverage/run"
)

type libraryEntry struct {
	FedBy *struct {
		Routes []string `json:"routes"`
	} `json:"fed_by"`
}

func capabilities(library map[string]libraryEntry) map[string][]string {
	fed := map[string][]string{}
	for id, entry := range library {
		if entry.FedBy == nil {
			continue
		}
		for _, route := range entry.FedBy.Routes {
			fed[route] = append(fed[route], id)
		}
	}
	return fed
}

func placeLibrary(events []run.Event, fed map[string][]string) idset.Set {
	ids := idset.Set{}
	for _, e := range events {
		if kind := e.Part.ContentType.Kind; kind != "" {
			ids.Add("kind:" + kind)
		}
		if route := e.Part.ContentType.Route; route != nil {
			ids.Add("route:" + *route)
		}
		for _, f := range e.Part.Flags {
			ids.Add("check:" + f.Check)
		}
	}
	for _, route := range ids.Sorted("route:") {
		ids.Add(fed[route]...)
	}
	return ids
}
