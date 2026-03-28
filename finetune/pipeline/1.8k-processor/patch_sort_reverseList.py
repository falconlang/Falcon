#!/usr/bin/env python3
"""
Patch MASTER.jsonl entries where .sort() / .reverseList() are used as statements
instead of being assigned back (Falcon returns a new list; it does not mutate).

Fix: `var.sort()` → `var = var.sort()`, same for .reverseList() and sort-with-lambda.

Safe approach:
1. Backup before touching anything
2. Only modify entries that currently FAIL with a sort/reverseList never-used error
3. Only apply patterns that are verified to turn a failing entry into a passing one
4. Skip any entry where the fix doesn't resolve the error
5. JSON round-trip check per entry
"""

import json
import re
import shutil
import subprocess
import sys
import tempfile
from pathlib import Path

MASTER = Path("/home/kumaraswamy/Documents/falcon/finetune/dataset/1.8k-codes-steps-output/MASTER.jsonl")
BACKUP = MASTER.with_suffix(".jsonl.bak_sort")
FALCON = "/home/kumaraswamy/Documents/falcon/lang/Falcon"

# Each pattern targets a specific form of standalone sort/reverseList statement.
# \s* allows zero indentation (top-level code).
PATTERNS = [
    # Inline: `[indent]var.sort()` as a whole line
    (re.compile(r'^(\s*)(\w+)\.(sort\(\))\s*$', re.MULTILINE), r'\1\2 = \2.\3'),
    # Inline: `[indent]var.reverseList()` as a whole line
    (re.compile(r'^(\s*)(\w+)\.(reverseList\(\))\s*$', re.MULTILINE), r'\1\2 = \2.\3'),
    # Inline sort with lambda (single-level braces): `[indent]var.sort { ... }`
    (re.compile(r'^(\s*)(\w+)\.(sort\s*\{[^}\n]*\})\s*$', re.MULTILINE), r'\1\2 = \2.\3'),

    # Multiline: var on its own line, .sort() indented on next line
    (re.compile(r'^(\s*)(\w+)\n(\s+)(\.sort\(\))\s*$', re.MULTILINE), r'\1\2 = \2\n\3\4'),
    # Multiline: var on its own line, .reverseList() indented on next line
    (re.compile(r'^(\s*)(\w+)\n(\s+)(\.reverseList\(\))\s*$', re.MULTILINE), r'\1\2 = \2\n\3\4'),
    # Multiline sort with lambda (single-level braces)
    (re.compile(r'^(\s*)(\w+)\n(\s+)(\.sort\s*\{[^}\n]*\})\s*$', re.MULTILINE), r'\1\2 = \2\n\3\4'),
]


def apply_fix(code: str) -> str:
    for pat, repl in PATTERNS:
        code = pat.sub(repl, code)
    return code


def extract_code(content: str) -> str:
    m = re.search(r'```falcon\n(.*?)```', content, re.DOTALL)
    return m.group(1).strip() if m else content.strip()


def replace_code_in_content(content: str, new_code: str) -> str:
    return re.sub(r'(```falcon\n).*?(```)', r'\g<1>' + new_code.replace('\\', '\\\\') + r'\n\2',
                  content, flags=re.DOTALL)


def run_falcon(code: str) -> tuple[int, str]:
    with tempfile.NamedTemporaryFile(mode='w', suffix='.mist', delete=False) as f:
        f.write(code)
        path = f.name
    r = subprocess.run([FALCON, "run", path], capture_output=True, text=True, timeout=10)
    Path(path).unlink(missing_ok=True)
    return r.returncode, r.stderr.strip()


def is_sort_reverselist_failure(err: str) -> bool:
    m = re.search(r"result of `(.+?)` is never used", err, re.DOTALL)
    if m:
        expr = m.group(1)
        return ".sort" in expr or ".reverseList" in expr
    return False


def main():
    if not MASTER.exists():
        print(f"ERROR: {MASTER} not found", file=sys.stderr)
        sys.exit(1)

    shutil.copy2(MASTER, BACKUP)
    print(f"Backup created: {BACKUP}")

    raw_lines = MASTER.read_text(encoding="utf-8").splitlines()
    print(f"Total lines: {len(raw_lines)}")

    patched_lines = list(raw_lines)
    patched_count = 0
    skipped_no_match = 0
    skipped_still_fails = 0
    skipped_no_error = 0

    for i, raw in enumerate(raw_lines):
        try:
            entry = json.loads(raw)
        except json.JSONDecodeError:
            continue

        messages = entry.get("messages", [])
        if len(messages) < 2:
            continue

        assistant_content = messages[1].get("content", "")
        code = extract_code(assistant_content)

        # Check if this entry currently fails with a sort/reverseList error
        rc, err = run_falcon(code)
        if rc == 0:
            skipped_no_error += 1
            continue
        if not is_sort_reverselist_failure(err):
            continue

        # Apply fix
        fixed_code = apply_fix(code)
        if fixed_code == code:
            print(f"  SKIP idx={i}: patterns didn't match — {err[:60]}")
            skipped_no_match += 1
            continue

        # Verify the fix actually resolves the error
        rc2, err2 = run_falcon(fixed_code)
        if rc2 != 0:
            print(f"  SKIP idx={i}: fix doesn't resolve error — {err2[:60]}")
            skipped_still_fails += 1
            continue

        # Update the entry
        messages[1]["content"] = replace_code_in_content(assistant_content, fixed_code)
        new_line = json.dumps(entry, ensure_ascii=False)
        try:
            json.loads(new_line)
        except json.JSONDecodeError as e:
            print(f"  SKIP idx={i}: JSON round-trip failed: {e}")
            continue

        patched_lines[i] = new_line
        patched_count += 1

        orig_l = code.splitlines()
        fix_l = fixed_code.splitlines()
        diffs = [(o, n) for o, n in zip(orig_l, fix_l) if o != n]
        print(f"  PATCHED idx={i}: {len(diffs)} line(s) changed")
        for o, n in diffs[:2]:
            print(f"    - {o.strip()}")
            print(f"    + {n.strip()}")

    MASTER.write_text("\n".join(patched_lines), encoding="utf-8")

    print(f"\nDone.")
    print(f"  Patched:              {patched_count}")
    print(f"  Skipped (no match):   {skipped_no_match}")
    print(f"  Skipped (still fail): {skipped_still_fails}")
    print(f"  Backup: {BACKUP}")


if __name__ == "__main__":
    main()
