#!/usr/bin/env python3
"""
Task 2 helper:
- Reads raw docs from --in_dir (txt/md), replaces terms using terms_map.json
- Writes replaced docs into --out_dir
- Writes replacement counts report

Usage:
  python tools/replace_terms.py --in_dir raw_docs --out_dir knowledge_base --map terms_map.json
"""
from __future__ import annotations
import argparse, json, pathlib, re

def load_map(path: str) -> dict[str,str]:
    with open(path, "r", encoding="utf-8") as f:
        return json.load(f)

def replace_text(text: str, mapping: dict[str,str]):
    counts = {}
    out = text
    for src in sorted(mapping.keys(), key=len, reverse=True):
        dst = mapping[src]
        out, n = re.subn(re.escape(src), dst, out)
        if n:
            counts[src] = counts.get(src, 0) + n
    return out, counts

def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--in_dir", required=True)
    ap.add_argument("--out_dir", required=True)
    ap.add_argument("--map", default="terms_map.json")
    ap.add_argument("--report", default="replace_report.json")
    args = ap.parse_args()

    mapping = load_map(args.map)
    in_dir = pathlib.Path(args.in_dir)
    out_dir = pathlib.Path(args.out_dir)
    out_dir.mkdir(parents=True, exist_ok=True)

    report = {}
    for p in in_dir.rglob("*"):
        if not p.is_file():
            continue
        if p.suffix.lower() not in (".md", ".mdx", ".txt"):
            continue
        text = p.read_text(encoding="utf-8", errors="ignore")
        replaced, counts = replace_text(text, mapping)
        rel = p.relative_to(in_dir)
        out_path = out_dir / rel.with_suffix(".md")
        out_path.parent.mkdir(parents=True, exist_ok=True)
        out_path.write_text(replaced, encoding="utf-8")
        report[str(rel)] = counts

    with open(args.report, "w", encoding="utf-8") as f:
        json.dump(report, f, ensure_ascii=False, indent=2)

    print(f"Done. Output: {out_dir} | report: {args.report}")

if __name__ == "__main__":
    main()
