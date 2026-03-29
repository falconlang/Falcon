#!/usr/bin/env python3
"""
Check how many programs in MASTER.jsonl have double-nested braces in func bodies.
Patterns caught:
  func foo(...) = { { ... } }
  func foo(...) { { ... } }
"""
import json
import re

MASTER_PATH = '/home/kumaraswamy/Documents/falcon/finetune/dataset/1.8k-codes-steps-output/MASTER.jsonl'
CODE_RE = re.compile(r'```falcon\n(.*?)```', re.DOTALL)

# func name(...) = { {   OR   func name(...) { {
DOUBLE_BRACE_RE = re.compile(r'func\s+\w+\s*\([^)]*\)\s*=?\s*\{\s*\{', re.DOTALL)


def main():
    hits = []  # (line_num, matched_snippet)

    with open(MASTER_PATH, 'r', encoding='utf-8') as f:
        lines = f.read().splitlines()

    for i, line in enumerate(lines):
        line_num = i + 1
        try:
            obj = json.loads(line)
        except json.JSONDecodeError:
            continue

        for msg in obj.get('messages', []):
            if msg.get('role') != 'assistant':
                continue
            m = CODE_RE.search(msg['content'])
            if not m:
                continue
            code = m.group(1)
            dm = DOUBLE_BRACE_RE.search(code)
            if dm:
                # grab a bit of context around the match
                start = dm.start()
                snippet = code[start:start+80].replace('\n', ' ')
                hits.append((line_num, snippet))

    print(f"Total entries : {len(lines)}")
    print(f"Double-brace  : {len(hits)}")
    if hits:
        print("\nLines with double-nested braces:")
        for line_num, snippet in hits:
            print(f"  Line {line_num:4d}: {snippet!r}")


if __name__ == '__main__':
    main()