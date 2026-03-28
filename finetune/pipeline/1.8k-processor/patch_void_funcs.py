#!/usr/bin/env python3
"""
Patch 12 MASTER.jsonl entries that have void function declarations
using `func name(args) = {` syntax. Removes the `=` sign.

Safe approach:
1. Creates a .bak backup before touching anything
2. Only modifies the 12 confirmed indices
3. Only patches inside ```falcon ... ``` code blocks in assistant messages
4. Verifies JSON round-trip integrity per line
5. Reports every change made
"""

import json
import re
import shutil
import sys
from pathlib import Path

MASTER = Path("/home/kumaraswamy/Documents/falcon/finetune/dataset/1.8k-codes-steps-output/MASTER.jsonl")
BACKUP = MASTER.with_suffix(".jsonl.bak_void")

TARGET_INDICES = {789, 965, 1669, 1719, 1720, 1747, 1748, 1767, 1780, 1788, 1817, 1827}

VOID_FUNC_RE = re.compile(r'(func\s+\w+\s*\([^)]*\))\s*=\s*\{')

def patch_code(code: str) -> tuple[str, int]:
    """Apply the void-func patch. Returns (patched_code, num_replacements)."""
    patched, count = VOID_FUNC_RE.subn(r'\1 {', code)
    return patched, count

def patch_falcon_blocks(text: str) -> tuple[str, int]:
    """Find all ```falcon ... ``` blocks and patch each one."""
    total = 0
    def replacer(m):
        nonlocal total
        code = m.group(1)
        patched, n = patch_code(code)
        total += n
        return f"```falcon\n{patched}\n```"
    result = re.sub(r'```falcon\n(.*?)\n```', replacer, text, flags=re.DOTALL)
    return result, total

def main():
    if not MASTER.exists():
        print(f"ERROR: {MASTER} not found", file=sys.stderr)
        sys.exit(1)

    # Step 1: Backup
    shutil.copy2(MASTER, BACKUP)
    print(f"Backup created: {BACKUP}")

    lines = MASTER.read_text(encoding="utf-8").splitlines()
    print(f"Total lines: {len(lines)}")

    patched_lines = list(lines)
    patch_report = {}

    for idx in sorted(TARGET_INDICES):
        if idx >= len(lines):
            print(f"  SKIP idx={idx}: out of range")
            continue

        line = lines[idx]
        try:
            entry = json.loads(line)
        except json.JSONDecodeError as e:
            print(f"  ERROR idx={idx}: JSON parse failed: {e}")
            continue

        messages = entry.get("messages", [])
        changed = False
        for msg in messages:
            if msg.get("role") != "assistant":
                continue
            content = msg.get("content", "")
            new_content, n = patch_falcon_blocks(content)
            if n > 0:
                msg["content"] = new_content
                changed = True
                patch_report[idx] = n

        if not changed:
            print(f"  WARN idx={idx}: no ```falcon blocks with void-func pattern found — skipping")
            continue

        # Verify JSON round-trip
        new_line = json.dumps(entry, ensure_ascii=False)
        try:
            json.loads(new_line)
        except json.JSONDecodeError as e:
            print(f"  ERROR idx={idx}: JSON round-trip failed: {e} — NOT patching")
            continue

        patched_lines[idx] = new_line
        print(f"  PATCHED idx={idx}: {patch_report[idx]} replacement(s)")

    # Step 2: Write output
    output = "\n".join(patched_lines)
    MASTER.write_text(output, encoding="utf-8")

    print(f"\nDone. Patched {len(patch_report)}/{len(TARGET_INDICES)} entries.")
    print(f"Original backed up at: {BACKUP}")

if __name__ == "__main__":
    main()
