// Package jsonvalue reads and writes JSON as Python's json module does: objects keep their key
// order, numbers keep their text, and floats are written in Python's repr.
//
// Its exported identifiers:
//
//   - [Object] is a JSON object in key order; [NewObject] makes one, and [Object.Set],
//     [Object.Get], [Object.Has], [Object.Str], and [Object.Keys] use it.
//   - [Float] is a number written in Python's float repr.
//   - [Parse] parses one JSON value.
//   - [Marshal] writes a value as json.dumps does with an indent of 2.
//   - [Truthy] reads a value as Python's bool does.
//   - [String] returns a value that is a string.
//   - [Compare] orders two values as Python does.
package jsonvalue
