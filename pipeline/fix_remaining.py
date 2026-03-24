#!/usr/bin/env python3
"""
Targeted fixer: scan failing entries, identify each error, apply all
applicable fixes (compound errors need multiple fixers), save if code changed
(even if still failing — enables multi-round progress).

Bugs fixed vs previous run:
- fix_range_for_in: was dropping the closing ')' in replacement
- Now applies ALL relevant fixers to each entry (not just one)
- New: fix_if_inline_assign — if (COND) VAR = EXPR → if (COND) { VAR = EXPR }
"""

import json
import re
import shutil
import subprocess
from pathlib import Path

FALCON_BIN = Path(__file__).parent.parent / "lang" / "Falcon"
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
    think_end = content.find("</think>")
    if think_end == -1:
        return re.sub(r"```falcon\s*.*?\s*```", f"```falcon\n{new_code}\n```",
                      content, flags=re.DOTALL)
    prefix = content[:think_end + len("</think>")]
    suffix = content[think_end + len("</think>"):]
    new_suffix = re.sub(r"```falcon\s*.*?\s*```", f"```falcon\n{new_code}\n```",
                        suffix, flags=re.DOTALL)
    return prefix + new_suffix


# ─── Fixers ───────────────────────────────────────────────────────────────────

def fix_underscore_vars(code: str) -> str:
    specific = {
        '_mat_out': 'matOut', '_rgb_color': 'rgbColor', '_rgb_a': 'rgbA',
        '_rgb_c': 'rgbC', '_rgb_': 'rgb', '_ri': 'ri', '_i': 'loopI',
        '_r': 'loopR', '_l': 'loopL', '_mk': 'mkList',
    }
    for old, new in specific.items():
        code = code.replace(old, new)
    code = re.sub(r'\b_([a-zA-Z])', lambda m: m.group(1).upper(), code)
    return code


def fix_range_for_in(code: str) -> str:
    """for (VAR in EXPR..EXPR) → for (VAR: EXPR..EXPR)  — note: restore closing )"""
    def repl(m):
        var, expr = m.group(1), m.group(2)
        if '..' in expr:
            return f'for ({var}: {expr})'   # ← closing ) restored
        return m.group(0)
    return re.sub(r'for\s*\(\s*(\w+)\s+in\s+([^)\n]+\.\.[^)\n]*)\)', repl, code)


def fix_segmentdec(code: str) -> str:
    code = re.sub(r'(\w+)\.segmentdec\(([^)]+)\)', r'dec(\1.segment(\2))', code)
    code = re.sub(r'(\w+)\.adddec\(([^)]+)\)', r'\1.add(dec(\2))', code)
    return code


def fix_step_keyword(code: str) -> str:
    """Rename 'step' as identifier; keep 'step N' in range loops."""
    def repl(m: re.Match) -> str:
        after = code[m.end():]
        if re.match(r'\s+[-\d]', after):
            return 'step'
        return 'stepVar'
    return re.sub(r'\bstep\b', repl, code)


def fix_if_no_parens(code: str) -> str:
    """if COND { → if (COND) {  when cond doesn't start with ("""
    lines = code.split('\n')
    out = []
    for line in lines:
        line = re.sub(
            r'\bif\s+(?!\()([^{}\n]+?)\s*\{',
            lambda m: f'if ({m.group(1).strip()}) {{',
            line
        )
        out.append(line)
    return '\n'.join(out)


def fix_while_no_parens(code: str) -> str:
    """while COND { → while (COND) {"""
    lines = code.split('\n')
    out = []
    for line in lines:
        line = re.sub(
            r'\bwhile\s+(?!\()([^{}\n]+?)\s*\{',
            lambda m: f'while ({m.group(1).strip()}) {{',
            line
        )
        out.append(line)
    return '\n'.join(out)


def fix_if_inline_assign(code: str) -> str:
    """
    if (COND) VAR = EXPR  →  if (COND) { VAR = EXPR }
    Falcon doesn't allow bare assignments as if-body (without braces).
    """
    lines = code.split('\n')
    out = []
    for line in lines:
        # Match: if (COND) WORD = EXPR  (not already followed by {)
        m = re.match(
            r'^(\s*(?:(?:else\s+)?if\s*\([^)]+\))\s+)(\w+\s*=\s*[^=].*)',
            line
        )
        if m and not m.group(2).startswith('{'):
            line = m.group(1) + '{ ' + m.group(2).rstrip() + ' }'
        out.append(line)
    return '\n'.join(out)


def fix_pow_parens(code: str) -> str:
    """(EXPR).pow(N) → ((EXPR) ^ (N))"""
    for _ in range(3):
        prev = code
        code = re.sub(r'\(([^()]+)\)\.pow\(([^)]+)\)', r'((\1) ^ (\2))', code)
        if code == prev:
            break
    return code


def fix_format_method(code: str) -> str:
    """x.format(N) → formatDecimal(x, N)"""
    code = re.sub(r'(\w+)\.format\((\d+)\)', r'formatDecimal(\1, \2)', code)
    code = re.sub(r'\(([^)]+)\)\.format\((\d+)\)', r'formatDecimal((\1), \2)', code)
    return code


def fix_binary_literals(code: str) -> str:
    """0b1010 → bin("1010")"""
    return re.sub(r'0b([01]+)', r'bin("\1")', code)


def fix_bare_list_return(code: str) -> str:
    """
    A line starting with '[' not as a continuation is parsed as subscript.
    Wrap as:  local retList = [...]\n retList
    """
    lines = code.split('\n')
    out = []
    i = 0
    while i < len(lines):
        line = lines[i]
        stripped = line.lstrip()
        indent = line[:len(line) - len(stripped)]

        if stripped.startswith('[') and stripped != '[]':
            prev_line = ''
            for j in range(i - 1, -1, -1):
                if lines[j].strip():
                    prev_line = lines[j].rstrip()
                    break
            continuation = re.search(
                r'[=,\[\({+\-\*/&|_\\]\s*$|->$|\belse\s*$', prev_line
            )
            if not continuation and prev_line:
                depth = 0
                list_lines = []
                j = i
                while j < len(lines):
                    list_lines.append(lines[j])
                    for ch in lines[j]:
                        if ch == '[':
                            depth += 1
                        elif ch == ']':
                            depth -= 1
                    if depth == 0:
                        break
                    j += 1

                if depth == 0 and j < len(lines):
                    raw = '\n'.join(list_lines).strip()
                    out.append(indent + 'local retList = ' + raw)
                    out.append(indent + 'retList')
                    i = j + 1
                    continue

        out.append(line)
        i += 1
    return '\n'.join(out)


def fix_if_partial_paren_then(code: str) -> str:
    """if (PARTIAL) REST then BODY → if ((PARTIAL) REST) BODY"""
    lines = code.split('\n')
    out = []
    for line in lines:
        m = re.match(
            r'^(\s*(?:else\s+)?if\s+)(\((?:[^()]+|\([^()]*\))*\))(\s+[^{}\n]+?)\s+then\s+(.*)',
            line
        )
        if m:
            prefix, paren_part, rest, body = (
                m.group(1), m.group(2), m.group(3).strip(), m.group(4)
            )
            if rest:
                line = f'{prefix}({paren_part} {rest}) {body}'
        out.append(line)
    return '\n'.join(out)


def fix_invalid_func_params(code: str) -> str:
    """func f("" _ x) → func f(x) and matching call sites."""
    code = re.sub(r'func\s+(\w+)\s*\(""\s*_\s*(\w+)\)',
                  lambda m: f'func {m.group(1)}({m.group(2)})', code)
    code = re.sub(r'(\w+)\(""\s*_\s*(\w+)\)',
                  lambda m: f'{m.group(1)}({m.group(2)})', code)
    return code


# ─── Error classification ─────────────────────────────────────────────────────

def has_error(err: str, code: str, pattern_err=None, pattern_code=None) -> bool:
    if pattern_err and re.search(pattern_err, err):
        return True
    if pattern_code and re.search(pattern_code, code):
        return True
    return False


def applicable_fixers(err: str, code: str) -> list:
    """Return list of fixers relevant to this error + code."""
    fixers = []

    if re.search(r'\b_\w+', code):
        fixers.append(fix_underscore_vars)

    if re.search(r'for\s*\(\s*\w+\s+in\s+[^)]*\.\.', code):
        fixers.append(fix_range_for_in)

    if 'segmentdec' in err or 'adddec' in code:
        fixers.append(fix_segmentdec)

    if re.search(r'\bstep\b', code) and not re.search(r'for\s*\(\w+\s*:\s*[^)]+step\s+\d', code):
        # Only rename if step appears as identifier (not purely as keyword in a working loop)
        if re.search(r'(func\s+step\b|local\s+step\b|for\s*\(\s*step[\s:]|,\s*step\)|\(step\))', code):
            fixers.append(fix_step_keyword)

    if re.search(r'Unexpected!\s*\(Else else\)', err) or re.search(r'Unexpected!\s*\(Dot \.\)', err) or re.search(r'Unexpected!\s*\(Assign', err):
        fixers.append(fix_if_no_parens)
        fixers.append(fix_while_no_parens)
        fixers.append(fix_if_inline_assign)

    if re.search(r'\bwhile\s+(?!\()', code):
        fixers.append(fix_while_no_parens)

    if re.search(r'\bif\s+(?!\()', code):
        fixers.append(fix_if_no_parens)
        fixers.append(fix_if_inline_assign)

    if re.search(r'Expected type Cl', err) or re.search(r'Unexpected!\s*\(CloseSquare', err):
        fixers.append(fix_bare_list_return)

    if re.search(r'\.pow\(\)', err):
        fixers.append(fix_pow_parens)

    if re.search(r'\.format\(\)', err):
        fixers.append(fix_format_method)

    if re.search(r'0b[01]+', code):
        fixers.append(fix_binary_literals)

    if re.search(r'func\s+\w+\s*\(""\s*_\s*\w+\)', code):
        fixers.append(fix_invalid_func_params)

    if re.search(r'\((?:[^()]+|\([^()]*\))*\)\s+\S.*?\s+then\b', code):
        fixers.append(fix_if_partial_paren_then)

    # Deduplicate preserving order
    seen = set()
    result = []
    for f in fixers:
        if f not in seen:
            seen.add(f)
            result.append(f)
    return result


def apply_targeted(code: str, err: str) -> str:
    fixers = applicable_fixers(err, code)
    for fixer in fixers:
        code = fixer(code)
    return code


# ─── main ─────────────────────────────────────────────────────────────────────

def main():
    lines = MASTER_REASON.read_text(encoding="utf-8").splitlines(keepends=False)
    while lines and not lines[-1].strip():
        lines.pop()
    total = len(lines)

    backup = MASTER_REASON.with_suffix(".jsonl.bak_rem")
    if not backup.exists():
        shutil.copy(MASTER_REASON, backup)
        print(f"Backup saved to {backup.name}")

    print("Scanning for failing entries...")
    failing = []
    for i, line in enumerate(lines):
        record = json.loads(line)
        asst = next((m for m in record["messages"] if m["role"] == "assistant"), None)
        if not asst:
            continue
        code = extract_code(asst["content"])
        if not code:
            continue
        ok, err = check(code)
        if not ok:
            failing.append((i, err, code))

    print(f"Found {len(failing)} failing entries.")

    changed = 0
    newly_fixed = 0
    out_lines = list(lines)

    for idx, err, code in failing:
        fixed = apply_targeted(code, err)

        if fixed == code:
            continue

        ok, _ = check(fixed)

        record = json.loads(out_lines[idx])
        asst = next(m for m in record["messages"] if m["role"] == "assistant")
        asst["content"] = replace_code_in_content(asst["content"], fixed)
        out_lines[idx] = json.dumps(record, ensure_ascii=False)
        changed += 1
        if ok:
            newly_fixed += 1

    MASTER_REASON.write_text("\n".join(out_lines) + "\n", encoding="utf-8")

    print()
    print("=" * 55)
    print(f"Total failing      : {len(failing)}")
    print(f"Changed entries    : {changed}  (saved unconditionally for compound errors)")
    print(f"Newly passing      : {newly_fixed}")
    print(f"Est. pass rate     ≈ {(total - len(failing) + newly_fixed) / total * 100:.1f}%")


if __name__ == "__main__":
    main()
