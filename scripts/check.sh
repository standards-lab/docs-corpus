#!/usr/bin/env bash
# Checks the corpus once MANIFEST.json is present: every manifest entry names a file of its size
# and sha256, every corpus file has an entry, and no two samples share a format and a `covers`
# set. Checks the coverage data once it is present, at $COVERAGE or docs/coverage.json: its
# vocabulary's IDs and entries have their axis's shape, every `covers` ID in it is in the
# vocabulary, every capability's fed_by route and every sample's library ID is a library ID, and
# no sample path appears twice. When ${DOCS_REFERENCE:-$HOME/docs}/MANIFEST.json exists, every
# sample in it appears in the coverage once, by path and sha256, and the coverage lists no other.
# Then runs lint-context and lint-references, and fails if any part fails. Reads no network.
# See docs/coverage.md
set -euo pipefail

cd "$(dirname "$0")/.."

fail=0
err() {
  echo "FAIL: $*" >&2
  fail=1
}

areas=(archive audio email image office onenote pdf text video)

if [ -f MANIFEST.json ]; then
  while IFS=$'\t' read -r path bytes sha; do
    if [ ! -f "$path" ]; then
      err "$path: in MANIFEST.json but missing"
      continue
    fi
    [ "$(stat -c %s "$path")" = "$bytes" ] || err "$path: size differs from MANIFEST.json"
    [ "$(sha256sum "$path" | cut -d' ' -f1)" = "$sha" ] || err "$path: sha256 differs from MANIFEST.json"
  done < <(jq -r '.files[] | [.path, (.bytes|tostring), .sha256] | @tsv' MANIFEST.json)

  existing_areas=()
  for area in "${areas[@]}"; do
    [ -d "$area" ] && existing_areas+=("$area")
  done
  if [ "${#existing_areas[@]}" -gt 0 ]; then
    comm -23 \
      <(find "${existing_areas[@]}" -type f | sort) \
      <(jq -r '.files[].path' MANIFEST.json | sort) |
      while read -r path; do
        echo "FAIL: $path: not in MANIFEST.json" >&2
        echo x
      done | grep -q x && fail=1
  fi

  while read -r duplicate; do
    err "samples share a format and scenarios: $duplicate"
  done < <(jq -r '
    [.files[] | select(.role == "sample")
      | {key: ([.format] + ((.covers // []) | sort) | join(" ")), path}]
    | group_by(.key)[] | select(length > 1) | map(.path) | join(", ")' MANIFEST.json)
fi

coverage=${COVERAGE:-docs/coverage.json}
if [ -f "$coverage" ]; then
  if problems=$(jq -r '
    def where: map(if type == "number" then "[\(.)]" else ".\(.)" end) | join("");
    .vocabulary as $v
    | ([$v.library, $v.trait, $v.feature] | map(keys) | add) as $ids
    | ($v.library | to_entries[]
        | select((.key | test("^(kind|route|check|capability):.")) | not)
        | "vocabulary.library: \(.key) is not kind:, route:, check:, or capability: and a value"),
      ($v.library | to_entries[]
        | select((.value.identifier | type) != "string"
            or ((.value.link // "") | test("^https://github[.]com/standards-lab/(docproc|docinfer)/blob/[0-9a-f]{40}/[^#]+#L[0-9]+$") | not))
        | "vocabulary.library: \(.key) needs its identifier and a link pinned to a docproc or docinfer commit, file, and line"),
      ($v.trait | to_entries[]
        | select((.key | test("^trait:[a-z-]+[.][a-z0-9-]+$")) | not)
        | "vocabulary.trait: \(.key) is not trait:<trait>.<value>"),
      ($v.feature | to_entries[]
        | select((.key | test("^feature:[a-z0-9-]+$")) | not)
        | "vocabulary.feature: \(.key) is not feature:<name>"),
      ([$v.trait, $v.feature] | add | to_entries[]
        | select((.value.rule | type) != "string" or .value.rule == "")
        | "vocabulary: \(.key) needs a one-line rule"),
      ($v.library | to_entries[] | .key as $k | (.value.fed_by.routes // [])[]
        | select((startswith("route:") and $v.library[.] != null) | not)
        | "vocabulary.library: \($k).fed_by.routes: \(.) is not a route: ID in the vocabulary; name one it defines"),
      (.samples // [] | to_entries[]
        | select((.value.path | type) != "string" or ((.value.sha256 // "") | test("^[0-9a-f]{64}$") | not))
        | "samples[\(.key)]: needs a path and a sha256 of 64 hex digits; copy both from its MANIFEST.json entry"),
      (.samples // [] | to_entries[] | .key as $i | (.value.library // [])[]
        | select($v.library[.] == null)
        | "samples[\($i)].library: \(.) is not a library ID in the vocabulary; rerun scripts/place-library.sh, or define the ID there with its link"),
      (.samples // [] | group_by(.path)[] | select(length > 1)
        | "samples: \(.[0].path) appears \(length) times; keep one entry per path"),
      (path(.. | objects | select(has("covers"))) as $p
        | getpath($p).covers[]
        | select(IN($ids[]) | not)
        | "\($p | where).covers: \(.) is not in the vocabulary; use an ID the vocabulary defines, or define it there with its link or rule")
  ' "$coverage"); then
    while read -r problem; do
      if [ -n "$problem" ]; then
        err "$coverage: $problem"
      fi
    done <<<"$problems"
  else
    err "$coverage: jq could not read it; it must be JSON with a vocabulary of library, trait, and feature objects"
  fi

  reference=${DOCS_REFERENCE:-$HOME/docs}/MANIFEST.json
  if [ -f "$reference" ]; then
    if problems=$(jq -r --slurpfile reference "$reference" '
      (.samples // [] | map({key: .path, value: .sha256}) | from_entries) as $listed
      | ($reference[0].files | map(select(.role == "sample") | {key: .path, value: .sha256}) | from_entries) as $expected
      | ($expected | to_entries[]
          | if $listed[.key] == null then
              "samples: \(.key) is a sample in \($ref) but missing; add it, as scripts/place-library.sh prints it"
            elif $listed[.key] != .value then
              "samples: \(.key) has sha256 \($listed[.key]), but \($ref) has \(.value); rerun scripts/place-library.sh and take its sha256"
            else empty end),
        ($listed | keys[] | select($expected[.] == null)
          | "samples: \(.) is not a sample in \($ref); remove its entry")
    ' --arg ref "$reference" "$coverage"); then
      while read -r problem; do
        if [ -n "$problem" ]; then
          err "$coverage: $problem"
        fi
      done <<<"$problems"
    else
      err "$coverage: jq could not compare it with $reference; both must be JSON, the reference with a files array"
    fi
  fi
fi

run_lint() {
  if ! command -v "$1" >/dev/null; then
    echo "FAIL: $1 is not on PATH; run mise install" >&2
    return 1
  fi
  "$1"
}

run_lint lint-context || fail=1
run_lint lint-references || fail=1

if [ "$fail" -ne 0 ]; then
  exit 1
fi
echo "All checks passed."
