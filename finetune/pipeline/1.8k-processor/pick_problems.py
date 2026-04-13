#!/usr/bin/env python3
"""
Pick 1800 new problems (50% medium / 50% hard, no easy) from the 8k pool,
excluding any already present in MASTER.jsonl.
Classification is text-heuristic based on problem statement keywords.

Output: finetune/dataset/1.8k-codes-steps-output-2/problems.txt
"""
import json
import random
import re
from pathlib import Path

BASE     = Path(__file__).resolve().parents[2] / "dataset"
MASTER   = BASE / "1.8k-codes-steps-output" / "MASTER.jsonl"
SRC_DIR  = BASE / "8k-uncleaned-problem-set"
OUT_DIR  = BASE / "1.8k-codes-steps-output-2"
OUT_FILE = OUT_DIR / "problems.txt"
SEED     = 42

TARGETS = {"easy": 0, "medium": 900, "hard": 900}

HARD_KW = re.compile(
    r"\b(recursion|recursive|nested|graph|algorithm|lambda|cipher|entropy|anagram"
    r"|topological|dynamic programming|memoiz|backtrack|permut|combination"
    r"|matrix|traversal|adjacency|bfs|dfs|heap|trie|segment tree)\b",
    re.I,
)
EASY_KW = re.compile(
    r"\b(declare|global variable|local variable|initialize|assign)\b",
    re.I,
)
MEDIUM_KW = re.compile(  # presence of these prevents easy classification
    r"\b(if|loop|for|while|function|sort|list|dict|map|filter|count|index"
    r"|reverse|class|object|method|return|string|array)\b",
    re.I,
)


def classify(text: str) -> str:
    if HARD_KW.search(text):
        return "hard"
    if EASY_KW.search(text) and not MEDIUM_KW.search(text):
        return "easy"
    return "medium"


# Build exclusion set from MASTER.jsonl
used: set[str] = set()
with open(MASTER) as f:
    for line in f:
        entry = json.loads(line)
        used.add(entry["messages"][0]["content"].strip())
print(f"Loaded {len(used)} used problems from MASTER.jsonl")

# Collect and classify unused problems from all PROBLEM*.json files
pools: dict[str, list[str]] = {"easy": [], "medium": [], "hard": []}
for path in sorted(SRC_DIR.glob("PROBLEM*.json")):
    with open(path) as f:
        data = json.load(f)
    added = 0
    for text in data.values():
        t = text.strip()
        if t not in used:
            pools[classify(t)].append(t)
            added += 1
    print(f"  {path.name}: {added} new problems")

print()
for tier, lst in pools.items():
    print(f"  {tier}: {len(lst)} available (need {TARGETS[tier]})")
print()

# Shuffle deterministically
random.seed(SEED)
for lst in pools.values():
    random.shuffle(lst)

# Select with fallback: if hard pool is short, fill remainder from medium
selected: list[str] = []
medium_used = 0

for tier in ("easy", "medium", "hard"):
    need = TARGETS[tier]
    avail = pools[tier]
    if tier == "medium":
        take = avail[:need]
        medium_used = len(take)
        selected.extend(take)
        if len(take) < need:
            print(f"WARNING: only {len(avail)} medium problems available, taking all")
    elif len(avail) >= need:
        selected.extend(avail[:need])
    else:
        print(f"WARNING: only {len(avail)} {tier} problems available, taking all")
        selected.extend(avail)
        deficit = need - len(avail)
        if tier == "hard":
            # fill from medium overflow
            extra = pools["medium"][medium_used:][:deficit]
            selected.extend(extra)
            print(f"  Filled {len(extra)} from medium overflow to compensate")
    print(f"  selected {tier}: {min(len(avail), need)}")

# Final shuffle so tiers are interleaved
random.seed(SEED + 1)
random.shuffle(selected)

# Write output
OUT_DIR.mkdir(parents=True, exist_ok=True)
with open(OUT_FILE, "w") as f:
    for p in selected:
        f.write(p + "\n")

print(f"\nWrote {len(selected)} problems to {OUT_FILE}")
