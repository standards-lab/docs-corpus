#!/usr/bin/env bash
# Prints the library placement of every ~/docs sample as a JSON array, sorted by path. Writes
# nothing.
# See docs/coverage.md
set -euo pipefail

cd "$(dirname "$0")/.."

run=$(jq -er '.library_placement.run' docs/coverage.json)
dir=${RUNS:-$HOME/runs/docs-corpus/docproc-final/docs/runs}/$run
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
