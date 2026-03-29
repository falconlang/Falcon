#!/usr/bin/env python3
"""
Fix inverted comparison operators in .max and .min lambda expressions in MASTER.jsonl.

.max { a, b -> ... } correct form: second param > first param  (bugs have first > second)
.min { a, b -> ... } correct form: second param < first param  (bugs have first < second)

Fix: swap the LHS and RHS of the comparison in the lambda body.
Handles multi-statement lambdas (line 167) by only swapping the final comparison line.
"""
import json
import re
import sys

MAX_BUG_LINES = {96, 102, 121, 122, 156, 922, 990, 1070, 1617, 1642, 1725, 1735, 1736, 1742, 1755}
MIN_BUG_LINES = {167, 422, 1070, 1344, 1735}


def find_matching_brace(s, start):
    """Return index of the closing } that matches the opening { at s[start]."""
    depth = 0
    for i in range(start, len(s)):
        if s[i] == '{':
            depth += 1
        elif s[i] == '}':
            depth -= 1
            if depth == 0:
                return i
    raise ValueError(f"No matching brace found starting at index {start}")


def fix_lambda(lambda_text, op):
    """
    Given a full lambda string like '{ a, b -> LHS op RHS }', return
    a version with LHS and RHS swapped around `op`.

    For multi-statement lambdas (local vars + final expression), only the
    final comparison line is affected; preceding statements are preserved.
    """
    op_str = f' {op} '

    arrow_idx = lambda_text.find('->')
    if arrow_idx == -1:
        return lambda_text

    body = lambda_text[arrow_idx + 2:]   # everything after '->'

    op_idx = body.rfind(op_str)           # last occurrence = the comparison
    if op_idx == -1:
        return lambda_text

    # ---- locate LHS ----
    # LHS is the expression on the current "line" before the operator.
    # We look backward from op_idx for the last \n in the body;
    # that newline separates preceding statements from the comparison line.
    before_op = body[:op_idx]
    last_nl = before_op.rfind('\n')
    lhs_line_start = last_nl + 1 if last_nl >= 0 else 0

    lhs_full = body[lhs_line_start:op_idx]   # may have leading whitespace
    lhs_stripped = lhs_full.lstrip()
    lhs_indent = lhs_full[:len(lhs_full) - len(lhs_stripped)]

    # ---- locate RHS ----
    after_op = body[op_idx + len(op_str):]
    # The lambda ends with ' }'; find the last one (closes the lambda)
    close_pos = after_op.rindex(' }')
    rhs = after_op[:close_pos].strip()
    closing = after_op[close_pos:]           # ' }'

    # ---- rebuild ----
    # Keep everything up to the comparison line, then swap
    new_body = (
        body[:lhs_line_start]
        + lhs_indent + rhs
        + op_str
        + lhs_stripped
        + closing
    )

    return lambda_text[:arrow_idx + 2] + new_body


def fix_lambdas_in_content(content, method, op):
    """Find every .method { p, q -> ... } lambda in content and fix its comparison."""
    pattern = rf'\.{method} \{{ \w+, \w+ -> '

    # Collect all (brace_start, brace_end) positions first, then replace in reverse
    replacements = []
    for m in re.finditer(pattern, content):
        brace_start = content.index('{', m.start())
        brace_end = find_matching_brace(content, brace_start)
        lambda_text = content[brace_start:brace_end + 1]
        fixed = fix_lambda(lambda_text, op)
        if fixed != lambda_text:
            replacements.append((brace_start, brace_end + 1, fixed))

    for start, end, replacement in reversed(replacements):
        content = content[:start] + replacement + content[end:]

    return content


def process_line(raw, line_num):
    """Parse one JSONL line, apply fixes, return (possibly modified) JSON string."""
    try:
        obj = json.loads(raw)
    except json.JSONDecodeError as e:
        print(f"  JSON parse error on line {line_num}: {e}", file=sys.stderr)
        return raw

    do_max = line_num in MAX_BUG_LINES
    do_min = line_num in MIN_BUG_LINES
    if not do_max and not do_min:
        return raw

    changed = False
    for msg in obj.get('messages', []):
        if msg.get('role') != 'assistant':
            continue
        content = msg['content']
        original = content
        if do_max:
            content = fix_lambdas_in_content(content, 'max', '>')
        if do_min:
            content = fix_lambdas_in_content(content, 'min', '<')
        if content != original:
            msg['content'] = content
            changed = True

    if changed:
        return json.dumps(obj, ensure_ascii=False)
    return raw


def extract_lambdas(content, method, op):
    """Return list of (original, fixed) lambda strings found in content."""
    pattern = rf'\.{method} \{{ \w+, \w+ -> '
    pairs = []
    for m in re.finditer(pattern, content):
        brace_start = content.index('{', m.start())
        brace_end = find_matching_brace(content, brace_start)
        original = content[brace_start:brace_end + 1]
        fixed = fix_lambda(original, op)
        if fixed != original:
            pairs.append((original, fixed))
    return pairs


def main():
    import argparse
    parser = argparse.ArgumentParser(description='Fix inverted .max/.min lambdas in MASTER.jsonl')
    parser.add_argument('--preview', action='store_true',
                        help='Show before/after without writing the file')
    args = parser.parse_args()

    path = 'finetune/dataset/1.8k-codes-steps-output/MASTER.jsonl'

    with open(path, 'r', encoding='utf-8') as f:
        lines = f.read().splitlines()

    fixed_lines = []
    fix_count = 0

    for i, line in enumerate(lines):
        line_num = i + 1
        fixed = process_line(line, line_num)
        if fixed != line:
            fix_count += 1
            if args.preview:
                # Show the changed lambdas
                obj_orig = json.loads(line)
                obj_fixed = json.loads(fixed)
                for mo, mf in zip(obj_orig['messages'], obj_fixed['messages']):
                    if mo.get('role') == 'assistant' and mo['content'] != mf['content']:
                        print(f"\n--- Line {line_num} BEFORE ---")
                        print(mo['content'])
                        print(f"--- Line {line_num} AFTER  ---")
                        print(mf['content'])
            else:
                print(f"  Fixed line {line_num}")
        fixed_lines.append(fixed)

    if args.preview:
        print(f"\n[Preview] {fix_count} line(s) would be modified. Run without --preview to apply.")
    else:
        with open(path, 'w', encoding='utf-8') as f:
            f.write('\n'.join(fixed_lines) + '\n')
        print(f"\nDone. {fix_count} line(s) modified.")


if __name__ == '__main__':
    main()