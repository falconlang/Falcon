#!/usr/bin/env python3
"""pipeline/deduplicate.py — Deduplicate and cap over-represented categories in problems/.

Usage:
    python3 pipeline/deduplicate.py [--problems problems/] [--output problems_clean/] [--dry-run]

Steps:
  1. Load all PROBLEM*.json files and exact-deduplicate (normalised text comparison).
  2. Classify each problem into one of ~20 keyword categories (first match wins).
  3. Cap each category: keep every ceil(count/cap)-th entry, spreading picks
     evenly from easy (low ID) to hard (high ID) within the category.
  4. Write the surviving problems back to --output, preserving original file
     assignment and ID keys.
"""

import argparse
import json
import math
import os
from pathlib import Path

# ---------------------------------------------------------------------------
# Category taxonomy
# Order matters: first matching category wins.
# Each entry: (category_name, cap, [keyword_substrings ...])
# ---------------------------------------------------------------------------

CATEGORIES: list[tuple[str, int, list[str]]] = [
    ("global_variable", 25,  ["declare a global", "global `", "global variable"]),
    ("local_variable",  20,  ["declare a local", "local `", "local variable"]),
    ("loop_while",      20,  ["while loop", "while ("]),
    ("loop_for",        30,  ["for loop", "for (i:", "for (k:", "iterate over"]),
    ("loop_each",       20,  ["for (", "in names", "each element", "for each"]),
    ("recursion",       30,  ["recursive", "recursion", "base case"]),
    ("list_lambda",     40,  [".map {", ".filter {", ".reduce(", ".sort {", ".min {", ".max {"]),
    ("string_ops",      50,  ["textlen", "trim()", "uppercase()", "lowercase()", "split(", "startswith", "contains(", "segment(", "replace("]),
    ("list_basic",      60,  ["listlen", "add(", "insert(", "remove(", "appendlist", "indexof", "containsitem", "reverseList", "toCsvRow", "slice("]),
    ("dict_basic",      50,  ["dictlen", "containskey", "getAtPath", "setAtPath", "mergeinto", "walktree", "topairs", "pairstodict"]),
    ("function_void",   40,  ["void function", "func "]),
    ("function_result", 40,  ["result function", "= {", "returns "]),
    ("math_basic",      40,  ["sqrt(", "factorial", "fibonacci", "prime", "arithmetic", "pow(", "log(", "sin(", "cos("]),
    ("statistics",      30,  ["mean", "average", "variance", "stddev", "percentile", "distribution", "median", "deviation"]),
    ("algorithm",       80,  ["sort", "search", "dynamic programming", "greedy", "pathfind", "binary search", "mergesort", "quicksort", "heap"]),
    ("graph",           50,  ["graph", "tree", "node", "edge", " traversal", "cycle", "adjacency", "bfs", "dfs"]),
    ("game",            35,  ["player", "health", "spawn", "game state", "attack", "enemy", "level", "score track"]),
    ("data_engineering",35,  ["pipeline", " etl", "record", "normalize", "schema", "transform", "ingest", "aggregat"]),
    ("ml",              35,  ["model", "training", "feature", "classification", "neuron", "gradient", "weight", "predict"]),
    ("cryptography",    25,  ["encrypt", "decrypt", "cipher", "caesar", "vigenere", "rot13", "xor key"]),
    # "other" has no cap — uncategorised problems are kept in full
]

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

def classify(text: str) -> tuple[str, int]:
    """Return (category, cap) for the first matching category, else ('other', -1)."""
    lower = text.lower()
    for name, cap, keywords in CATEGORIES:
        if any(kw.lower() in lower for kw in keywords):
            return name, cap
    return "other", -1


def evenly_spaced(items: list, cap: int) -> list:
    """Return at most `cap` items evenly spread across the list."""
    if len(items) <= cap:
        return items
    step = math.ceil(len(items) / cap)
    return items[::step][:cap]


def normalise(text: str) -> str:
    """Canonical form for exact-dedup comparison."""
    return " ".join(text.lower().split())


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

def main() -> None:
    parser = argparse.ArgumentParser(description="Deduplicate and cap Falcon problem sets.")
    parser.add_argument("--problems",  default="problems/",       help="Source problems directory")
    parser.add_argument("--output",    default="problems_clean/", help="Output directory")
    parser.add_argument("--dry-run",   action="store_true",       help="Report only, do not write files")
    parser.add_argument("--other-cap", type=int, default=1100,    help="Cap for uncategorised 'other' problems (default: 1100)")
    args = parser.parse_args()

    # ------------------------------------------------------------------
    # 1. Load all problems, preserving file source
    # ------------------------------------------------------------------
    all_problems: list[dict] = []   # {id, text, file}
    for p in sorted(Path(args.problems).glob("PROBLEM*.json")):
        with open(p, encoding="utf-8") as f:
            data = json.load(f)
        for pid, text in data.items():
            all_problems.append({"id": str(pid), "text": str(text), "file": p.name})

    total_before = len(all_problems)

    # ------------------------------------------------------------------
    # 2. Exact dedup
    # ------------------------------------------------------------------
    seen_norm: set[str] = set()
    deduped: list[dict] = []
    exact_dropped = 0
    for p in all_problems:
        norm = normalise(p["text"])
        if norm in seen_norm:
            exact_dropped += 1
        else:
            seen_norm.add(norm)
            deduped.append(p)

    # ------------------------------------------------------------------
    # 3. Classify into categories
    # ------------------------------------------------------------------
    # Sort by numeric ID for even-spread sampling to work correctly
    deduped.sort(key=lambda x: int(x["id"]) if x["id"].isdigit() else 0)

    by_category: dict[str, list[dict]] = {}
    for p in deduped:
        cat, _ = classify(p["text"])
        by_category.setdefault(cat, []).append(p)

    # ------------------------------------------------------------------
    # 4. Cap each category
    # ------------------------------------------------------------------
    kept: list[dict] = []
    cap_map = {name: cap for name, cap, _ in CATEGORIES}

    print(f"\n{'Category':<20} {'Before':>7} {'After':>7} {'Dropped':>8}")
    print("-" * 46)

    for cat, problems in sorted(by_category.items()):
        cap = cap_map.get(cat, -1)
        if cap == -1:
            cap = args.other_cap
        if cap == -1:
            selected = problems
        else:
            selected = evenly_spaced(problems, cap)
        dropped = len(problems) - len(selected)
        kept.extend(selected)
        print(f"  {cat:<18} {len(problems):>7} {len(selected):>7} {dropped:>8}")

    total_after = len(kept)
    print("-" * 46)
    print(f"  {'TOTAL':<18} {total_before:>7} {total_after:>7} {total_before - total_after:>8}")
    print(f"\n  Exact duplicates removed: {exact_dropped}")
    print()

    if args.dry_run:
        print("Dry run — no files written.")
        return

    # ------------------------------------------------------------------
    # 5. Write output files (same filenames, same ID keys)
    # ------------------------------------------------------------------
    out_dir = Path(args.output)
    out_dir.mkdir(parents=True, exist_ok=True)

    # Group survivors back by source file
    by_file: dict[str, dict[str, str]] = {}
    for p in kept:
        by_file.setdefault(p["file"], {})[p["id"]] = p["text"]

    for filename, entries in sorted(by_file.items()):
        out_path = out_dir / filename
        with open(out_path, "w", encoding="utf-8") as f:
            json.dump(entries, f, indent=2, ensure_ascii=False)

    print(f"Written {len(by_file)} files to '{args.output}'.")


if __name__ == "__main__":
    main()
