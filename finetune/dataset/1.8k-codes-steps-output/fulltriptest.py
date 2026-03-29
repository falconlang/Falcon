#!/usr/bin/env python3
"""
Full roundtrip test:
  Falcon code -> lexer -> parser -> AST -> Blockly serializer -> XML -> Blockly parser -> AST

Uses `Falcon reformat` which performs the complete pipeline and exits 1 on any error.
"""
import json
import re
import subprocess
import sys

FALCON_BIN = '/home/kumaraswamy/Documents/falcon/lang/Falcon'
MASTER_PATH = '/home/kumaraswamy/Documents/falcon/finetune/dataset/1.8k-codes-steps-output/MASTER.jsonl'

CODE_RE = re.compile(r'```falcon\n(.*?)```', re.DOTALL)


def extract_code(content: str) -> str | None:
    m = CODE_RE.search(content)
    return m.group(1) if m else None


def reformat_check(code: str) -> tuple[bool, str]:
    result = subprocess.run(
        [FALCON_BIN, 'reformat'],
        input=code,
        capture_output=True,
        text=True,
        timeout=5,
    )
    return result.returncode == 0, result.stderr.strip()


def main():
    passed = []
    failed = []
    skipped = 0

    with open(MASTER_PATH, 'r', encoding='utf-8') as f:
        lines = f.read().splitlines()

    total = len(lines)
    for i, line in enumerate(lines):
        line_num = i + 1
        try:
            obj = json.loads(line)
        except json.JSONDecodeError:
            skipped += 1
            continue

        code = None
        for msg in obj.get('messages', []):
            if msg.get('role') == 'assistant':
                code = extract_code(msg['content'])
                break

        if code is None:
            skipped += 1
            continue

        ok, err = reformat_check(code)
        if ok:
            passed.append(line_num)
        else:
            failed.append((line_num, err))

        done = len(passed) + len(failed)
        if done % 100 == 0:
            print(f"  {done}/{total - skipped} checked...", file=sys.stderr)

    checked = len(passed) + len(failed)
    print(f"\n{'='*50}")
    print(f"Total entries : {total}")
    print(f"Skipped       : {skipped}  (no falcon block)")
    print(f"Checked       : {checked}")
    print(f"PASSED        : {len(passed)}  ({100*len(passed)/checked:.1f}%)")
    print(f"FAILED        : {len(failed)}  ({100*len(failed)/checked:.1f}%)")
    print(f"{'='*50}")

    if failed:
        print(f"\nFailing lines:")
        for line_num, err in failed:
            short_err = err.splitlines()[0] if err else '(no stderr)'
            print(f"  Line {line_num:4d}: {short_err}")


if __name__ == '__main__':
    main()