#!/usr/bin/env python3
"""
For all still-failing entries in MASTER_REASON.jsonl:
  1. Find the matching entry in MASTER.jsonl by user content
  2. Replace the code block (after </think>) with the ground-truth code
  3. Also fix known fixer-introduced bugs:
     - For loop colon→in on negative ranges: for (x in -N .. M) → for (x: -N .. M)
     - Variable names starting with _ (replace with valid names)
"""

import json
import subprocess
import re
import shutil
from pathlib import Path

FALCON_BIN  = Path(__file__).parent.parent / "lang" / "Falcon"
MASTER_REASON = Path(__file__).parent.parent / "answer_reasoning" / "MASTER_REASON.jsonl"
MASTER_JSONL  = Path(__file__).parent.parent / "answers" / "MASTER.jsonl"


def check(code: str) -> tuple[bool, str]:
    r = subprocess.run([str(FALCON_BIN), "format"], input=code,
                       capture_output=True, text=True, timeout=10)
    return r.returncode == 0, r.stderr.strip()


def extract_code(content: str) -> str | None:
    c = re.sub(r"<think>.*?</think>", "", content, flags=re.DOTALL).strip()
    m = re.search(r"```falcon\s*(.*?)\s*```", c, re.DOTALL)
    return m.group(1).strip() if m else None


def replace_code_in_content(content: str, new_code: str) -> str:
    think_end = content.find("</think>")
    if think_end == -1:
        return re.sub(r"```falcon\s*.*?\s*```", f"```falcon\n{new_code}\n```", content, flags=re.DOTALL)
    prefix = content[:think_end + len("</think>")]
    suffix = content[think_end + len("</think>"):]
    new_suffix = re.sub(r"```falcon\s*.*?\s*```", f"```falcon\n{new_code}\n```", suffix, flags=re.DOTALL)
    return prefix + new_suffix


def fix_for_in_on_range(code: str) -> str:
    """
    Fix: for (VAR in -N .. M) should be for (VAR: -N .. M)
    The original fixer wrongly converted negative-start ranges.
    Also fix: for (VAR in N .. M) where N is a number → for (VAR: N .. M)
    """
    def repl(m):
        var = m.group(1)
        rest = m.group(2)
        return f"for ({var}: {rest}"
    # for (VAR in NUMBER .. ...) or (VAR in -NUMBER ...)
    code = re.sub(r'for\s*\(\s*(\w+)\s+in\s+([-\d])', repl, code)
    return code


def fix_underscore_vars(code: str) -> str:
    """Replace _-prefixed variable names introduced by fixers with valid names."""
    replacements = {
        '_mat_out': 'matOut',
        '_rgb_color': 'rgbColor',
        '_rgb_c': 'rgbC',
        '_rgb_': 'rgb',
        '_ri': 'ri',
        '_i': 'loopI',
        '_r': 'loopR',
        '_l': 'loopL',
        '_mk': 'mkList',
    }
    for old, new in replacements.items():
        code = code.replace(old, new)
    # Generic: any remaining _identifier
    code = re.sub(r'\b_(\w)', lambda m: m.group(1).upper() if m.group(1).isalpha() else m.group(1), code)
    return code


# -------- build lookup from MASTER.jsonl --------

print("Loading MASTER.jsonl ground truth...")
master_by_user: dict[str, str] = {}
with open(MASTER_JSONL, encoding="utf-8") as f:
    for line in f:
        rec = json.loads(line)
        msgs = rec.get("messages", [])
        user_msg = next((m["content"] for m in msgs if m["role"] == "user"), None)
        asst_msg = next((m["content"] for m in msgs if m["role"] == "assistant"), None)
        if user_msg and asst_msg:
            # Extract code from assistant content (plain ```falcon block)
            cm = re.search(r"```falcon\s*(.*?)\s*```", asst_msg, re.DOTALL)
            if cm:
                master_by_user[user_msg.strip()] = cm.group(1).strip()

print(f"Loaded {len(master_by_user)} ground truth entries.")


# -------- process MASTER_REASON.jsonl --------

lines = MASTER_REASON.read_text(encoding="utf-8").splitlines(keepends=False)
while lines and not lines[-1].strip():
    lines.pop()
total = len(lines)

backup = MASTER_REASON.with_suffix(".jsonl.bak_gt")
if not backup.exists():
    shutil.copy(MASTER_REASON, backup)
    print(f"Backup saved to {backup.name}")

fixed_by_gt = 0
fixed_by_local = 0
still_failing = 0
already_ok = 0
no_gt_match = 0

out_lines = []
for i, line in enumerate(lines):
    if (i + 1) % 200 == 0:
        print(f"  [{i+1}/{total}]  gt={fixed_by_gt}  local={fixed_by_local}  still={still_failing}", flush=True)

    record = json.loads(line)
    messages = record["messages"]
    asst = next((m for m in messages if m["role"] == "assistant"), None)
    user = next((m for m in messages if m["role"] == "user"), None)
    if not asst or not user:
        out_lines.append(line)
        already_ok += 1
        continue

    code = extract_code(asst["content"])
    if not code:
        out_lines.append(line)
        already_ok += 1
        continue

    ok, _ = check(code)
    if ok:
        out_lines.append(line)
        already_ok += 1
        continue

    # --- Try local fixes first (fixer-induced bugs) ---
    fixed = fix_for_in_on_range(code)
    fixed = fix_underscore_vars(fixed)
    ok2, _ = check(fixed)
    if ok2 and fixed != code:
        new_content = replace_code_in_content(asst["content"], fixed)
        asst["content"] = new_content
        out_lines.append(json.dumps(record, ensure_ascii=False))
        fixed_by_local += 1
        continue

    # --- Try ground truth code ---
    user_key = user["content"].strip()
    gt_code = master_by_user.get(user_key)
    if not gt_code:
        # Try partial match (first 200 chars)
        prefix = user_key[:200]
        gt_code = next((v for k, v in master_by_user.items() if k[:200] == prefix), None)

    if gt_code:
        ok_gt, _ = check(gt_code)
        if ok_gt:
            new_content = replace_code_in_content(asst["content"], gt_code)
            asst["content"] = new_content
            out_lines.append(json.dumps(record, ensure_ascii=False))
            fixed_by_gt += 1
            continue
        else:
            # GT code itself doesn't parse?  Apply local fixers + underscore fix to it too
            fixed_gt = fix_for_in_on_range(gt_code)
            fixed_gt = fix_underscore_vars(fixed_gt)
            ok_gt2, _ = check(fixed_gt)
            if ok_gt2:
                new_content = replace_code_in_content(asst["content"], fixed_gt)
                asst["content"] = new_content
                out_lines.append(json.dumps(record, ensure_ascii=False))
                fixed_by_gt += 1
                continue
    else:
        no_gt_match += 1

    out_lines.append(line)
    still_failing += 1

MASTER_REASON.write_text("\n".join(out_lines) + "\n", encoding="utf-8")

print()
print("=" * 55)
print(f"Total entries      : {total}")
print(f"Already passing    : {already_ok}")
print(f"Fixed by local     : {fixed_by_local}")
print(f"Fixed by GT        : {fixed_by_gt}")
print(f"Still failing      : {still_failing}")
print(f"  (no GT match)    : {no_gt_match}")
print(f"New pass rate      : {(already_ok + fixed_by_local + fixed_by_gt) / total * 100:.1f}%")
