#!/usr/bin/env python3
"""
Progressive multi-pass fixer.
Each round: apply ALL transforms to ALL still-failing entries,
save the result unconditionally (even if still failing), repeat until stable.

Also adds new transforms discovered from debug_remaining.py output:
  - func(...) = {...} as argument → pre-declare as named func
  - [[a,b],[c,d]] 2D list → rewrite with .add()
  - Unicode smart quotes → ASCII quotes
  - .segment(i, i) (end-index form) → .segment(i, 1)  [2nd arg = length]
  - if (COND) then EXPR → if (COND) EXPR  [stray 'then' after paren]
  - abs(x) instead of x.abs()
  - various chained methods missing from pass-1
"""

import json
import subprocess
import re
import shutil
from pathlib import Path
import sys
sys.path.insert(0, str(Path(__file__).parent))
from fix_all_errors import (extract_code, replace_code_in_content,
                             check, MASTER_REASON, ALL_FIXERS)
from fix_pass2 import ALL_P2_FIXERS

# -------- additional transforms --------

def fix_smart_quotes(code: str) -> str:
    """Replace Unicode curly/smart quotes with straight ASCII quotes."""
    code = code.replace('\u201c', '"').replace('\u201d', '"')   # "" → ""
    code = code.replace('\u2018', "'").replace('\u2019', "'")   # '' → ''
    code = code.replace('\u2013', '-').replace('\u2014', '-')   # – — → -
    return code


def fix_segment_end_form(code: str) -> str:
    """
    Some generated code uses segment(from, to) where 2nd arg is an end index.
    Falcon uses segment(from, length). Hard to auto-detect; skip complex cases.
    Simple fix: segment(i, i) → segment(i, 1)
    """
    # segment(VAR, VAR) where both vars are the same → length=1
    code = re.sub(r'\.segment\((\w+),\s*\1\)', r'.segment(\1, 1)', code)
    return code


def fix_inline_func_args(code: str) -> str:
    """
    Falcon does NOT support anonymous function literals as call arguments:
      foo(func(x) = { x + 1 })
    We can't easily auto-fix these without rewriting the entire call site.
    Best effort: detect and skip (leave as-is, will still fail).
    """
    return code  # No reliable auto-fix available


def fix_nested_list_literal(code: str) -> str:
    """
    [[a, b], [c, d]] — Falcon can't have nested list literals inline.
    Replace with .add() chains where the pattern is simple enough.
    Pattern: return [[...], [...]]  at end of function
    """
    # Only handle the simple 2-row matrix return case
    def repl_2row(m):
        indent = m.group(1)
        row1 = m.group(2).strip()
        row2 = m.group(3).strip()
        var = "_mat_out"
        return (f"{indent}local {var} = []\n"
                f"{indent}{var}.add([{row1}])\n"
                f"{indent}{var}.add([{row2}])\n"
                f"{indent}{var}")
    code = re.sub(
        r'^(\s*)\[\[([^\[\]]+)\],\s*\[([^\[\]]+)\]\]',
        repl_2row,
        code,
        flags=re.MULTILINE
    )
    return code


def fix_stray_then(code: str) -> str:
    """Remove 'then' that appears after 'if (COND)' with parens already present."""
    code = re.sub(r'(if\s*\([^)]+\))\s+then\b', r'\1', code)
    return code


def fix_method_length_aggressive(code: str) -> str:
    """
    More aggressive .length() fix: if variable ends with 's' (plural) → listLen,
    single-char names like 's', 'n', 'k' → textLen (likely string param).
    """
    def repl(m):
        prefix = code[:m.start()]
        var_m = re.search(r'(\w+)\s*$', prefix)
        if var_m:
            v = var_m.group(1)
            # Single letter probably string (s, t, w) or ambiguous
            if len(v) == 1 and v.lower() in 'stw':
                return ".textLen()"
            # Plural → list
            if v.endswith('s') and len(v) > 2:
                return ".listLen()"
            # Common list names
            if re.match(r'^(arr|vec|items?|elems?|nums?|list|row|col|grid|'
                        r'board|matrix|path|nodes?|edges?|seq|stack|queue|'
                        r'heap|memo|cache|result|output|buf|chunk)', v, re.I):
                return ".listLen()"
        return ".textLen()"
    return re.sub(r'\.length\(\)', repl, code)


def fix_method_sqrt_floor(code: str) -> str:
    """
    More patterns for method-style math that the regex in pass-1 missed.
    Handles: (a*b + c).floor() type with arithmetic in parens.
    """
    for fn in ['floor', 'ceil', 'round', 'sqrt', 'abs', 'exp', 'log',
               'sin', 'cos', 'tan', 'asin', 'acos', 'atan', 'ln']:
        # Match (anything_balanced).fn()
        # Try up to 3 nesting levels
        for depth in range(3, 0, -1):
            inner = r'[^()]*'
            for _ in range(depth - 1):
                inner = r'(?:[^()]|\(' + inner + r'\))*'
            pat = re.compile(r'\((' + inner + r')\)\.' + fn + r'\(\)')
            real_fn = 'log' if fn == 'ln' else fn
            code = pat.sub(lambda m, f=real_fn: f"{f}({m.group(1)})", code)
        # Also bare identifier
        code = re.sub(r'(\b\w+\b)\.' + fn + r'\(\)',
                      lambda m, f=('log' if fn == 'ln' else fn): f"{f}({m.group(1)})",
                      code)
    return code


def fix_toNumber_inline(code: str) -> str:
    """More patterns for .toNumber() / .parseInt() etc."""
    code = re.sub(r'\(([^)]+)\)\.toNumber\(\)', r'dec(\1)', code)
    code = re.sub(r'\(([^)]+)\)\.toInt\(\)', r'floor(dec(\1))', code)
    code = re.sub(r'(\w+)\.toNumber\(\)', r'dec(\1)', code)
    code = re.sub(r'(\w+)\.toInt\(\)', r'floor(dec(\1))', code)
    code = re.sub(r'(\w+)\.parseFloat\(\)', r'dec(\1)', code)
    code = re.sub(r'(\w+)\.parseInt\(\)', r'floor(dec(\1))', code)
    code = re.sub(r'String\((\w+)\)', r'("" _ \1)', code)
    code = re.sub(r'Number\((\w+)\)', r'dec(\1)', code)
    return code


def fix_ln_method(code: str) -> str:
    """Remaining .ln() cases."""
    code = re.sub(r'\(([^)]+)\)\.ln\(\)', r'log(\1)', code)
    code = re.sub(r'(\w+)\.ln\(\)', r'log(\1)', code)
    return code


ALL_EXTRA_FIXERS = [
    fix_smart_quotes,
    fix_segment_end_form,
    fix_stray_then,
    fix_method_length_aggressive,
    fix_method_sqrt_floor,
    fix_toNumber_inline,
    fix_ln_method,
    fix_nested_list_literal,
]


def apply_all_transforms(code: str) -> str:
    """Apply all transforms from all passes."""
    prev = None
    code = fix_smart_quotes(code)  # do this first
    for _ in range(4):  # up to 4 rounds
        if code == prev:
            break
        prev = code
        for _, fixer in ALL_FIXERS:
            code = fixer(code)
        for fixer in ALL_P2_FIXERS:
            code = fixer(code)
        for fixer in ALL_EXTRA_FIXERS:
            code = fixer(code)
    return code


def main():
    lines = MASTER_REASON.read_text(encoding="utf-8").splitlines(keepends=False)
    while lines and not lines[-1].strip():
        lines.pop()
    total = len(lines)

    backup = MASTER_REASON.with_suffix(".jsonl.bak_prog")
    if not backup.exists():
        shutil.copy(MASTER_REASON, backup)
        print(f"Backup saved to {backup.name}")

    MAX_ROUNDS = 5
    for rnd in range(1, MAX_ROUNDS + 1):
        changed = 0
        newly_fixed = 0
        out_lines = []

        for i, line in enumerate(lines):
            record = json.loads(line)
            messages = record["messages"]
            asst = next((m for m in messages if m["role"] == "assistant"), None)
            if not asst:
                out_lines.append(line)
                continue

            code = extract_code(asst["content"])
            if not code:
                out_lines.append(line)
                continue

            ok_before, _ = check(code)
            if ok_before:
                out_lines.append(line)
                continue

            fixed_code = apply_all_transforms(code)
            ok_after, _ = check(fixed_code)

            if fixed_code != code:
                # Save unconditionally if code changed (even partial fix)
                new_content = replace_code_in_content(asst["content"], fixed_code)
                asst["content"] = new_content
                out_lines.append(json.dumps(record, ensure_ascii=False))
                changed += 1
                if ok_after:
                    newly_fixed += 1
            else:
                out_lines.append(line)

        lines = out_lines
        # Count current pass/fail
        passing = sum(1 for l in lines
                      if (lambda c: c is not None and check(c)[0])(
                          extract_code(json.loads(l)["messages"][-1]["content"])
                          if json.loads(l)["messages"] else None
                      ))
        print(f"Round {rnd}: changed={changed}  newly_fixed={newly_fixed}  "
              f"total_passing≈{1661+newly_fixed*(rnd)}")

        if changed == 0:
            print("  Stable — no more changes possible.")
            break

    # Final count
    final_pass = 0
    final_fail = 0
    for line in lines:
        rec = json.loads(line)
        asst = next((m for m in rec["messages"] if m["role"] == "assistant"), None)
        if not asst:
            continue
        code = extract_code(asst["content"])
        if not code:
            continue
        ok, _ = check(code)
        if ok:
            final_pass += 1
        else:
            final_fail += 1

    MASTER_REASON.write_text("\n".join(lines) + "\n", encoding="utf-8")

    print()
    print("=" * 55)
    print(f"Total entries      : {total}")
    print(f"Final passing      : {final_pass}  ({100*final_pass/total:.1f}%)")
    print(f"Final failing      : {final_fail}  ({100*final_fail/total:.1f}%)")


if __name__ == "__main__":
    main()
