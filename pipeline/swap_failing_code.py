#!/usr/bin/env python3
"""
For each failing entry in MASTER_REASON.jsonl:
  1. Find matching entry in MASTER.jsonl by user prompt
  2. Take the clean code from MASTER.jsonl
  3. Verify it passes the Falcon parser
  4. Substitute the code block in MASTER_REASON.jsonl while preserving <think>...</think>
"""

import json
import re
import subprocess
from pathlib import Path

FALCON_BIN = Path(__file__).parent.parent / "lang" / "Falcon"
MASTER      = Path(__file__).parent.parent / "answers" / "MASTER.jsonl"
MASTER_REASON = Path(__file__).parent.parent / "answer_reasoning" / "MASTER_REASON.jsonl"


def check(code: str) -> tuple[bool, str]:
    r = subprocess.run([str(FALCON_BIN), "format"], input=code,
                       capture_output=True, text=True, timeout=10)
    return r.returncode == 0, r.stderr.strip()


def extract_code(content: str) -> str | None:
    c = re.sub(r"<think>.*?</think>", "", content, flags=re.DOTALL).strip()
    m = re.search(r"```falcon\s*(.*?)\s*```", c, re.DOTALL)
    return m.group(1).strip() if m else None


def replace_code_in_content(content: str, new_code: str) -> str:
    """Replace the ```falcon...``` block after </think>, keeping <think> intact."""
    think_end = content.find("</think>")
    if think_end == -1:
        return re.sub(r"```falcon\s*.*?\s*```", f"```falcon\n{new_code}\n```",
                      content, flags=re.DOTALL)
    prefix = content[:think_end + len("</think>")]
    suffix = content[think_end + len("</think>"):]
    new_suffix = re.sub(r"```falcon\s*.*?\s*```", f"```falcon\n{new_code}\n```",
                        suffix, flags=re.DOTALL)
    return prefix + new_suffix


def get_user_content(record: dict) -> str:
    return next(m["content"] for m in record["messages"] if m["role"] == "user")


def main():
    reason_lines = MASTER_REASON.read_text(encoding="utf-8").splitlines(keepends=False)
    while reason_lines and not reason_lines[-1].strip():
        reason_lines.pop()

    master_lines = MASTER.read_text(encoding="utf-8").splitlines(keepends=False)

    # Build lookup: user_content -> assistant content (from MASTER.jsonl)
    master_lookup: dict[str, str] = {}
    for line in master_lines:
        rec = json.loads(line)
        user = get_user_content(rec)
        asst = next((m["content"] for m in rec["messages"] if m["role"] == "assistant"), None)
        if asst:
            master_lookup[user] = asst

    print("Scanning for failing entries in MASTER_REASON.jsonl...")
    failing = []
    for i, line in enumerate(reason_lines):
        rec = json.loads(line)
        asst = next((m for m in rec["messages"] if m["role"] == "assistant"), None)
        if not asst:
            continue
        code = extract_code(asst["content"])
        if not code:
            continue
        ok, err = check(code)
        if not ok:
            failing.append((i, err, code, rec))

    print(f"Found {len(failing)} failing entries.")

    out_lines = list(reason_lines)
    swapped = 0
    no_match = 0
    master_bad = 0
    already_good = 0

    for idx, err, bad_code, rec in failing:
        user = get_user_content(rec)
        master_asst = master_lookup.get(user)

        if master_asst is None:
            print(f"  [NO MATCH] idx {idx}: {user[:70]!r}")
            no_match += 1
            continue

        clean_code = extract_code(master_asst)
        if not clean_code:
            print(f"  [NO CODE IN MASTER] idx {idx}")
            no_match += 1
            continue

        ok, new_err = check(clean_code)
        if not ok:
            print(f"  [MASTER ALSO BAD] idx {idx}: {new_err[:80]}")
            master_bad += 1
            continue

        # Good — substitute
        asst_msg = next(m for m in rec["messages"] if m["role"] == "assistant")
        new_content = replace_code_in_content(asst_msg["content"], clean_code)
        asst_msg["content"] = new_content
        out_lines[idx] = json.dumps(rec, ensure_ascii=False)
        swapped += 1

    MASTER_REASON.write_text("\n".join(out_lines) + "\n", encoding="utf-8")

    print()
    print("=" * 55)
    print(f"Total failing     : {len(failing)}")
    print(f"Swapped (fixed)   : {swapped}")
    print(f"No match in MASTER: {no_match}")
    print(f"MASTER also bad   : {master_bad}")
    total = len(reason_lines)
    remaining = len(failing) - swapped
    print(f"Est. pass rate    ≈ {(total - remaining) / total * 100:.1f}%")


if __name__ == "__main__":
    main()
