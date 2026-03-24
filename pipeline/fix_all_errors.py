#!/usr/bin/env python3
"""
Comprehensive fixer for MASTER_REASON.jsonl syntax errors.
Applies fixes in order; re-checks after each round until stable or no progress.

Fixes applied (in order of application):
  F1. .length()    → .textLen() or .listLen() (heuristic)
  F2. .append(x)   → .add(x)
  F3. .contains(x) → .containsItem(x)
  F4. for (x: VAR) → for (x in VAR)   [each-loop colon form]
  F5. print(       → println(
  F6. .floor() / .ceil() / .round() / .sqrt() / .abs() / .exp() / .log()
                   → standalone function calls
  F7. if COND then → if (COND)        [remove 'then', add parens]
  F8. semicolons   → newline
  F9. .red()/.green()/.blue() on a color → splitColor()  [best-effort]
  F10. while COND do → while (COND)
  F11. .size()     → .listLen() or .textLen()
"""

import json
import subprocess
import re
import shutil
from pathlib import Path

FALCON_BIN = Path(__file__).parent.parent / "lang" / "Falcon"
MASTER_REASON = Path(__file__).parent.parent / "answer_reasoning" / "MASTER_REASON.jsonl"

# ---------- parser ----------

def check(code: str) -> tuple[bool, str]:
    result = subprocess.run(
        [str(FALCON_BIN), "format"],
        input=code, capture_output=True, text=True, timeout=10,
    )
    return result.returncode == 0, result.stderr.strip()


# ---------- code extraction / replacement ----------

def extract_code(content: str) -> str | None:
    no_think = re.sub(r"<think>.*?</think>", "", content, flags=re.DOTALL).strip()
    m = re.search(r"```falcon\s*(.*?)\s*```", no_think, re.DOTALL)
    return m.group(1).strip() if m else None


def replace_code_in_content(content: str, new_code: str) -> str:
    think_end = content.find("</think>")
    if think_end == -1:
        return re.sub(r"```falcon\s*.*?\s*```", f"```falcon\n{new_code}\n```", content, flags=re.DOTALL)
    prefix = content[:think_end + len("</think>")]
    suffix = content[think_end + len("</think>"):]
    new_suffix = re.sub(r"```falcon\s*.*?\s*```", f"```falcon\n{new_code}\n```", suffix, flags=re.DOTALL)
    return prefix + new_suffix


# ---------- individual fixers ----------

LIST_VAR = re.compile(
    r"^(list|arr|items|nums|numbers|elems|elements|result|parts|tokens|"
    r"words|rows|cols|stack|queue|heap|entries|records|keys|values|pairs|"
    r"data|path|perm|choices|pool|bucket|counts|scores|bits|bytes|digits|"
    r"chars|ops|instrs|instructions|signals|samples|neighbors|children|"
    r"nodes|edges|seq|sequence|matrix|grid|row|col|board|line|lines|"
    r"sorted|filtered|mapped|chunk|chunks|subset|groups|table|pages|"
    r"memo|cache|hist|histogram|freq|freqs|weights|probs|deck|hand|"
    r"run|runs|cluster|clusters|candidates|output|input|pq|queue2|"
    r"adj|graph|visited|parent|dist|order|topo|components|uf|dsu|"
    r"prefix|suffix|res|acc|buf|buffer|out|all|many|some|few)$",
    re.IGNORECASE
)


def smart_length(code: str) -> str:
    def repl(m):
        prefix = code[:m.start()]
        var_m = re.search(r'(\w+)\s*$', prefix)
        if var_m and LIST_VAR.match(var_m.group(1)):
            return ".listLen()"
        return ".textLen()"
    return re.sub(r'\.length\(\)', repl, code)


def smart_size(code: str) -> str:
    def repl(m):
        prefix = code[:m.start()]
        var_m = re.search(r'(\w+)\s*$', prefix)
        if var_m and LIST_VAR.match(var_m.group(1)):
            return ".listLen()"
        return ".textLen()"
    return re.sub(r'\.size\(\)', repl, code)


def fix_append(code: str) -> str:
    return re.sub(r'\.append\(', '.add(', code)


def fix_contains(code: str) -> str:
    # .contains(x) → .containsItem(x)  but NOT .containsKey() or .containsAny() etc.
    return re.sub(r'\.contains\((?!Key|Any|All|Item)', '.containsItem(', code)


def fix_for_colon(code: str) -> str:
    """
    for (VAR: EXPR)  →  for (VAR in EXPR)
    Only when EXPR is an identifier or method chain, NOT a number.
    We detect: for (word: number) which is range — leave those alone.
    """
    def repl(m):
        var = m.group(1)
        space = m.group(2)
        expr = m.group(3)
        # If the expression after : looks like it starts with a digit → range loop, leave it
        if re.match(r'^\s*\d', expr):
            return m.group(0)
        return f"for ({var} in{space}{expr}"
    # Match: for ( VARNAME : [non-digit] ...
    return re.sub(r'for\s*\(\s*(\w+)\s*:(\s*)(?!\d)(\S)', repl, code)


def fix_print(code: str) -> str:
    # print( → println(  but not println( itself
    return re.sub(r'\bprint\s*\(', 'println(', code)


def fix_method_math(code: str) -> str:
    """
    Convert   EXPR.floor()  →  floor(EXPR)
    Handles simple identifiers and (balanced) parenthesized expressions.
    """
    MATH_METHODS = ['floor', 'ceil', 'round', 'sqrt', 'abs', 'exp', 'log',
                    'sin', 'cos', 'tan', 'asin', 'acos', 'atan']

    for fn in MATH_METHODS:
        pattern = re.compile(r'(\w+|\([^()]*(?:\([^()]*\)[^()]*)*\))\.' + fn + r'\(\)')
        while True:
            new_code = pattern.sub(lambda m, f=fn: f"{f}({m.group(1)})", code)
            if new_code == code:
                break
            code = new_code
    return code


def fix_if_then(code: str) -> str:
    """
    if COND then EXPR  →  if (COND) EXPR
    Also handles  if (COND) then EXPR  →  if (COND) EXPR
    """
    # Remove 'then' that appears after if (...)
    code = re.sub(r'(if\s*\([^)]+\))\s+then\b', r'\1', code)
    # Convert: if COND then  (where COND has no leading paren)
    # Heuristic: grab until we hit 'then'
    # We use a line-by-line approach to avoid cross-line greediness issues
    lines = code.split('\n')
    out = []
    for line in lines:
        # Match: if <non-paren start> ... then
        # but only when 'if' is not already followed by '('
        line = re.sub(
            r'\bif\s+(?!\()([^{}\n]+?)\s+then\b',
            lambda m: f"if ({m.group(1).strip()})",
            line
        )
        out.append(line)
    return '\n'.join(out)


def fix_while_do(code: str) -> str:
    """while COND do { → while (COND) {"""
    code = re.sub(r'\bwhile\s+(?!\()([^{]+?)\s+do\b', lambda m: f"while ({m.group(1).strip()})", code)
    # Remove stray 'do' after while (COND)
    code = re.sub(r'(while\s*\([^)]+\))\s+do\b', r'\1', code)
    return code


def fix_semicolons(code: str) -> str:
    """Replace ; with newline (outside of strings)."""
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
        else:
            if c in ('"', "'"):
                in_str = True
                str_char = c
                result.append(c)
            elif c == ';':
                result.append('\n')
            else:
                result.append(c)
        i += 1
    return ''.join(result)


def fix_color_methods(code: str) -> str:
    """
    Replace .red() / .green() / .blue() on a color variable with splitColor().
    Pattern: local r = color.red()  →  local rgb = splitColor(color)\n  local r = rgb[1]
    This is a best-effort transformation; may not always be correct.
    """
    # Simple pattern: local VAR = COLORVAR.red()
    code = re.sub(
        r'local\s+(\w+)\s*=\s*(\w+)\.red\(\)',
        lambda m: f"local _rgb_{m.group(2)} = splitColor({m.group(2)})\n  local {m.group(1)} = _rgb_{m.group(2)}[1]",
        code
    )
    code = re.sub(
        r'local\s+(\w+)\s*=\s*(\w+)\.green\(\)',
        lambda m: f"local {m.group(1)} = _rgb_{m.group(2)}[2]",
        code
    )
    code = re.sub(
        r'local\s+(\w+)\s*=\s*(\w+)\.blue\(\)',
        lambda m: f"local {m.group(1)} = _rgb_{m.group(2)}[3]",
        code
    )
    # Also: COLORVAR.red() used inline
    code = re.sub(r'(\w+)\.red\(\)', r'splitColor(\1)[1]', code)
    code = re.sub(r'(\w+)\.green\(\)', r'splitColor(\1)[2]', code)
    code = re.sub(r'(\w+)\.blue\(\)', r'splitColor(\1)[3]', code)
    return code


def fix_repeat(code: str) -> str:
    """
    [INIT].repeat(N)  is not valid in Falcon.
    Replace with a loop-generated list. Best effort for simple cases.
    Pattern: [VALUE].repeat(EXPR)  →  manual expansion (skip, too complex)
    Just note: we can't easily fix this without knowing types.
    Instead, use a helper pattern common in these entries.
    """
    # Common pattern: local dp = [false].repeat(n + 1)
    # Falcon way: build with a loop or use list literal approach
    # For now: replace with a known pattern that generates a list of N copies
    def repl(m):
        val = m.group(1).strip()
        n_expr = m.group(2).strip()
        return f"(func _mk() = {{ local _l = []\n  for (_i: 1 .. {n_expr}) {{ _l.add({val}) }}\n  _l }})()"
    code = re.sub(r'\[([^\]]+)\]\.repeat\(([^)]+)\)', repl, code)
    return code


ALL_FIXERS = [
    ("length",       smart_length),
    ("size",         smart_size),
    ("append",       fix_append),
    ("contains",     fix_contains),
    ("for-colon",    fix_for_colon),
    ("print",        fix_print),
    ("method-math",  fix_method_math),
    ("if-then",      fix_if_then),
    ("while-do",     fix_while_do),
    ("semicolons",   fix_semicolons),
    ("color-methods",fix_color_methods),
]


def apply_fixes(code: str) -> str:
    for _, fixer in ALL_FIXERS:
        code = fixer(code)
    return code


# ---------- main ----------

def main():
    lines = MASTER_REASON.read_text(encoding="utf-8").splitlines(keepends=False)
    # strip trailing empty lines
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

        # Apply all fixes
        fixed_code = apply_fixes(code)
        ok2, _ = check(fixed_code)
        if ok2 and fixed_code != code:
            new_content = replace_code_in_content(asst["content"], fixed_code)
            asst["content"] = new_content
            out_lines.append(json.dumps(record, ensure_ascii=False))
            fixed_count += 1
        else:
            out_lines.append(line)
            if not ok:
                still_failing += 1

    backup = MASTER_REASON.with_suffix(".jsonl.bak_all")
    if not backup.exists():
        shutil.copy(MASTER_REASON, backup)
        print(f"Backup saved to {backup.name}")

    MASTER_REASON.write_text("\n".join(out_lines) + "\n", encoding="utf-8")

    print()
    print("=" * 55)
    print(f"Total entries      : {total}")
    print(f"Already passing    : {already_ok}")
    print(f"Fixed by script    : {fixed_count}")
    print(f"Still failing      : {still_failing}")
    print(f"New pass rate      : {(already_ok + fixed_count) / total * 100:.1f}%")


if __name__ == "__main__":
    main()
