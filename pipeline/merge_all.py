#!/usr/bin/env python3
"""Merge all COMBINED*.json files into answers/MASTER.json.

Applies a series of fixers to each answer before writing:
  1. Strip semicolons (not valid Falcon syntax)
  2. Add {} around single-line for/else bodies
  3. Rename 'step' loop variable (reserved keyword)
  4. Convert inline func(x)={body} lambdas to named functions

Usage:
    python3 pipeline/merge_all.py
"""

import json
import re
from pathlib import Path

ANSWERS_DIR = Path("answers")
OUTPUT = ANSWERS_DIR / "MASTER.json"


# ---------------------------------------------------------------------------
# Individual fixers
# ---------------------------------------------------------------------------

def strip_semicolons(code: str) -> str:
    return code.replace(";", "")


def _find_close_paren(s: str, start: int) -> int:
    """Return index of the ) that matches the ( at s[start], or -1."""
    depth = 0
    for i in range(start, len(s)):
        if s[i] == '(':
            depth += 1
        elif s[i] == ')':
            depth -= 1
            if depth == 0:
                return i
    return -1


def fix_missing_braces(code: str) -> str:
    """Add {} around single-line for/else bodies that lack them.

    Handles:
      for (x in list) stmt        →  for (x in list) { stmt }
      for (i: 1..n) stmt          →  for (i: 1..n) { stmt }
      } else stmt                 →  } else { stmt }
      else stmt  (line-start)     →  else { stmt }
    """
    lines = code.split('\n')
    out = []
    for line in lines:
        stripped = line.lstrip()
        indent = line[:len(line) - len(stripped)]

        # --- for/while (...) body without braces ---
        m = re.match(r'(for|while)\s*\(', stripped)
        if m:
            paren_open = m.end() - 1  # position of '(' in stripped
            close = _find_close_paren(stripped, paren_open)
            if close != -1:
                after = stripped[close + 1:].lstrip()
                if after and not after.startswith('{'):
                    line = indent + stripped[:close + 1] + ' { ' + after + ' }'

        # --- } else <non-if, non-brace> body on same line ---
        m2 = re.match(r'^(.*\})\s+else\s+(?!if\b)(?!\{)(.+)$', line)
        if m2:
            line = m2.group(1) + ' else { ' + m2.group(2).strip() + ' }'

        # --- standalone else <non-if, non-brace> at line start ---
        m3 = re.match(r'^(\s*else)\s+(?!if\b)(?!\{)(.+)$', line)
        if m3 and '}'.join(line.split('}')[:-1]).count('}') == 0:
            # only if there's no closing } before the else on this line
            line = m3.group(1) + ' { ' + m3.group(2).strip() + ' }'

        out.append(line)
    return '\n'.join(out)


def _non_string_parts(code: str):
    """Yield (start, end, is_string) tuples splitting code at string literals."""
    pattern = re.compile(r'"(?:[^"\\]|\\.)*"')
    prev = 0
    for m in pattern.finditer(code):
        if m.start() > prev:
            yield prev, m.start(), False
        yield m.start(), m.end(), True
        prev = m.end()
    if prev < len(code):
        yield prev, len(code), False


def _replace_outside_strings(code: str, pattern: str, repl) -> str:
    """Apply re.sub(pattern, repl) only to non-string-literal portions of code."""
    parts = []
    for start, end, is_str in _non_string_parts(code):
        chunk = code[start:end]
        if not is_str:
            chunk = re.sub(pattern, repl, chunk)
        parts.append(chunk)
    return ''.join(parts)


def fix_step_variable(code: str) -> str:
    """Rename 'step' identifier everywhere it's used as a name (it's a reserved keyword).

    Preserves 'step N' keyword usage like: for (i: 1..10 step 2)
    Skips occurrences inside string literals.
    """
    if not re.search(r'\bstep\b', code):
        return code
    code = _replace_outside_strings(code, r'\bstep\b', 'stepVar')
    # Restore keyword form: '.. end step N' where N is a (possibly negative) number
    code = _replace_outside_strings(code, r'\bstepVar\b(?=\s+-?\d)', 'step')
    return code


def fix_hex_literals(code: str) -> str:
    """Replace hex literals like 0x1A with their decimal equivalents.

    Falcon has no hex literal syntax; 0x1A is parsed as integer 0, then identifier x1A.
    """
    def to_dec(m: re.Match) -> str:
        try:
            return str(int(m.group(0), 16))
        except ValueError:
            return m.group(0)
    return _replace_outside_strings(code, r'0[xX][0-9a-fA-F]+', to_dec)


def fix_scientific_notation(code: str) -> str:
    """Replace scientific notation (e.g. 1e-6, 2.5e3) with plain decimal literals.

    Falcon has no scientific notation support; the 'e' is parsed as a variable name.
    """
    def expand(m: re.Match) -> str:
        try:
            val = float(m.group(0))
            # Format without scientific notation
            formatted = f'{val:.20f}'.rstrip('0').rstrip('.')
            if '.' not in formatted and 'e' not in formatted:
                return formatted
            return formatted
        except ValueError:
            return m.group(0)

    return _replace_outside_strings(code, r'\d+(?:\.\d+)?[eE][+-]?\d+', expand)


def fix_standalone_list_literal(code: str) -> str:
    """Wrap multi-element list literals used as standalone expressions.

    When [a, b, c] appears at the start of an expression (not an assignment),
    the parser treats [ as a subscript operator and fails on the comma.
    Fix: local listResult = [a, b, c]  followed by  listResult
    """
    lines = code.split('\n')
    out = []
    # Track unclosed '[' from previous lines to skip continuation lines
    open_sq = 0
    for line in lines:
        stripped = line.lstrip()
        indent = line[:len(line) - len(stripped)]

        # Count bracket balance in stripped line (ignoring strings) to detect multi-line literals
        sq_open = sum(1 for c in re.sub(r'"[^"]*"', '', stripped) if c == '[')
        sq_close = sum(1 for c in re.sub(r'"[^"]*"', '', stripped) if c == ']')
        balanced = sq_open == sq_close

        if (open_sq == 0
                and balanced
                and stripped.startswith('[')
                and not re.match(r'(local|global)\s+\w+\s*=\s*\[', stripped)):
            # Check for a top-level comma (multi-element literal)
            depth = 0
            has_comma = False
            in_str = False
            for c in stripped:
                if c == '"' and not in_str:
                    in_str = True
                elif c == '"' and in_str:
                    in_str = False
                elif not in_str:
                    if c == '[':
                        depth += 1
                    elif c == ']':
                        depth -= 1
                    elif c == ',' and depth == 1:
                        has_comma = True
                        break
            if has_comma:
                line = indent + 'local listResult = ' + stripped + '\n' + indent + 'listResult'

        # Update running bracket depth for continuation detection
        in_str = False
        for c in line:
            if c == '"' and not in_str:
                in_str = True
            elif c == '"' and in_str:
                in_str = False
            elif not in_str:
                if c == '[':
                    open_sq += 1
                elif c == ']':
                    open_sq = max(0, open_sq - 1)

        out.append(line)
    return '\n'.join(out)


def fix_inline_func_lambdas(code: str) -> str:
    """Convert inline anonymous func(params)={body} to named top-level functions.

    Uses a brace-balanced parser to handle nested {} in the body correctly.
    func(x)={x*2}  →  func fn1(x) = { x*2 }  (prepended, call site gets fn1)
    """
    defs = []
    result = []
    fn_count = [0]
    i = 0

    while i < len(code):
        # Match 'func(' that is NOT a named function (not preceded by space+name)
        m = re.match(r'func\s*\(', code[i:])
        if m and (i == 0 or not (code[i - 1].isalnum() or code[i - 1] == '_')):
            paren_open = i + m.end() - 1
            close_paren = _find_close_paren(code, paren_open)
            if close_paren != -1:
                after = code[close_paren + 1:]
                m2 = re.match(r'\s*=\s*\{', after)
                if m2:
                    body_start = close_paren + 1 + m2.end()
                    # Find balanced closing }
                    depth = 1
                    j = body_start
                    while j < len(code) and depth > 0:
                        if code[j] == '{':
                            depth += 1
                        elif code[j] == '}':
                            depth -= 1
                        j += 1
                    body = code[body_start:j - 1].strip()
                    params = code[paren_open + 1:close_paren].strip()
                    fn_count[0] += 1
                    name = f'fn{fn_count[0]}'
                    defs.append(f'func {name}({params}) = {{ {body} }}')
                    result.append(name)
                    i = j
                    continue
        result.append(code[i])
        i += 1

    code = ''.join(result)
    if defs:
        code = '\n'.join(defs) + '\n' + code
    return code


def fix_lambda_args(code: str) -> str:
    """Convert { x -> body } used as a function call argument to a named function.

    Uses a brace-balanced parser to handle nested {} in the body correctly.
    Falcon's { x -> ... } syntax only works as a method-chain lambda (.map, .filter, etc.).
    """
    defs = []
    result = []
    lam_count = [0]
    i = 0

    while i < len(code):
        # Look for pattern: preceded by ( or ,  then optional whitespace then {
        if code[i] == '{' and i > 0:
            # Check what came before (skip whitespace)
            j = i - 1
            while j >= 0 and code[j] == ' ':
                j -= 1
            if j >= 0 and code[j] in '(,':
                # Check for lambda syntax: { params -> body }
                inner_start = i + 1
                # Find matching }
                depth = 1
                k = inner_start
                while k < len(code) and depth > 0:
                    if code[k] == '{':
                        depth += 1
                    elif code[k] == '}':
                        depth -= 1
                    k += 1
                inner = code[inner_start:k - 1].strip()
                lm = re.match(r'^(\w+(?:\s*,\s*\w+)*)\s*->\s*(.+)$', inner, re.DOTALL)
                if lm:
                    params = lm.group(1).strip()
                    body = lm.group(2).strip()
                    lam_count[0] += 1
                    name = f'lam{lam_count[0]}'
                    defs.append(f'func {name}({params}) = {{ {body} }}')
                    result.append(' ' + name)
                    i = k
                    continue
        result.append(code[i])
        i += 1

    code = ''.join(result)
    if defs:
        code = '\n'.join(defs) + '\n' + code
    return code


def fix_if_unary_minus(code: str) -> str:
    """Fix: if (cond) -expr else rest  →  if (cond) { -expr } else { rest }

    The parser fails to parse an else clause when the if-branch starts with unary minus.
    """
    lines = code.split('\n')
    out = []
    for line in lines:
        m = re.search(r'\bif\s*\(', line)
        if m:
            close = _find_close_paren(line, m.end() - 1)
            if close != -1:
                after = line[close + 1:].lstrip()
                if after.startswith('-') and re.search(r'\belse\b', after):
                    else_m = re.search(r'\belse\b', after)
                    if_branch = after[:else_m.start()].strip()
                    else_rest = after[else_m.end():].strip()
                    before = line[:m.start()]
                    cond = line[m.start():close + 1]
                    # Only wrap else_rest if it doesn't already have braces
                    else_part = else_rest if else_rest.startswith('{') else '{ ' + else_rest + ' }'
                    line = before + cond + ' { ' + if_branch + ' } else ' + else_part
        out.append(line)
    return '\n'.join(out)


def fix_dict_get(code: str) -> str:
    """Add missing default to .get(key) — Falcon requires .get(key, default)."""
    def replacer(m: re.Match) -> str:
        content = m.group(1)
        # Check if there's already a top-level comma (meaning 2+ args already)
        depth = 0
        for c in content:
            if c in '([':
                depth += 1
            elif c in ')]':
                depth -= 1
            elif c == ',' and depth == 0:
                return m.group(0)  # already has default
        return f'.get({content}, [])'

    # Handle one level of nested parens inside the arg
    return re.sub(r'\.get\(([^)]*(?:\([^)]*\)[^)]*)*)\)', replacer, code)


def fix_multilocal(code: str) -> str:
    """Split multiple local/global declarations crammed onto one line.

    'local x = expr local y = expr2'  →  'local x = expr\nlocal y = expr2'
    """
    # Insert newline before a 'local'/'global' that follows a non-newline character
    return re.sub(r'(?<!\n)([ \t]+)(local|global)(\s+\w)', r'\n\2\3', code)


def fix_code(code: str) -> str:
    code = strip_semicolons(code)
    code = fix_hex_literals(code)
    code = fix_scientific_notation(code)
    code = fix_missing_braces(code)
    code = fix_step_variable(code)
    code = fix_inline_func_lambdas(code)
    code = fix_lambda_args(code)
    code = fix_if_unary_minus(code)
    code = fix_dict_get(code)
    code = fix_standalone_list_literal(code)
    code = fix_multilocal(code)
    return code


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

def main() -> None:
    combined_files = sorted(ANSWERS_DIR.glob("COMBINED*.json"))
    if not combined_files:
        print("No COMBINED*.json files found in answers/")
        return

    merged: dict = {}
    for path in combined_files:
        with open(path, encoding="utf-8") as f:
            data = json.load(f)
        for pid, entry in data.items():
            if "answer" in entry:
                entry["answer"] = fix_code(entry["answer"])
            merged[pid] = entry
        print(f"  Loaded {len(data):>5} entries from {path.name}")

    with open(OUTPUT, "w", encoding="utf-8") as f:
        json.dump(merged, f, indent=2, ensure_ascii=False)

    print(f"\nWrote {len(merged)} entries to {OUTPUT}")


if __name__ == "__main__":
    main()