#!/usr/bin/env python3
"""
Patch MASTER.jsonl entries failing with "result of X is never used".

Strategy:
  - Parse the error to find the offending expression (e.g. `heapPush(myHeap, 3)`)
  - If it's a user-defined function call, extract the function name
  - Strip `=` only from THAT specific function's declaration
  - If the body then has a trailing consumable bare-var return, remove it too
  - Verify fix passes before committing

Also handles:
  - Double-wrapped SmartBody: func f() = { { expr } } -> func f() = { expr }
"""

import json, re, shutil, subprocess, sys, tempfile
from pathlib import Path

MASTER = Path("finetune/dataset/1.8k-codes-steps-output/MASTER.jsonl")
BACKUP = MASTER.with_suffix(".jsonl.bak_never_used")
FALCON = "lang/Falcon"

DOUBLE_WRAP_RE = re.compile(
    r'(func\s+\w+\s*\([^)]*\)\s*=\s*\n\s*\{)\n(\s*)\{([^{}\n]*)\}\s*\n(\s*\})',
    re.MULTILINE
)

def fix_double_wrap(code):
    def replacer(m):
        return m.group(1) + '\n' + m.group(2) + m.group(3).strip() + '\n' + m.group(4)
    return DOUBLE_WRAP_RE.sub(replacer, code)

def make_func_void(code, func_name):
    """Strip `=` from a specific named function declaration only."""
    return re.sub(
        r'(func\s+' + re.escape(func_name) + r'\s*\([^)]*\))\s*=\s*\{',
        r'\1 {',
        code
    )

def remove_trailing_bare_var(code, func_name):
    """Remove the last bare-variable line from the body of a specific void function.

    Works line-by-line to avoid leaking across function boundaries.
    Finds `func func_name(...) {`, tracks brace depth to locate the closing `}`,
    then removes the last line before it if it is solely an identifier.
    """
    lines = code.splitlines(keepends=True)
    header_re = re.compile(r'^\s*func\s+' + re.escape(func_name) + r'\s*\(')
    bare_var_re = re.compile(r'^\s+(\w+)\s*$')

    i = 0
    while i < len(lines):
        if header_re.match(lines[i]):
            # Find the opening '{' (may be on same line or next)
            start = i
            brace_start = None
            for j in range(i, min(i + 3, len(lines))):
                if '{' in lines[j]:
                    brace_start = j
                    break
            if brace_start is None:
                i += 1
                continue

            # Walk forward counting braces to find the matching '}'
            depth = 0
            end = None
            for j in range(brace_start, len(lines)):
                depth += lines[j].count('{') - lines[j].count('}')
                if depth == 0:
                    end = j
                    break
            if end is None:
                i += 1
                continue

            # Find last non-empty line before the closing brace
            last_body = end - 1
            while last_body > brace_start and lines[last_body].strip() == '':
                last_body -= 1

            if bare_var_re.match(lines[last_body]):
                del lines[last_body]
                return ''.join(lines)

            i = end + 1
        else:
            i += 1
    return code

def extract_func_name_from_error(err):
    """Extract function name if the error is about a user-defined function call."""
    m = re.search(r"result of `(\w+)\(", err)
    return m.group(1) if m else None

def func_name_at_line(code, lineno):
    """Return the name of the func declaration whose body contains the given line number."""
    code_lines = code.splitlines()
    func_re = re.compile(r'^\s*func\s+(\w+)\s*\(')
    last_func = None
    for idx, line in enumerate(code_lines, 1):
        m = func_re.match(line)
        if m:
            last_func = m.group(1)
        if idx == lineno:
            return last_func
    return None

def extract_func_name_from_consumable_error(err, code):
    """For 'Expected a consumable' errors, find which func contains the offending line."""
    m = re.search(r'\[line (\d+)\]', err)
    if not m:
        return None
    lineno = int(m.group(1))
    return func_name_at_line(code, lineno)

def fix_double_wrap(code):
    def replacer(m):
        return m.group(1) + '\n' + m.group(2) + m.group(3).strip() + '\n' + m.group(4)
    return DOUBLE_WRAP_RE.sub(replacer, code)

def try_fix_cannot_include(code):
    """For 'Cannot include a statement' with no line info, try stripping = from each
    returning func one at a time until one resolves the error."""
    returning_funcs = re.findall(r'func\s+(\w+)\s*\([^)]*\)\s*=\s*\{', code)
    for name in returning_funcs:
        candidate = make_func_void(code, name)
        if candidate == code:
            continue
        # may expose a trailing-var error too, try trimming as well
        for c in [candidate, remove_trailing_bare_var(candidate, name)]:
            rc, _ = run_falcon(c)
            if rc == 0:
                return "void:" + name, c
    return None, None

def try_single_fix(code, err):
    """Apply one round of void-strip or double-wrap fix.
    Returns (label, new_code, passed) where passed=True if the fix fully resolves the error.
    Returns (None, None, False) if no candidate could be generated.
    May return a candidate that still fails — caller handles iteration.
    """
    if "Cannot include a statement" in err:
        label, fixed = try_fix_cannot_include(code)
        return label, fixed, fixed is not None
    elif "Expected a consumable" in err:
        func_name = extract_func_name_from_consumable_error(err, code)
    else:
        func_name = extract_func_name_from_error(err)

    candidates = []

    if func_name:
        a = make_func_void(code, func_name)
        if a != code:
            ab = remove_trailing_bare_var(a, func_name)
            if ab != a:
                # Prefer void+trim first — avoids exposing bare-var errors in next iteration
                candidates.append(("void+trim:" + func_name, ab))
            candidates.append(("void:" + func_name, a))

    c = fix_double_wrap(code)
    if c != code:
        candidates.append(("double_wrap", c))
        if func_name:
            ac = make_func_void(c, func_name)
            if ac != c:
                candidates.append(("double_wrap+void:" + func_name, ac))

    # Try candidates in order; prefer those that fully pass
    best_label, best_candidate = None, None
    for label, candidate in candidates:
        rc, new_err = run_falcon(candidate)
        if rc == 0:
            return label, candidate, True
        # Keep as best partial fix (different error = progress)
        if best_candidate is None and candidate != code:
            best_label, best_candidate = label, candidate

    if best_candidate is not None:
        return best_label, best_candidate, False

    return None, None, False

def try_fixes(code, err):
    """Iteratively apply single fixes until the code passes or no progress is made."""
    current, current_err = code, err
    labels = []
    seen = set()
    for _ in range(10):  # at most 10 rounds
        label, fixed, passed = try_single_fix(current, current_err)
        if fixed is None or fixed == current or fixed in seen:
            break
        seen.add(fixed)
        labels.append(label)
        if passed:
            return "+".join(labels), fixed
        rc, new_err = run_falcon(fixed)
        if rc == 0:
            return "+".join(labels), fixed
        current, current_err = fixed, new_err
    return None, None

def extract_code(content):
    m = re.search(r'```falcon\n(.*?)```', content, re.DOTALL)
    return m.group(1).strip() if m else content.strip()

def replace_code_in_content(content, new_code):
    return re.sub(
        r'(```falcon\n).*?(```)',
        lambda m: m.group(1) + new_code + '\n' + m.group(2),
        content, flags=re.DOTALL
    )

def run_falcon(code):
    with tempfile.NamedTemporaryFile(mode='w', suffix='.mist', delete=False) as f:
        f.write(code); path = f.name
    r = subprocess.run([FALCON, "run", path], capture_output=True, text=True, timeout=10)
    Path(path).unlink(missing_ok=True)
    return r.returncode, r.stderr.strip()

def main():
    if not MASTER.exists():
        print(f"ERROR: {MASTER} not found", file=sys.stderr); sys.exit(1)

    shutil.copy2(MASTER, BACKUP)
    print(f"Backup: {BACKUP}")

    raw_lines = MASTER.read_text(encoding="utf-8").splitlines()
    print(f"Total lines: {len(raw_lines)}\n")

    patched_lines = list(raw_lines)
    patched = skipped_no_fix = 0

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

        rc, err = run_falcon(code)
        if rc == 0 or ("never used" not in err and "Expected a consumable" not in err and "Cannot include a statement" not in err):
            continue

        label, fixed = try_fixes(code, err)
        current = fixed if fixed is not None else code
        applied = [label] if label else []

        rc_final, _ = run_falcon(current)
        if rc_final != 0 or current == code:
            print(f"  SKIP idx={i}: no fix worked — {err.splitlines()[-1][:70]}")
            skipped_no_fix += 1
            continue

        messages[1]["content"] = replace_code_in_content(assistant_content, current)
        new_line = json.dumps(entry, ensure_ascii=False)
        try:
            json.loads(new_line)
        except json.JSONDecodeError as e:
            print(f"  SKIP idx={i}: JSON round-trip failed: {e}"); continue

        patched_lines[i] = new_line
        patched += 1

        orig_l = code.splitlines()
        fix_l = current.splitlines()
        diffs = [(o, n) for o, n in zip(orig_l, fix_l) if o != n]
        print(f"  PATCHED idx={i} ({' → '.join(applied)}): {len(diffs)} line(s) changed")
        for o, n in diffs[:2]:
            print(f"    - {o.strip()}")
            print(f"    + {n.strip()}")

    MASTER.write_text("\n".join(patched_lines), encoding="utf-8")
    print(f"\nDone. Patched: {patched}, no-fix: {skipped_no_fix}")

if __name__ == "__main__":
    main()
