#!/usr/bin/env bash
# Fetches every sourced gap candidate in docs/coverage.json, or the copy COVERAGE names, but those
# whose gap is covered_by another sample, from its pinned URL and places it at
# <area>/<source>/<name>: the area is the gap id's prefix, the source the GitHub repository's name
# in lower case, or wikimedia-commons, and the name the URL's last segment, decoded, with a
# trailing .txt dropped from a name that already has an extension. Each fetched file must match
# the path, size, and sha256 docs/candidates.json records for its gap; a gap it doesn't record yet
# is recorded with its file's size, sha256, and format as file -b --mime-type reports it, with
# file's version. Nothing is placed or recorded unless every
# candidate is fetched and matches, and the record keeps the entries of the gaps it skips.
# Usage: scripts/fetch-candidates.sh
# See docs/coverage.md
set -euo pipefail

if [ "$#" -ne 0 ]; then
  echo "FAIL: usage: mise run fetch-candidates" >&2
  exit 2
fi

cd "$(dirname "$0")/.."

for tool in curl file jq sha256sum; do
  if ! command -v "$tool" >/dev/null; then
    echo "FAIL: $tool is not on PATH; run mise install, or install it on the host" >&2
    exit 1
  fi
done

coverage=${COVERAGE:-docs/coverage.json}
record=docs/candidates.json
recorded='[]'
if [ -f "$record" ]; then
  recorded=$(jq -c . "$record")
fi
file_version=$(file --version | head -n 1)

staging=$(mktemp -d -p . .fetch.XXXXXX)
trap 'rm -rf "$staging"' EXIT

fail=0
err() {
  echo "FAIL: $*" >&2
  fail=1
}

entries=()
paths=()
while IFS=$'\t' read -r gap url; do
  case $url in
    https://raw.githubusercontent.com/* | https://media.githubusercontent.com/media/*)
      repo=${url#https://*/}
      repo=${repo#media/}
      repo=${repo#*/}
      source=${repo%%/*}
      source=${source,,}
      ;;
    https://upload.wikimedia.org/*)
      source=wikimedia-commons
      ;;
    *)
      err "gaps.$gap.candidate: $url is neither a GitHub raw URL nor a Wikimedia Commons upload; name the source's layout in scripts/fetch-candidates.sh"
      continue
      ;;
  esac
  name=${url##*/}
  name=$(printf '%b' "${name//%/\\x}")
  if [[ $name =~ ^(.+\.[A-Za-z0-9]+)\.txt$ ]]; then
    name=${BASH_REMATCH[1]}
  fi
  path="${gap%-*}/$source/$name"
  echo "fetch-candidates: $gap: $url"
  if ! curl -fsSL -o "$staging/$gap" "$url"; then
    err "gaps.$gap.candidate: $url: download failed; check the URL and its ref still resolve"
    continue
  fi
  bytes=$(stat -c %s "$staging/$gap")
  sha=$(sha256sum "$staging/$gap" | cut -d' ' -f1)
  entry=$(jq -c --arg gap "$gap" '.[] | select(.gap == $gap)' <<<"$recorded")
  if [ -z "$entry" ]; then
    entry=$(jq -nc --arg gap "$gap" --arg path "$path" --arg source "$source" --arg sha "$sha" \
      --argjson bytes "$bytes" --arg format "$(file -b --mime-type "$staging/$gap")" --arg by "$file_version" \
      '{gap: $gap, path: $path, source: $source, sha256: $sha, bytes: $bytes, format: $format, format_by: $by}')
  elif ! jq -e --arg path "$path" --arg sha "$sha" --argjson bytes "$bytes" \
    '.path == $path and .sha256 == $sha and .bytes == $bytes' <<<"$entry" >/dev/null; then
    err "gaps.$gap.candidate: $url gave $bytes bytes, sha256 $sha, at $path, but $record records $(jq -r '"\(.bytes) bytes, sha256 \(.sha256), at \(.path)"' <<<"$entry"); a pinned candidate must not change, so name another candidate in its gap"
    continue
  fi
  entries+=("$entry")
  paths+=("$gap"$'\t'"$path")
done < <(jq -r '.gaps[] | select(.candidate.kind == "sourced" and .covered_by == null) | [.id, .candidate.url] | @tsv' "$coverage")

if [ "$fail" -ne 0 ]; then
  echo "FAIL: nothing was placed or recorded" >&2
  exit 1
fi

sourced=$(jq -c '[.gaps[] | select(.candidate.kind == "sourced" and .covered_by == null) | .id]' "$coverage")
printf '%s\n' "${entries[@]}" |
  jq -s --argjson recorded "$recorded" --argjson sourced "$sourced" \
    '. + [$recorded[] | select(.gap as $g | $sourced | index($g) | not)] | sort_by(.path)' >"$staging/candidates.json"

for line in "${paths[@]}"; do
  gap=${line%%$'\t'*}
  path=${line#*$'\t'}
  mkdir -p "$(dirname "$path")"
  mv -f "$staging/$gap" "$path"
done
mv -f "$staging/candidates.json" "$record"
echo "fetch-candidates: placed ${#paths[@]} candidates and wrote $record"
