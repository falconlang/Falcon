#!/usr/bin/env python3
"""Collect parse errors from MASTER_REASON.jsonl and tally by error type."""

import json
import subprocess
import re
from pathlib import Path
from collections import Counter

FALCON_BIN = Path(__file__).parent.parent / "lang" / "Falcon"
MASTER_REASON = Path(__file__).parent.parent / "answer_reasoning" / "MASTER_REASON.jsonl"


def extract_code(content: str) -> str | None:
    content = re.sub(r"<think>.*?</think>", "", content, flags=re.DOTALL).strip()
    m = re.search(r"```falcon\s*(.*?)\s*```", content, re.DOTALL)
    return m.group(1).strip() if m else None


def check(code: str):
    """Returns (ok, stderr_text)."""
    result = subprocess.run(
        [str(FALCON_BIN), "format"],
        input=code, capture_output=True, text=True, timeout=10,
    )
    return result.returncode == 0, result.stderr.strip()


def normalize_error(msg: str) -> str:
    """Strip line/col numbers to get a canonical error key."""
    # e.g. "Error: unexpected token at line 3 col 5: 'foo'" -> "unexpected token: 'foo'"
    msg = re.sub(r"\bat line \d+\b", "", msg)
    msg = re.sub(r"\bcol(?:umn)? \d+\b", "", msg)
    msg = re.sub(r"\d+:\d+", "", msg)
    msg = re.sub(r"\s{2,}", " ", msg).strip()
    return msg


errors: Counter = Counter()
failed_examples: dict[str, list] = {}   # error_key -> [(idx, code)]

with open(MASTER_REASON, encoding="utf-8") as f:
    lines = f.readlines()

total = len(lines)
failed = 0
for i, line in enumerate(lines):
    if (i + 1) % 200 == 0:
        print(f"  [{i+1}/{total}] failed so far: {failed}", flush=True)
    record = json.loads(line)
    msg = next((m for m in record["messages"] if m["role"] == "assistant"), None)
    if not msg:
        continue
    code = extract_code(msg["content"])
    if not code:
        continue
    ok, stderr = check(code)
    if not ok:
        failed += 1
        key = normalize_error(stderr) or "<empty stderr>"
        errors[key] += 1
        if key not in failed_examples:
            failed_examples[key] = []
        if len(failed_examples[key]) < 3:
            failed_examples[key].append((i, code[:300]))

print(f"\nTotal failed: {failed}\n")
print("=" * 70)
print("TOP 20 ERRORS (by frequency)")
print("=" * 70)
for rank, (key, count) in enumerate(errors.most_common(20), 1):
    print(f"\n#{rank}  [{count}x]  {key}")
    for idx, snippet in failed_examples[key][:2]:
        print(f"  --- idx {idx} ---")
        print("  " + snippet.replace("\n", "\n  ")[:200])
