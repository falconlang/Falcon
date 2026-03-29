#!/usr/bin/env python3
"""
Half-trip normalize: for every entry in MASTER.jsonl run the falcon code through
  lex -> parser -> AST -> String   (Falcon format)
and write the canonical output back into the assistant content.

Entries that fail format are left unchanged and reported.
"""
import json
import re
import subprocess
import sys

FALCON_BIN = '/home/kumaraswamy/Documents/falcon/lang/Falcon'
MASTER_PATH = '/home/kumaraswamy/Documents/falcon/finetune/dataset/1.8k-codes-steps-output/MASTER.jsonl'

CODE_RE = re.compile(r'```falcon\n(.*?)```', re.DOTALL)


def run_format(code: str) -> tuple[bool, str]:
    result = subprocess.run(
        [FALCON_BIN, 'format'],
        input=code,
        capture_output=True,
        text=True,
        timeout=5,
    )
    return result.returncode == 0, result.stdout


def main():
    with open(MASTER_PATH, 'r', encoding='utf-8') as f:
        lines = f.read().splitlines()

    out_lines = []
    changed = 0
    failed = []

    for i, line in enumerate(lines):
        line_num = i + 1
        try:
            obj = json.loads(line)
        except json.JSONDecodeError:
            out_lines.append(line)
            continue

        modified = False
        for msg in obj.get('messages', []):
            if msg.get('role') != 'assistant':
                continue

            content = msg['content']
            m = CODE_RE.search(content)
            if not m:
                continue

            original_code = m.group(1)
            ok, formatted = run_format(original_code)

            if not ok:
                failed.append(line_num)
                continue

            formatted_code = formatted.rstrip('\n') + '\n'
            new_block = f'```falcon\n{formatted_code}```'
            new_content = content[:m.start()] + new_block + content[m.end():]

            if new_content != content:
                msg['content'] = new_content
                modified = True

        if modified:
            changed += 1
            out_lines.append(json.dumps(obj, ensure_ascii=False))
        else:
            out_lines.append(line)

        done = i + 1
        if done % 200 == 0:
            print(f"  {done}/{len(lines)} processed...", file=sys.stderr)

    with open(MASTER_PATH, 'w', encoding='utf-8') as f:
        f.write('\n'.join(out_lines) + '\n')

    print(f"\nDone.")
    print(f"  Lines rewritten : {changed}")
    print(f"  Format failures : {len(failed)}")
    if failed:
        print(f"  Failed lines    : {failed}")


if __name__ == '__main__':
    main()