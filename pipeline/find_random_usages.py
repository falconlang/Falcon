#!/usr/bin/env python3
"""
Print code blocks from MASTER_REASON.jsonl that contain 'random'.
Standalone random() is not a defined function — only list.random() works.
"""

import json
import re
import sys

JSONL_PATH = "../answer_reasoning/MASTER_REASON.jsonl"

def extract_code_blocks(content):
    return re.findall(r'```falcon(.*?)```', content, re.DOTALL)

def has_standalone_random(code):
    """Returns True if code uses random() not preceded by a dot (i.e. not list.random())."""
    return bool(re.search(r'(?<!\.)random\s*\(', code))

with open(JSONL_PATH) as f:
    entries = [json.loads(line) for line in f]

print(f"Scanning {len(entries)} entries...\n")

found = 0
for i, entry in enumerate(entries, 1):
    for msg in entry["messages"]:
        if msg["role"] != "assistant":
            continue
        codes = extract_code_blocks(msg["content"])
        for code in codes:
            if "random" not in code:
                continue
            standalone = has_standalone_random(code)
            tag = " [STANDALONE random() BUG]" if standalone else " [list.random() — OK]"
            print(f"=== Entry {i}{tag} ===")
            print(code.strip())
            print()
            found += 1

print(f"Total entries with 'random': {found}")
