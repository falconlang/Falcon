#!/usr/bin/env python3
"""
Check how many codes in MASTER_REASON.jsonl pass the Falcon syntax parser.
Uses `lang/Falcon format` (reads from stdin, exits 1 on parse error).
"""

import json
import subprocess
import sys
import re
from pathlib import Path

FALCON_BIN = Path(__file__).parent.parent / "lang" / "Falcon"
MASTER_REASON = Path(__file__).parent.parent / "answer_reasoning" / "MASTER_REASON.jsonl"


def extract_code(content: str) -> str | None:
    """Extract the falcon code block from assistant content."""
    # Strip <think>...</think> first
    content = re.sub(r"<think>.*?</think>", "", content, flags=re.DOTALL).strip()
    # Extract ```falcon ... ```
    m = re.search(r"```falcon\s*(.*?)\s*```", content, re.DOTALL)
    if m:
        return m.group(1).strip()
    return None


def check_syntax(code: str) -> bool:
    """Return True if code passes Falcon format (parse) check."""
    result = subprocess.run(
        [str(FALCON_BIN), "format"],
        input=code,
        capture_output=True,
        text=True,
        timeout=10,
    )
    return result.returncode == 0


def main():
    passed = 0
    failed = 0
    no_code = 0
    failed_indices = []

    with open(MASTER_REASON, encoding="utf-8") as f:
        lines = f.readlines()

    total = len(lines)
    for i, line in enumerate(lines):
        if (i + 1) % 100 == 0 or i == 0:
            print(f"  [{i+1}/{total}] passed={passed} failed={failed} no_code={no_code}", flush=True)

        try:
            record = json.loads(line)
        except json.JSONDecodeError as e:
            print(f"  [idx {i}] JSON error: {e}", file=sys.stderr)
            failed += 1
            failed_indices.append(i)
            continue

        messages = record.get("messages", [])
        assistant_msg = next((m for m in messages if m["role"] == "assistant"), None)
        if not assistant_msg:
            no_code += 1
            continue

        code = extract_code(assistant_msg["content"])
        if not code:
            print(f"  [idx {i}] No falcon code block found", file=sys.stderr)
            no_code += 1
            continue

        try:
            ok = check_syntax(code)
        except subprocess.TimeoutExpired:
            print(f"  [idx {i}] Timeout", file=sys.stderr)
            failed += 1
            failed_indices.append(i)
            continue

        if ok:
            passed += 1
        else:
            failed += 1
            failed_indices.append(i)

    print()
    print("=" * 50)
    print(f"Total entries : {total}")
    print(f"Passed        : {passed}  ({100*passed/total:.1f}%)")
    print(f"Failed        : {failed}  ({100*failed/total:.1f}%)")
    print(f"No code block : {no_code}")
    print()
    if failed_indices:
        print(f"Failed indices (first 50): {failed_indices[:50]}")


if __name__ == "__main__":
    main()
