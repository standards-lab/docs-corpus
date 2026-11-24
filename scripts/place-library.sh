#!/usr/bin/env bash
# Prints the library placement of every ~/docs sample as a JSON array, sorted by path: each
# sample's path and sha256 from ${DOCS_REFERENCE:-$HOME/docs}/MANIFEST.json, and `library`, the
# sorted kind:, route:, check:, and capability: IDs that docproc's kept run observed for the file
# and every part found inside it. The run is .library_placement.run in docs/coverage.json, read
# from ${RUNS:-$HOME/runs/docproc/final/runs}/<run>/; a capability is placed wherever a route in
# its vocabulary entry's fed_by.routes is. A sample the run never reached gets an empty library.
# Reads the run, the manifest, and docs/coverage.json; writes nothing.
# See docs/coverage.md
set -euo pipefail

cd "$(dirname "$0")/.."

run=$(jq -er '.library_placement.run' docs/coverage.json)
dir=${RUNS:-$HOME/runs/docproc/final/runs}/$run
manifest=${DOCS_REFERENCE:-$HOME/docs}/MANIFEST.json

jq -n \
  --slurpfile run "$dir/manifest.json" \
  --slurpfile docs "$manifest" \
  --slurpfile coverage docs/coverage.json '
  ($run[0].inputs[0] + "/") as $root
  | [$coverage[0].vocabulary.library | to_entries[]
      | select(.key | startswith("capability:"))
      | .key as $capability | .value.fed_by.routes[] | {route: ., $capability}]
    as $fed
  | (reduce inputs as $e ({};
      .[$e.input | ltrimstr($root)] += [
        ($e.part.content_type.kind // empty | select(. != "") | "kind:" + .),
        ($e.part.content_type.route // empty | "route:" + .),
        ($e.part.flags // [] | .[].check | "check:" + .)
      ]))
    as $observed
  | [$docs[0].files[] | select(.role == "sample")
      | ($observed[.path] // [] | unique) as $ids
      | {path, sha256,
         library: ($ids + [$fed[] | select(IN(.route; $ids[])) | .capability] | unique)}]
  | sort_by(.path)
' "$dir/record.jsonl"
