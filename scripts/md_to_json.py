#!/usr/bin/env python3
"""Transform problems/*.md files into probjson/*.json files.
Each JSON maps problem number (string) -> problem text."""

import os
import re
import json
import glob

os.makedirs("probjson", exist_ok=True)

problem_re = re.compile(r'^(\d+)\.\s+(.+)$', re.MULTILINE)

for md_file in sorted(glob.glob("problems/*.md")):
    basename = os.path.basename(md_file)
    name = os.path.splitext(basename)[0]

    with open(md_file, 'r', encoding='utf-8') as f:
        content = f.read()

    problems = {}
    for m in problem_re.finditer(content):
        num = m.group(1)
        text = m.group(2).strip()
        problems[num] = text

    out_path = f"probjson/{name}.json"
    with open(out_path, 'w', encoding='utf-8') as f:
        json.dump(problems, f, indent=2, ensure_ascii=False)

    print(f"{name}.json  —  {len(problems)} problems")
