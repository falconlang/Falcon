#!/usr/bin/env python3
"""
Find MASTER.jsonl entries where a for-loop variable is reassigned inside the loop body.
e.g. for (x in list) { x = x * 2 }  <- x is the loop var, being reassigned
"""

import json
import re
from pathlib import Path

MASTER_FILE = Path("/home/kumaraswamy/Documents/falcon/finetune/dataset/1.8k-codes-steps-output/MASTER.jsonl")


def extract_falcon_code(content: str) -> str:
    match = re.search(r'```falcon\n(.*?)```', content, re.DOTALL)
    if match:
        return match.group(1)
    return content


def extract_block(code: str, start: int) -> str:
    """Extract the content of a { } block starting at `start` (index of opening brace)."""
    depth = 0
    i = start
    result = []
    while i < len(code):
        ch = code[i]
        if ch == '{':
            depth += 1
        elif ch == '}':
            depth -= 1
            if depth == 0:
                return ''.join(result)
        if depth > 0:
            result.append(ch)
        i += 1
    return ''.join(result)


def find_loop_var_mutations(code: str) -> list[tuple[str, str]]:
    """
    Returns list of (loop_var, loop_body) for each for-loop where the loop
    variable is reassigned inside the body.
    """
    # Match: for (VAR in ...) or for (VAR: range)
    # Loop var is the first identifier after '('
    for_pattern = re.compile(r'\bfor\s*\(\s*(\w+)\s*(?:in|:)', )
    matches = []

    for m in for_pattern.finditer(code):
        loop_var = m.group(1)

        # Find the opening brace of the loop body after this match
        brace_pos = code.find('{', m.end())
        if brace_pos == -1:
            continue

        body = extract_block(code, brace_pos)

        # Check if loop_var is assigned inside body
        # Match: loop_var = ... but NOT loop_var == or loop_var ===
        assign_pattern = re.compile(
            r'\b' + re.escape(loop_var) + r'\s*(?:\+|-|\*|/|%)?=(?!=)'
        )
        if assign_pattern.search(body):
            matches.append((loop_var, body.strip()))

    return matches


def main():
    results = []

    with open(MASTER_FILE, 'r', encoding='utf-8') as f:
        for line_num, line in enumerate(f, 1):
            line = line.strip()
            if not line:
                continue
            entry = json.loads(line)
            code = extract_falcon_code(entry["messages"][1]["content"])
            mutations = find_loop_var_mutations(code)
            if mutations:
                results.append({
                    "line": line_num,
                    "user": entry["messages"][0]["content"][:80],
                    "code": code.strip(),
                    "mutations": [{"var": v, "body": b} for v, b in mutations]
                })

    print(f"Found {len(results)} entries with loop variable mutations:\n")
    for r in results:
        print(f"--- Entry #{r['line']} ---")
        print(f"User: {r['user']}")
        for m in r['mutations']:
            print(f"  Loop var '{m['var']}' reassigned in body:")
            print(f"    {m['body'][:200]}")
        print()

    print(f"Total: {len(results)} / ", end="")
    with open(MASTER_FILE, 'r') as f:
        total = sum(1 for l in f if l.strip())
    print(f"{total} entries")


if __name__ == "__main__":
    main()
