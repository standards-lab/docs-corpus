// Package run reads a kept docproc run: its record.jsonl's events, grouped by input path relative
// to the run's input, and its outputs' blobs in the content store beside the runs directory.
//
// Its exported identifiers:
//
//   - [Load] reads one run.
//   - [Record] is a loaded run; [Record.Events] and [Record.Outputs] return an input's events,
//     [Record.Text] the text the run extracted from it, and [Record.Rows] its record outputs'
//     lines.
//   - [Event] is one record line, [Output] an output event's output, with [Output.MetaRaw]
//     returning one meta value's JSON, and [Row] one record output line.
package run
