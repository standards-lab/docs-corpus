#!/usr/bin/env bash
# Checks the corpus once MANIFEST.json is present: every manifest entry names a file of its size
# and sha256, every corpus file has an entry, and no two samples share a format and a `covers`
# set. Checks the coverage data once it is present, at $COVERAGE or docs/coverage.json: its
# vocabulary's IDs and entries have their axis's shape, every `covers`, trait, feature, and judged
# ID in it is in the vocabulary, every capability's fed_by route and every sample's library ID is a
# library ID, and no sample path appears twice. Every sample has its tie-break facts and exactly
# one disposition: a trim names an adopted sample whose covers contain its own and that the
# tie-break prefers, a replace names a gap whose covers contain its own, no adopted sample is
# non-redistributable or sensitive, and no adopted sample's covers equal, with its format, or lie
# inside another's. When ${DOCS_REFERENCE:-$HOME/docs}/MANIFEST.json exists, every sample in it
# appears in the coverage once, by path and sha256, with the manifest's format, bytes,
# license_status, redistributable, and sensitivity, the coverage lists no other, and every truth
# and doc file in it is linked. Then runs lint-context and lint-references, and fails if any part
# fails. Reads no network.
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
        | "\($p | where).covers: \(.) is not in the vocabulary; use an ID the vocabulary defines, or define it there with its link or rule"),
      (.samples // [] | to_entries[] | .key as $i | .value as $s
        | (["traits", "trait"], ["features", "feature"])
        | .[0] as $field | .[1] as $axis
        | ($s[$field] // [])[]
        | select($v[$axis][.] == null)
        | "samples[\($i)].\($field): \(.) is not a \($axis) ID in the vocabulary; rerun scripts/place-traits.py, or define the ID there with its rule"),
      (.samples // [] | to_entries[] | .key as $i
        | (.value.judged // {}) | keys[]
        | select(IN($ids[]) | not)
        | "samples[\($i)].judged: \(.) is not in the vocabulary; judge an ID the vocabulary defines")
  ' "$coverage"); then
    while read -r problem; do
      if [ -n "$problem" ]; then
        err "$coverage: $problem"
      fi
    done <<<"$problems"
  else
    err "$coverage: jq could not read it; it must be JSON with a vocabulary of library, trait, and feature objects"
  fi

  if problems=$(jq -r '
    def covers: (.library + (.traits // []) + (.features // [])) | unique;
    def within($outer): (covers - ($outer | covers)) == [];
    def rank: [(.truth // []) == [], .license_status != "verified", .redistributable != true, .bytes, .path];
    (.samples // []) as $samples
    | ($samples | map(select(.disposition == "adopt"))) as $adopted
    | ($samples | map({key: .path, value: .}) | from_entries) as $by
    | ((.gaps // []) | map({key: .id, value: .}) | from_entries) as $gaps
    | ($samples[]
        | select((.format | type) != "string" or (.bytes | type) != "number" or (.license_status | type) != "string"
            or (.redistributable | type) != "boolean" or (.truth | type) != "array" or (.docs | type) != "array")
        | "\(.path): needs format, bytes, license_status, redistributable, truth, and docs; copy them from its MANIFEST.json entry"),
      ($samples[]
        | if .disposition == "adopt" then
            select(has("trimmed_for") or has("gap"))
            | "\(.path): an adopted sample names no trimmed_for or gap; remove them"
          elif .disposition == "trim" then
            .path as $path | covers as $c | rank as $r | .trimmed_for as $for
            | if has("gap") then "\($path): a trim names no gap; remove it"
              elif $by[$for // ""].disposition != "adopt" then
                "\($path): trimmed_for \($for) is not an adopted sample; name the adopted sample that covers its scenarios"
              elif (within($by[$for]) | not) then
                "\($path): \($for)\u0027s covers lack \($c - ($by[$for] | covers) | join(", ")); name an adopted sample whose covers contain this one\u0027s, or replace it"
              else
                ([$adopted[] | select(.path != $for and (. as $a | $by[$path] | within($a)) and rank < ($by[$for] | rank))][0].path // empty)
                | "\($path): \(.) also covers its scenarios and wins the tie-break over \($for); name \(.)"
              end
          elif .disposition == "replace" then
            if has("trimmed_for") then "\(.path): a replace names no trimmed_for; remove it"
            elif $gaps[.gap // ""] == null then "\(.path): gap \(.gap) is not in .gaps; add it there with the covers it must exercise"
            elif (covers - $gaps[.gap].covers) != [] then
              "\(.path): gap \(.gap)\u0027s covers lack \(covers - $gaps[.gap].covers | join(", ")); add them to the gap or name a gap that has them"
            else empty end
          else
            "\(.path): has no disposition; set disposition to adopt, trim with trimmed_for, or replace with gap"
          end),
      ($adopted[]
        | select(.redistributable != true or .sensitive == true)
        | "\(.path): is adopted but \(if .sensitive == true then "sensitive" else "not redistributable" end); replace it with a gap, or trim it for an adopted sample that covers it"),
      ($adopted | group_by([.format, covers])[] | select(length > 1)
        | "\(map(.path) | join(", ")): adopted samples share a format and covers; keep the one the tie-break prefers and trim the rest for it"),
      ($adopted[] as $a | $adopted[]
        | select(.path != $a.path and within($a) and (covers != ($a | covers)))
        | "\(.path): its covers are inside \($a.path)\u0027s; trim it for \($a.path)"),
      ($adopted[] as $a | $adopted[]
        | select(.path != $a.path and .format != $a.format and covers == ($a | covers) and rank > ($a | rank))
        | "\(.path): its covers equal \($a.path)\u0027s, which wins the tie-break; trim it for \($a.path)")
  ' "$coverage"); then
    while read -r problem; do
      if [ -n "$problem" ]; then
        err "$coverage: $problem"
      fi
    done <<<"$problems"
  else
    err "$coverage: jq could not check its dispositions; each sample needs library, traits, features, and disposition fields"
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
          | "samples: \(.) is not a sample in \($ref); remove its entry"),
        (($reference[0].files | map({key: .path, value: .}) | from_entries) as $files
        | (.samples // [])[] as $s | $files[$s.path] // empty
        | (. as $e | ["format", "bytes", "license_status", "redistributable"][]
            | select($s[.] != $e[.])
            | "\($s.path).\(.): \($s[.] | tojson) differs from \($ref)\u0027s \($e[.] | tojson); copy it from there"),
          (select((.sensitive != null or .local_only != null) != ($s.sensitive == true))
            | "\($s.path).sensitive: \($ref) marks it \(if .sensitive != null or .local_only != null then "sensitive or local_only; set sensitive to true" else "neither sensitive nor local_only; remove sensitive" end)"),
          (($s.truth + $s.docs)[] | select($files[.] == null)
            | "\($s.path): \(.) is not a file in \($ref); link only its truth and doc files")),
      ([(.samples // [])[] | .truth[], .docs[]] + [(.docs_without_sample // [])[].path]) as $linked
      | ($reference[0].files[] | select(.role == "truth" or .role == "doc") | .path
          | select(IN($linked[]) | not)
          | "\(.): a \($ref) truth or doc file no sample links; add it to its sample\u0027s truth or docs, or to docs_without_sample with its handling")
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
