#!/usr/bin/env bash
# Packs the sample files MANIFEST.json lists into one uncompressed tar per format area, the top
# directory of each sample path, written to <dir>/<area>.tar with its members at their manifest
# paths in sorted order. Truth and doc files are not packed: git carries them. Fails when a sample
# file is missing or differs from MANIFEST.json, or when an area's tar would reach 2 GiB.
# Usage: scripts/pack.sh <dir>
# See README.md
set -euo pipefail

if [ "$#" -ne 1 ]; then
  echo "FAIL: usage: mise run pack -- <dir>" >&2
  exit 2
fi
out=$(realpath -m -- "$1")

cd "$(dirname "$0")/.."

for tool in bsdtar jq sha256sum; do
  if ! command -v "$tool" >/dev/null; then
    echo "FAIL: $tool is not on PATH; run mise install" >&2
    exit 1
  fi
done

limit=$((2 * 1024 * 1024 * 1024))
fail=0
err() {
  echo "FAIL: $*" >&2
  fail=1
}

while IFS=$'\t' read -r path bytes sha gap; do
  if [ -n "$gap" ]; then
    restore="restore it with mise run fetch-candidates"
  else
    restore="restore it with mise run fetch -- <release-url|dir>, or, for a sourced input, fetch it from its url in docs/sourced.json"
  fi
  if [ ! -f "$path" ]; then
    err "$path: in MANIFEST.json but missing; $restore"
  elif [ "$(stat -c %s "$path")" != "$bytes" ]; then
    err "$path: size differs from MANIFEST.json; $restore"
  elif [ "$(sha256sum "$path" | cut -d' ' -f1)" != "$sha" ]; then
    err "$path: sha256 differs from MANIFEST.json; $restore"
  fi
done < <(jq -r '.files[] | select(.role == "sample") | [.path, (.bytes|tostring), .sha256, (.gap // "")] | @tsv' MANIFEST.json)

if [ "$fail" -ne 0 ]; then
  exit 1
fi

mkdir -p "$out"

while read -r area; do
  tar="$out/$area.tar"
  jq -j --arg area "$area" '
    [.files[] | select(.role == "sample") | .path | select(split("/")[0] == $area)]
    | sort[] | "\(.)\u0000"' MANIFEST.json |
    LC_ALL=C bsdtar -c -f "$tar" --format paxr --null -T - -n \
      --mtime '1970-01-01 00:00:00 UTC' --uid 0 --gid 0 --numeric-owner \
      --no-acls --no-xattrs --no-fflags
  size=$(stat -c %s "$tar")
  if [ "$size" -ge "$limit" ]; then
    rm -f "$tar"
    err "$tar: $size bytes reaches the 2 GiB release asset limit; split $area into smaller format areas"
  else
    echo "pack: $tar ($size bytes)"
  fi
done < <(jq -r '[.files[] | select(.role == "sample") | .path | split("/")[0]] | unique[]' MANIFEST.json)

if [ "$fail" -ne 0 ]; then
  exit 1
fi
