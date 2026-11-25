#!/usr/bin/env bash
# Fetches the sample files MANIFEST.json lists from one uncompressed tar per format area,
# <source>/<area>.tar, where <source> is the base URL of a version's release assets or a local
# directory scripts/pack.sh wrote. The areas are the top directories of the sample paths; an area
# with no samples has no tar. Every file is verified against MANIFEST.json's size and sha256 in a
# staging directory, and moved to its manifest path only when all of them match.
# Usage: scripts/fetch.sh <url|dir>
# See README.md
set -euo pipefail

if [ "$#" -ne 1 ]; then
  echo "FAIL: usage: mise run fetch -- <url|dir>" >&2
  exit 2
fi
source=${1%/}

tools=(bsdtar jq sha256sum)
case $source in
  http://* | https://*)
    remote=1
    tools+=(curl)
    ;;
  *)
    remote=0
    source=$(realpath -m -- "$source")
    ;;
esac
cd "$(dirname "$0")/.."

for tool in "${tools[@]}"; do
  if ! command -v "$tool" >/dev/null; then
    echo "FAIL: $tool is not on PATH; run mise install" >&2
    exit 1
  fi
done

if [ "$remote" -eq 0 ] && [ ! -d "$source" ]; then
  echo "FAIL: $source: not a directory; name a directory mise run pack wrote, or a release URL" >&2
  exit 1
fi

staging=$(mktemp -d -p . .fetch.XXXXXX)
trap 'rm -rf "$staging"' EXIT

fail=0
err() {
  echo "FAIL: $*" >&2
  fail=1
}

declare -A unread=()
refetch="fetch from the release of this checkout's version, or repack the tar with mise run pack"

while read -r area; do
  tar="$source/$area.tar"
  echo "fetch: $tar"
  if [ "$remote" -eq 1 ]; then
    if ! curl -fsSL -o "$staging/$area.tar" "$tar"; then
      err "$tar: download failed; check the URL names this version's release assets"
      unread[$area]=1
      continue
    fi
    tar="$staging/$area.tar"
  elif [ ! -f "$tar" ]; then
    err "$tar: missing; pack it with mise run pack -- $source"
    unread[$area]=1
    continue
  fi
  mkdir -p "$staging/files"
  if ! bsdtar -x -f "$tar" -C "$staging/files" --no-same-owner --no-same-permissions; then
    err "$tar: bsdtar could not extract it; $refetch"
    unread[$area]=1
  fi
  if [ "$remote" -eq 1 ]; then
    rm -f "$tar"
  fi
done < <(jq -r '[.files[] | select(.role == "sample") | .path | split("/")[0]] | unique[]' MANIFEST.json)

while IFS=$'\t' read -r path bytes sha; do
  file="$staging/files/$path"
  area=${path%%/*}
  if [ -n "${unread[$area]:-}" ]; then
    continue
  elif [ ! -f "$file" ] || [ -L "$file" ]; then
    err "$path: in MANIFEST.json but not a file in $area.tar; $refetch"
  elif [ "$(stat -c %s "$file")" != "$bytes" ]; then
    err "$path: size in $area.tar differs from MANIFEST.json; $refetch"
  elif [ "$(sha256sum "$file" | cut -d' ' -f1)" != "$sha" ]; then
    err "$path: sha256 in $area.tar differs from MANIFEST.json; $refetch"
  fi
done < <(jq -r '.files[] | select(.role == "sample") | [.path, (.bytes|tostring), .sha256] | @tsv' MANIFEST.json)

if [ "$fail" -ne 0 ]; then
  echo "FAIL: nothing was placed" >&2
  exit 1
fi

count=0
while read -r path; do
  mkdir -p "$(dirname "$path")"
  mv -f "$staging/files/$path" "$path"
  count=$((count + 1))
done < <(jq -r '.files[] | select(.role == "sample") | .path' MANIFEST.json)
echo "fetch: placed $count sample files"
