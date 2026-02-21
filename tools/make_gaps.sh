#!/usr/bin/env bash
set -euo pipefail

SRC_DIR="data/docs/knowledge_base"
DST_DIR="data/knowledge_base_removed"

mkdir -p "$DST_DIR"

# выбери 2–3 ключевых сущности
FILES=("synth_flux.md" "fluxblade.md" "slipdrive.md")

for f in "${FILES[@]}"; do
  if [ -f "$SRC_DIR/$f" ]; then
    mv "$SRC_DIR/$f" "$DST_DIR/$f"
    echo "moved: $f"
  else
    echo "skip (not found): $f"
  fi
done

echo "Done. Rebuild index: docker compose run --rm ingest"
