#!/usr/bin/env python3
"""
Reformat all code entries in MASTER.json by passing them through
the Falcon AST formatter (Falcon format), producing a new MASTER.json
with clean, consistently formatted code.
"""

import json
import subprocess
import sys
import os

FALCON = os.path.join(os.path.dirname(__file__), '..', 'lang', 'Falcon')
INPUT = os.path.join(os.path.dirname(__file__), '..', 'answers', 'MASTER.json')
OUTPUT = os.path.join(os.path.dirname(__file__), '..', 'answers', 'MASTER.json')

def format_code(code: str) -> str | None:
    try:
        result = subprocess.run(
            [FALCON, 'format'],
            input=code,
            capture_output=True,
            text=True,
            timeout=10
        )
        if result.returncode != 0:
            return None
        return result.stdout
    except Exception:
        return None

def main():
    with open(INPUT, 'r') as f:
        data = json.load(f)

    total = len(data)
    failed = []
    formatted = {}

    for i, (key, entry) in enumerate(data.items(), 1):
        code = entry['answer']
        result = format_code(code)
        if result is None:
            failed.append(key)
            formatted[key] = entry  # keep original on failure
            print(f'[{i}/{total}] FAIL  {key}', file=sys.stderr)
        else:
            formatted[key] = {
                'problem': entry['problem'],
                'answer': result
            }
            if i % 100 == 0:
                print(f'[{i}/{total}] ok', file=sys.stderr)

    with open(OUTPUT, 'w') as f:
        json.dump(formatted, f, indent=2, ensure_ascii=False)

    print(f'\nDone. {total - len(failed)}/{total} formatted. {len(failed)} failed.')
    if failed:
        print('Failed keys:', failed[:20])

if __name__ == '__main__':
    main()
