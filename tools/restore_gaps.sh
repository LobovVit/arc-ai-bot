#!/usr/bin/env bash
set -euo pipefail

SRC_DIR="data/knowledge_base_removed"
DST_DIR="data/docs/knowledge_base"

mkdir -p "$DST_DIR"

for f in "$SRC_DIR"/*.md "$SRC_DIR"/*.txt; do
  [ -e "$f" ] || continue
  base="$(basename "$f")"
  mv "$f" "$DST_DIR/$base"
  echo "restored: $base"
done

echo "Done. Rebuild index: docker compose run --rm ingest"
