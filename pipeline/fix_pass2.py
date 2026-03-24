#!/usr/bin/env python3
"""
Second-pass fixer: handles remaining errors after fix_all_errors.py
New fixes:
  P1. `and`  → `&&`  (keyword to operator)
  P2. `or`   → `||`
  P3. `not`  → `!`
  P4. .toNumber() / .toInt() / .toFloat() → dec() wrapper
  P5. .ln()  → log()  standalone
  P6. .pow(n) → ^ n  (rewrite as expression)
  P7. .repeat(n) — replace with helper loop expression
  P8. .toString() → "" _ x  or just leave (strings are auto-converted)
  P9. [[a,b],[c,d]] nested list literal — not supported; rewrite with .add()
  P10. Re-apply all fixes from pass-1 (compound errors may now be solvable)
"""

import json
import subprocess
import re
import shutil
from pathlib import Path
import sys
sys.path.insert(0, str(Path(__file__).parent))
from fix_all_errors import (
    extract_code, replace_code_in_content, apply_fixes as pass1_fixes,
    check, MASTER_REASON
)

# ---------- new fixers ----------

def fix_and_or(code: str) -> str:
    """Replace `and` / `or` / `not` keywords with Falcon operators, outside strings."""
    result = []
    in_str = False
    str_char = None
    i = 0
    while i < len(code):
        c = code[i]
        if in_str:
            result.append(c)
            if c == str_char and (i == 0 or code[i-1] != '\\'):
                in_str = False
            i += 1
        else:
            if c in ('"', "'"):
                in_str = True
                str_char = c
                result.append(c)
                i += 1
            else:
                # Check for keyword at word boundary
                matched = False
                for kw, op in [("and", "&&"), ("or", "||")]:
                    if code[i:i+len(kw)] == kw:
                        before = code[i-1] if i > 0 else ' '
                        after = code[i+len(kw)] if i+len(kw) < len(code) else ' '
                        if not (before.isalnum() or before == '_') and \
                           not (after.isalnum() or after == '_'):
                            result.append(op)
                            i += len(kw)
                            matched = True
                            break
                if not matched:
                    result.append(c)
                    i += 1
    return ''.join(result)


def fix_not_kw(code: str) -> str:
    """Replace `not EXPR` → `!EXPR` at word boundaries."""
    return re.sub(r'\bnot\s+', '!', code)


def fix_to_number(code: str) -> str:
    """
    VAR.toNumber() → dec(VAR)
    VAR.toInt()    → floor(dec(VAR))
    VAR.toFloat()  → dec(VAR)
    VAR.parseInt() → floor(dec(VAR))
    """
    code = re.sub(r'(\w+)\.toNumber\(\)', r'dec(\1)', code)
    code = re.sub(r'(\w+)\.toFloat\(\)', r'dec(\1)', code)
    code = re.sub(r'(\w+)\.toInt\(\)', r'floor(dec(\1))', code)
    code = re.sub(r'(\w+)\.parseInt\(\)', r'floor(dec(\1))', code)
    code = re.sub(r'(\w+)\.parseFloat\(\)', r'dec(\1)', code)
    return code


def fix_ln(code: str) -> str:
    """VAR.ln() → log(VAR)"""
    return re.sub(r'(\w+)\.ln\(\)', r'log(\1)', code)


def fix_pow(code: str) -> str:
    """
    VAR.pow(N) → VAR ^ (N)
    (EXPR).pow(N) → (EXPR) ^ (N)
    """
    code = re.sub(r'(\w+)\.pow\(([^)]+)\)', r'(\1 ^ (\2))', code)
    return code


def fix_to_string(code: str) -> str:
    """VAR.toString() → ("" _ VAR)"""
    code = re.sub(r'(\w+)\.toString\(\)', r'("" _ \1)', code)
    return code


def fix_repeat(code: str) -> str:
    """
    [VAL].repeat(N) → list built by a for loop wrapped in an IIFE-style func.
    Use Falcon's result-function: (func _r() = { local l = []\n  for (_i: 1 .. N) { l.add(VAL) }\n  l })()
    BUT Falcon doesn't support IIFE. Instead emit a helper global and call it.
    Simpler: replace inline with a comprehension-like approach using map.
    Actually simplest valid Falcon: need a named helper.
    Strategy: rewrite as just the initializer at declaration time using a for loop.
    """
    # Pattern: local NAME = [VAL].repeat(N)
    def repl(m):
        name = m.group(1)
        val = m.group(2).strip()
        n_expr = m.group(3).strip()
        return (f"local {name} = []\n"
                f"for (_ri: 1 .. {n_expr}) {{ {name}.add({val}) }}")
    code = re.sub(
        r'local\s+(\w+)\s*=\s*\[([^\]]+)\]\.repeat\(([^)]+)\)',
        repl,
        code
    )
    # Fallback: any [VAL].repeat(N) not in assignment → leave (too complex)
    return code


def fix_nested_list_literal(code: str) -> str:
    """
    Falcon list literals are 1D: [a, b, c]. Nested [[a,b],[c,d]] may fail.
    Skip — too complex to rewrite without type info.
    """
    return code


ALL_P2_FIXERS = [
    fix_and_or,
    fix_not_kw,
    fix_to_number,
    fix_ln,
    fix_pow,
    fix_to_string,
    fix_repeat,
]


def apply_all(code: str) -> str:
    # Re-apply pass-1 first (for compound error cases)
    code = pass1_fixes(code)
    # Then apply pass-2
    for fixer in ALL_P2_FIXERS:
        code = fixer(code)
    # One more pass-1 in case pass-2 exposed new fixable things
    code = pass1_fixes(code)
    return code


def main():
    lines = MASTER_REASON.read_text(encoding="utf-8").splitlines(keepends=False)
    while lines and not lines[-1].strip():
        lines.pop()
    total = len(lines)

    fixed_count = 0
    already_ok = 0
    still_failing = 0

    out_lines = []
    for i, line in enumerate(lines):
        if (i + 1) % 200 == 0:
            print(f"  [{i+1}/{total}]  fixed={fixed_count}  still_failing={still_failing}", flush=True)

        record = json.loads(line)
        messages = record["messages"]
        asst = next((m for m in messages if m["role"] == "assistant"), None)
        if not asst:
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

        fixed_code = apply_all(code)
        ok2, err2 = check(fixed_code)
        if ok2 and fixed_code != code:
            new_content = replace_code_in_content(asst["content"], fixed_code)
            asst["content"] = new_content
            out_lines.append(json.dumps(record, ensure_ascii=False))
            fixed_count += 1
        else:
            out_lines.append(line)
            still_failing += 1

    backup = MASTER_REASON.with_suffix(".jsonl.bak_p2")
    if not backup.exists():
        shutil.copy(MASTER_REASON, backup)
        print(f"Backup saved to {backup.name}")

    MASTER_REASON.write_text("\n".join(out_lines) + "\n", encoding="utf-8")

    print()
    print("=" * 55)
    print(f"Total entries      : {total}")
    print(f"Already passing    : {already_ok}")
    print(f"Fixed by pass-2    : {fixed_count}")
    print(f"Still failing      : {still_failing}")
    print(f"New pass rate      : {(already_ok + fixed_count) / total * 100:.1f}%")


if __name__ == "__main__":
    main()
