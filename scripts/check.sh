#!/usr/bin/env bash
# Checks the tools/coverage Go module, then the corpus files against MANIFEST.json once it is
# present, the coverage data, and MANIFEST.json against the coverage data and git, and runs
# lint-context and lint-references. Writes nothing but Go's caches.
# See docs/coverage.md
# See architecture/standards/go-elemental/principles/release-and-ci.md
set -euo pipefail

cd "$(dirname "$0")/.."

go_modules=(tools/coverage)

for tool in go gofmt golangci-lint; do
  if ! command -v "$tool" >/dev/null; then
    echo "FAIL: $tool is not on PATH; run mise install" >&2
    exit 1
  fi
done

for module in "${go_modules[@]}"; do
  echo "check: $module"
  (
    cd "$module"
    export GOWORK=off
    go build ./...
    go vet ./...
    unformatted=$(gofmt -l .)
    if [ -n "$unformatted" ]; then
      echo "FAIL: $module: gofmt would reformat:" >&2
      echo "$unformatted" >&2
      exit 1
    fi
    go fix -diff ./...
    go mod tidy -diff
    go test -race ./...
    golangci-lint run ./...
  )
done

fail=0
err() {
  echo "FAIL: $*" >&2
  fail=1
}

areas=(archive audio email exec image office onenote pdf text video)

if [ -f MANIFEST.json ]; then
  while IFS=$'\t' read -r path role bytes sha gap; do
    if [ "$role" = sample ] && [ -n "$gap" ]; then
      payload="mise run fetch -- <release-url|dir>, or mise run fetch-candidates"
    elif [ "$role" = sample ]; then
      payload="mise run fetch -- <release-url|dir>, or mise run manifest --copy > MANIFEST.json"
    else
      payload="git checkout -- $path"
    fi
    if [ ! -f "$path" ]; then
      err "$path: in MANIFEST.json but missing; restore it with $payload"
      continue
    fi
    [ "$(stat -c %s "$path")" = "$bytes" ] || err "$path: size differs from MANIFEST.json; restore it with $payload"
    [ "$(sha256sum "$path" | cut -d' ' -f1)" = "$sha" ] || err "$path: sha256 differs from MANIFEST.json; restore it with $payload"
  done < <(jq -r '.files[] | [.path, .role, (.bytes|tostring), .sha256, (.gap // "")] | @tsv' MANIFEST.json)

  existing_areas=()
  for area in "${areas[@]}"; do
    [ -d "$area" ] && existing_areas+=("$area")
  done
  if [ "${#existing_areas[@]}" -gt 0 ]; then
    comm -23 \
      <(find "${existing_areas[@]}" -type f | sort) \
      <(jq -r '.files[].path' MANIFEST.json | sort) |
      while read -r path; do
        echo "FAIL: $path: not in MANIFEST.json; remove it, or adopt it in docs/coverage.json and regenerate with mise run manifest > MANIFEST.json" >&2
        echo x
      done | grep -q x && fail=1
  fi

  while read -r duplicate; do
    err "samples share a format and scenarios: $duplicate; trim all but one with mise run dispose, then mise run manifest > MANIFEST.json"
  done < <(jq -r '
    [.files[] | select(.role == "sample")
      | {key: ([.format] + ((.covers // []) | sort) | join(" ")), path}]
    | group_by(.key)[] | select(length > 1) | map(.path) | join(", ")' MANIFEST.json)
fi

stop_if_failed() {
  if [ "$fail" -ne 0 ]; then
    exit 1
  fi
}

stop_if_failed

report() {
  local problem
  while read -r problem; do
    if [ -n "$problem" ]; then
      err "$coverage: $problem"
    fi
  done <<<"$1"
}

coverage=${COVERAGE:-docs/coverage.json}
if [ ! -f "$coverage" ]; then
  err "$coverage: missing; restore docs/coverage.json from git, or point COVERAGE at the coverage data"
else
  if problems=$(jq -r '
    def where: map(if type == "number" then "[\(.)]" else ".\(.)" end) | join("");
    .vocabulary as $v
    | ([$v.library, $v.trait, $v.feature] | map(keys) | add) as $ids
    | (select((.samples | type) != "array")
        | "samples: is not an array; list every ~/docs sample there, as scripts/place-library.sh prints them"),
      (select((.gaps | type) != "array")
        | "gaps: is not an array; list the gaps there, or [] when there are none"),
      ($v.library | to_entries[]
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
        | "samples[\($i)].\($field): \(.) is not a \($axis) ID in the vocabulary; run mise run place-traits again, or define the ID there with its rule"),
      (.samples // [] | to_entries[] | .key as $i
        | (.value.judged // {}) | keys[]
        | select(IN($ids[]) | not)
        | "samples[\($i)].judged: \(.) is not in the vocabulary; judge an ID the vocabulary defines")
  ' "$coverage"); then
    report "$problems"
  else
    err "$coverage: jq could not read it; it must be JSON with a vocabulary of library, trait, and feature objects"
  fi

  if problems=$(jq -r '
    def covers: (.library + (.traits // []) + (.features // [])) | unique;
    def within($outer): (covers - ($outer | covers)) == [];
    def rank: [(.truth // []) == [], .license_status != "verified", .redistributable != true, .bytes, .path];
    ($placed[0] | map({key: .path, value: .covers}) | from_entries) as $covers
    | ($fetched[0] | map({key: .gap, value: .}) | from_entries) as $files
    | ([(.gaps // [])[] | select(.candidate.kind == "generated") | .candidate.inputs[]?]) as $inputs
    | (.samples // []) as $samples
    | ($samples | map({key: .path, value: .}) | from_entries) as $entries
    | ([(.gaps // [])[] | select(.candidate.kind? == "sourced" and .covered_by == null)
        | .candidate.license_status as $status | $files[.id] // empty | select($covers[.path] != null)
        | {key: .path, value: {license_status: $status, redistributable: ($status == "verified"), truth: []}}]
        | from_entries) as $sourced
    | ($entries + $sourced) as $readable
    | ([(.gaps // [])[] | select(IN(.candidate.kind?; "sourced", "generated") and .covered_by == null)
        | .candidate as $c | $files[.id] // empty | select($covers[.path] != null)
        | (if $c.kind == "generated" then
             [$c.inputs[]? | $readable[.]] as $in
             | {license_status: (first($in[] | (.license_status? // "") | select(. != "verified")) // "verified"),
                redistributable: ($in != [] and all($in[]; .redistributable? == true)),
                truth: (if .truth_from != null and IN(.truth_from; $c.inputs[]?) then ($readable[.truth_from].truth // []) else [] end)}
           else
             {license_status: $c.license_status, redistributable: ($c.license_status == "verified"), truth: []}
           end) as $facts
        | {path, gap, format, bytes, library: $covers[.path], disposition: "adopt"} + $facts]) as $candidates
    | ($samples | map(select(.disposition == "adopt")) + $candidates) as $adopted
    | ($samples + $candidates | map({key: .path, value: .}) | from_entries) as $by
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
        | select(.path != $a.path and within($a) and (covers != ($a | covers)) and (IN(.path; $inputs[]) | not))
        | if has("gap") then "\(.path): gap \(.gap)\u0027s candidate, its covers are inside \($a.path)\u0027s; run mise run dispose, which records the gap covered_by \($a.path)"
          else "\(.path): its covers are inside \($a.path)\u0027s; trim it for \($a.path)" end),
      ($adopted[] as $a | $adopted[]
        | select(.path != $a.path and .format != $a.format and covers == ($a | covers) and rank > ($a | rank) and (IN(.path; $inputs[]) | not))
        | "\(.path): its covers equal \($a.path)\u0027s, which wins the tie-break; \(if has("gap") then "run mise run dispose, which records gap \(.gap) covered_by it" else "trim it for \($a.path)" end)"),
      ((.gaps // [])[] | select(.covered_by != null) | .covered_by as $by_path
        | if $by[$by_path].disposition != "adopt" then
            "gaps.\(.id).covered_by: \($by_path) is neither an adopted sample nor a placed candidate; run mise run dispose"
          elif (.covers - ($by[$by_path] | covers)) != [] then
            "gaps.\(.id).covered_by: \($by_path)\u0027s covers lack \(.covers - ($by[$by_path] | covers) | join(", ")); run mise run dispose"
          else empty end)
  ' --slurpfile fetched docs/candidates.json --slurpfile placed docs/corpus-placement.json "$coverage"); then
    report "$problems"
  else
    err "$coverage: jq could not check its dispositions; each sample needs library, traits, features, and disposition fields, and docs/candidates.json and docs/corpus-placement.json must be JSON arrays"
  fi

  if problems=$(jq -r '
    def text: type == "string" and . != "";
    def github: test("^https://([a-z]+[.])?github(usercontent)?[.]com/");
    ((.samples // []) | map(select(.disposition == "adopt") | .path)) as $adopted
    | ($fetched[0] | map({key: .gap, value: .path}) | from_entries) as $fetched_paths
    | [(.gaps // [])[] | select(.candidate.kind? == "sourced" and .covered_by == null) | $fetched_paths[.id] // empty] as $sourced
    | ((.gaps // []) | map(.covers // []) | add // []) as $gapped
    | ((.samples // []) | map(select(.disposition == "adopt") | (.library + (.traits // []) + (.features // []))[]) | unique) as $exercised
    | ((.gaps // []) | group_by(.id)[] | select(length > 1)
        | "gaps: \(.[0].id) appears \(length) times; give each gap its own id"),
      ((.gaps // [])[] | .id as $id | .candidate as $c
        | if ($c | type) != "object" then
            "gaps.\($id): has no candidate; name a sourced one (url, ref, license, license_url, license_read, license_status \"verified\") or a generated one (command, inputs, generator, generator_license), or record it unavailable (searched, reason)"
          else
            (select(($c.kind == "sourced" or $c.kind == "generated") and (($c.fit | text) | not))
              | "gaps.\($id).candidate: needs a fit line; say its quality and how it exercises the gap\u0027s covers"),
            (if $c.kind == "sourced" then
              (select(($c.url | text) and ($c.url | startswith("https://")) | not)
                | "gaps.\($id).candidate: needs its https url; record where the file is fetched"),
              (select(($c.ref | text) | not)
                | "gaps.\($id).candidate: needs its ref; record the commit or revision the url is pinned to"),
              (select(($c.url | text) and ($c.url | github) and ((($c.ref // "") | test("^[0-9a-f]{40}$")) and ($c.url | contains($c.ref)) | not))
                | "gaps.\($id).candidate: a GitHub url must be pinned to a commit; put the 40-hex commit in the url and in ref"),
              (select(($c.license | text) | not)
                | "gaps.\($id).candidate: needs its license; record the license read at its source"),
              (select(($c.license_url | text) and ($c.license_url | startswith("https://")) | not)
                | "gaps.\($id).candidate: needs its license_url; record the https URL of the license you read"),
              (select(($c.license_read // "") | test("^[0-9]{4}-[0-9]{2}-[0-9]{2}$") | not)
                | "gaps.\($id).candidate: needs license_read; record the date you read the license as YYYY-MM-DD"),
              (select($c.license_status != "verified")
                | "gaps.\($id).candidate: license_status is \($c.license_status | tojson); read the license at its source and set \"verified\", or name another candidate")
            elif $c.kind == "generated" then
              (select(($c.command | text) | not)
                | "gaps.\($id).candidate: has no command; record the exact command that makes it"),
              (select(($c.generator | text) | not)
                | "gaps.\($id).candidate: needs its generator; record the tool and its version"),
              (select(($c.generator_license | text) | not)
                | "gaps.\($id).candidate: needs its generator_license; record the generator\u0027s license"),
              (if ($c.inputs | type) != "array" or $c.inputs == [] then
                "gaps.\($id).candidate: needs its inputs; list the adopted samples and sourced candidates the command reads"
              else
                $c.inputs[] | select(IN($adopted[], $sourced[]) | not)
                | "gaps.\($id).candidate.inputs: \(.) is neither an adopted sample nor a sourced candidate in docs/candidates.json; generate only from those"
              end)
            elif $c.kind == "unavailable" then
              (select(($c.searched | text) | not)
                | "gaps.\($id).candidate: an unavailable candidate needs searched; record the suites, files, and searches you tried"),
              (select(($c.reason | text) | not)
                | "gaps.\($id).candidate: an unavailable candidate needs reason; say why no sourced or generated file can exercise the covers")
            else
              "gaps.\($id).candidate: kind is \($c.kind | tojson); set kind to \"sourced\", \"generated\", or \"unavailable\""
            end)
          end),
      ([.vocabulary.library, .vocabulary.trait, .vocabulary.feature] | map(keys) | add[]
        | select(IN($exercised[]) | not) | select(IN($gapped[]) | not)
        | "vocabulary: \(.) is exercised by no adopted sample and is in no gap\u0027s covers; adopt a sample that exercises it, or add it to a gap\u0027s covers")
  ' --slurpfile fetched docs/candidates.json "$coverage"); then
    report "$problems"
  else
    err "$coverage: jq could not check its gaps; each gap needs an id, covers, and a candidate, and docs/candidates.json must be a JSON array"
  fi

  reference=${DOCS_REFERENCE:-$HOME/docs}/MANIFEST.json
  if [ -n "${DOCS_REFERENCE:-}" ] && [ ! -f "$reference" ]; then
    err "$reference: missing; point DOCS_REFERENCE at a directory with MANIFEST.json, or unset it"
  elif [ -f "$reference" ]; then
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
      report "$problems"
    else
      err "$coverage: jq could not compare it with $reference; both must be JSON, the reference with a files array"
    fi
  fi
fi

stop_if_failed

if [ -f MANIFEST.json ]; then
  regenerate="regenerate it with mise run manifest > MANIFEST.json"
  if ! tracked=$(git ls-files -z | jq -Rs 'split("\u0000") | map(select(. != ""))'); then
    err "git ls-files failed; run the check in a git checkout of docs-corpus"
  elif problems=$(jq -r --argjson tracked "$tracked" --slurpfile coverage "$coverage" --arg regenerate "$regenerate" '
    def covers: (.library + (.traits // []) + (.features // [])) | unique;
    ($coverage[0].vocabulary | [.library, .trait, .feature] | map(keys) | add) as $ids
    | ($coverage[0].samples | map(select(.disposition == "adopt") | {key: .path, value: covers}) | from_entries) as $adopted
    | ($coverage[0].gaps | map({key: .id, value: .}) | from_entries) as $gaps
    | ([$coverage[0].gaps[] | select(.candidate.kind == "generated") | .candidate.inputs[]?]) as $inputs
    | ($tracked | map({key: ., value: true}) | from_entries) as $in_git
    | [.files[] | select(.role == "sample")] as $samples
    | (.files | map({key: .path, value: .}) | from_entries) as $by_path
    | (select(.version != 1)
        | "MANIFEST.json: version is \(.version | tojson), not 1; \($regenerate)"),
      (.files[] | select(IN(.role; "sample", "truth", "doc") | not)
        | "\(.path): role is \(.role | tojson); set sample, truth, or doc, and \($regenerate)"),
      (.files[] | select(.role != "sample" and .license_status == "reported")
        | "\(.path): a \(.role) file whose license is only reported; read it at its source, add the reading to docs/license-readings.json, and \($regenerate)"),
      ($samples[] | select((has("gap") | not) and ((.license_url // "") == "" or (.license_read // "") == ""))
        | "\(.path): an adopted sample without license_url and license_read; read its license at its source, add the reading to docs/license-readings.json, and \($regenerate)"),
      (.files[] | select(.role != "sample" and ((.license_url // "") == "" or (.license_read // "") == ""))
        | "\(.path): a \(.role) file without license_url and license_read; read its license at its source, add the reading to docs/license-readings.json, and \($regenerate)"),
      (.files[] | .path as $p | (.covers // [])[] | select(IN($ids[]) | not)
        | "\($p).covers: \(.) is not in docs/coverage.json\u0027s vocabulary; place the sample with an ID the vocabulary defines, or define it there, then \($regenerate)"),
      ($samples[] | select($in_git[.path])
        | "\(.path): a sample file is tracked by git, but samples are release assets; run git rm --cached -- \(.path | @sh)"),
      (.files[] | select(.role != "sample" and ($in_git[.path] | not))
        | "\(.path): a \(.role) file is not tracked by git; run git add -f -- \(.path | @sh)"),
      ($samples[] | select(has("gap") | not) | select($adopted[.path] == null)
        | "\(.path): a sample that is not adopted in docs/coverage.json; \($regenerate)"),
      ($samples | map(.path)) as $listed
      | ($adopted | keys[] | select(IN($listed[]) | not)
        | "\(.): adopted in docs/coverage.json but not a sample in MANIFEST.json; \($regenerate)"),
      ($samples[] | select($adopted[.path] != null and ((.covers // []) | unique) != $adopted[.path])
        | "\(.path): covers differ from docs/coverage.json\u0027s; \($regenerate)"),
      ($samples[] as $outer | $samples[]
        | select(.path != $outer.path and (((.covers // []) - ($outer.covers // [])) == [])
            and ((.format == $outer.format and ((.covers // []) | unique) == (($outer.covers // []) | unique)) | not)
            and (IN(.path; $inputs[]) | not))
        | "\(.path): its covers lie inside \($outer.path)\u0027s; a sample stays only for a scenario no other covers, so trim it for \($outer.path) by mise run dispose, then \($regenerate)"),
      ($samples[] | select(has("gap")) | .gap as $id | $gaps[$id] as $g
        | if ($g.candidate.kind? | IN("sourced", "generated")) | not then
            "\(.path): a candidate for gap \($id), which has no sourced or generated candidate in docs/coverage.json; \($regenerate)"
          elif $g.covered_by != null then
            "\(.path): gap \($id) is covered_by \($g.covered_by), so its candidate leaves the corpus; \($regenerate)"
          else
            (. as $f | ["url", "ref", "license_url", "license_read", "command", "inputs", "generator", "generator_license"]
              + (if $g.candidate.kind == "sourced" then ["license", "license_status"] else [] end)
              | .[] | select(($f[.] // null) != ($g.candidate[.] // null))
              | "\($f.path).\(.): differs from gap \($id)\u0027s candidate; never edit a candidate\u0027s provenance here, \($regenerate)"),
            (select($g.candidate.kind == "sourced" and .truth != [])
              | "\(.path): a sourced candidate inherits no truth; \($regenerate)"),
            (select($g.candidate.kind == "generated") | . as $f | [$g.candidate.inputs[]? | $by_path[.] // {}] as $in
              | ({license: ([$in[].license] | reduce .[] as $l ([]; if any(.[]; . == $l) then . else . + [$l] end) | join("; ")),
                  license_status: (first($in[] | (.license_status // "") | select(. != "verified")) // "verified"),
                  redistributable: ($in != [] and all($in[]; .redistributable == true))}
                | to_entries[] | select($f[.key] != .value)
                | "\($f.path).\(.key): \($f[.key] | tojson), but its inputs give \(.value | tojson); a generated candidate\u0027s license follows from its inputs\u0027, so \($regenerate)"),
              (($f.truth // []) - [$in[].truth[]?] | select(. != [])
                | "\($f.path): its truth \(map(.path) | join(", ")) is not its inputs\u0027 truth; a generated candidate inherits only an input\u0027s truth, unchanged, so fix truth_from in docs/candidates.json and \($regenerate)")),
            (($g.covers - (.covers // [])) as $missing | select($missing != [])
              | "\(.path): its placed covers lack gap \($id)\u0027s \($missing | join(", ")); run docproc over the corpus and mise run place-traits --corpus > docs/corpus-placement.json as docs/coverage.md states, then mise run dispose to split the covers it does not confirm into an open gap, and \($regenerate)")
          end),
      ([$samples[] | {path, covers}] | sort_by(.path)) as $rows
      | (($placed[0] - $rows) + ($rows - $placed[0]) | map(.path) | unique | select(. != [])
        | "docs/corpus-placement.json: its rows for \(join(", ")) differ from MANIFEST.json\u0027s samples\u0027 paths and covers; regenerate it with mise run place-traits --corpus > docs/corpus-placement.json as docs/coverage.md states, then mise run dispose > docs/coverage.json and mise run manifest > MANIFEST.json"),
      ([$samples[].gap] as $filled | $coverage[0].gaps[]
        | select(IN(.candidate.kind; "sourced", "generated") and .covered_by == null and (IN(.id; $filled[]) | not))
        | "gaps.\(.id): its \(.candidate.kind) candidate is not a sample in MANIFEST.json; run mise run \(if .candidate.kind == "sourced" then "fetch-candidates" else "generate-candidates" end), then \($regenerate)"),
      ($judged[0] | to_entries[] | .key as $p
        | if any($samples[]; .path == $p) | not then
            "docs/judged.json: \($p) is not a sample in MANIFEST.json; judge only the corpus\u0027s samples"
          else
            .value | keys[] | select(IN($ids[]) | not)
            | "docs/judged.json: \($p): \(.) is not in docs/coverage.json\u0027s vocabulary; judge an ID the vocabulary defines"
          end)
  ' --slurpfile judged docs/judged.json --slurpfile placed docs/corpus-placement.json MANIFEST.json); then
    while read -r problem; do
      if [ -n "$problem" ]; then
        err "$problem"
      fi
    done <<<"$problems"
  else
    err "MANIFEST.json: jq could not check it against $coverage; $regenerate"
  fi

  if [ -f "${DOCS_REFERENCE:-$HOME/docs}/MANIFEST.json" ] && ! go run ./tools/coverage/cmd/manifest | cmp -s - MANIFEST.json; then
    err "MANIFEST.json: differs from what mise run manifest prints, and it is never edited by hand; $regenerate"
  fi
fi

stop_if_failed

run_lint() {
  if ! command -v "$1" >/dev/null; then
    echo "FAIL: $1 is not on PATH; run mise install" >&2
    return 1
  fi
  "$1"
}

run_lint lint-context || fail=1
run_lint lint-references || fail=1

stop_if_failed
echo "All checks passed."
