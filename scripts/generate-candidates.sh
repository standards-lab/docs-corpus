#!/usr/bin/env bash
# Generates every generated gap candidate in docs/coverage.json, or the copy COVERAGE names, that
# docs/candidates.json doesn't record yet, but those whose gap is covered_by another sample, by
# running its recorded command once from the repository root with bash -o pipefail. Each input
# must match its size and sha256 in MANIFEST.json. The directory of each corpus path the command
# names is created first, and the command must write exactly one new file under the format
# directories, which is recorded with its path, size, sha256, and format as file -b --mime-type
# reports it, with file's version. Nothing regenerates a recorded candidate. If any command fails,
# the files this run made are removed and nothing is recorded.
# Usage: scripts/generate-candidates.sh
# See docs/coverage.md
set -euo pipefail

if [ "$#" -ne 0 ]; then
  echo "FAIL: usage: mise run generate-candidates" >&2
  exit 2
fi

cd "$(dirname "$0")/.."

for tool in file jq sha256sum; do
  if ! command -v "$tool" >/dev/null; then
    echo "FAIL: $tool is not on PATH; run mise install, or install it on the host" >&2
    exit 1
  fi
done

coverage=${COVERAGE:-docs/coverage.json}
record=docs/candidates.json
areas=(archive audio email exec image office onenote pdf text video)
recorded='[]'
if [ -f "$record" ]; then
  recorded=$(jq -c . "$record")
fi
file_version=$(file --version | head -n 1)

listing() {
  local existing=()
  for area in "${areas[@]}"; do
    [ -d "$area" ] && existing+=("$area")
  done
  find "${existing[@]}" -type f | sort
}

made=()
entries=()
fail() {
  echo "FAIL: $*" >&2
  if [ "${#made[@]}" -gt 0 ]; then
    rm -f -- "${made[@]}"
  fi
  echo "FAIL: nothing was recorded, and the files this run made are removed" >&2
  exit 1
}

while IFS=$'\t' read -r gap command; do
  if jq -e --arg gap "$gap" 'any(.[]; .gap == $gap)' <<<"$recorded" >/dev/null; then
    continue
  fi
  while IFS=$'\t' read -r input bytes sha; do
    if [ -z "$bytes" ]; then
      fail "gaps.$gap.candidate.inputs: $input is not a sample in MANIFEST.json; regenerate it with mise run manifest > MANIFEST.json"
    fi
    if [ ! -f "$input" ] || [ "$(stat -c %s "$input")" != "$bytes" ] || [ "$(sha256sum "$input" | cut -d' ' -f1)" != "$sha" ]; then
      fail "gaps.$gap.candidate.inputs: $input is missing or differs from MANIFEST.json; restore it with mise run fetch -- <release-url|dir>"
    fi
  done < <(jq -r --arg gap "$gap" --slurpfile manifest MANIFEST.json '
    ($manifest[0].files | map({key: .path, value: .}) | from_entries) as $files
    | .gaps[] | select(.id == $gap) | .candidate.inputs[]
    | [., ($files[.].bytes // ""), ($files[.].sha256 // "")] | @tsv' "$coverage")
  set -f
  for word in $command; do
    word=${word//\'/}
    word=${word//\"/}
    for area in "${areas[@]}"; do
      if [[ $word == "$area"/*/* ]]; then
        mkdir -p "$(dirname "$word")"
      fi
    done
  done
  set +f
  echo "generate-candidates: $gap: $command"
  before=$(listing)
  status=0
  bash -o pipefail -c "$command" </dev/null || status=$?
  mapfile -t new < <(comm -13 <(echo "$before") <(listing))
  made+=("${new[@]}")
  if [ "$status" -ne 0 ]; then
    fail "gaps.$gap.candidate.command: failed; a recorded command is never rewritten, so name another candidate in its gap"
  fi
  if [ "${#new[@]}" -ne 1 ]; then
    fail "gaps.$gap.candidate.command: wrote ${#new[@]} new files under the format directories, not one: ${new[*]}"
  fi
  path=${new[0]}
  entries+=("$(jq -nc --arg gap "$gap" --arg path "$path" --arg sha "$(sha256sum "$path" | cut -d' ' -f1)" \
    --argjson bytes "$(stat -c %s "$path")" --arg format "$(file -b --mime-type "$path")" --arg by "$file_version" \
    '{gap: $gap, path: $path, source: "generated", sha256: $sha, bytes: $bytes, format: $format, format_by: $by}')")
done < <(jq -r '.gaps[] | select(.candidate.kind == "generated" and .covered_by == null) | [.id, .candidate.command] | @tsv' "$coverage")

if [ "${#entries[@]}" -eq 0 ]; then
  echo "generate-candidates: every generated candidate is recorded in $record"
  exit 0
fi
printf '%s\n' "${entries[@]}" | jq -s --argjson recorded "$recorded" '$recorded + . | sort_by(.path)' >"$record.tmp"
mv -f "$record.tmp" "$record"
echo "generate-candidates: generated ${#entries[@]} candidates and wrote $record"
