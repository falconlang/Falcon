#!/usr/bin/env python3
"""
Produce result.jsonl from MASTER.jsonl and merged_chunks.json.

For each entry, combines:
- Falcon code from MASTER.jsonl
- Steps from merged_chunks.json
"""

import json
import re
from pathlib import Path

INPUT_DIR = Path("/home/kumaraswamy/Documents/falcon/finetune/dataset/1.8k-codes-steps-output/")
MASTER_FILE = INPUT_DIR / "MASTER.jsonl"
CHUNKS_FILE = INPUT_DIR / "merged_chunks.json"
OUTPUT_FILE = INPUT_DIR / "result.jsonl"


def extract_falcon_code(content: str) -> str:
    """Extract falcon code from ```falcon ... ``` block."""
    match = re.search(r'```falcon\n(.*?)```', content, re.DOTALL)
    if match:
        return match.group(1).strip()
    return content.strip()


def main():
    # Load merged chunks (steps)
    with open(CHUNKS_FILE, 'r', encoding='utf-8') as f:
        steps = json.load(f)

    # Count entries in MASTER.jsonl
    master_entries = []
    with open(MASTER_FILE, 'r', encoding='utf-8') as f:
        for line in f:
            line = line.strip()
            if line:
                master_entries.append(json.loads(line))

    print(f"Loaded {len(master_entries)} entries from MASTER.jsonl")
    print(f"Loaded {len(steps)} steps from merged_chunks.json")

    if len(master_entries) != len(steps):
        print(f"Warning: Entry count mismatch! Master: {len(master_entries)}, Steps: {len(steps)}")

    result_entries = []

    for i, entry in enumerate(master_entries):
        user_content = entry["messages"][0]["content"]
        original_assistant = entry["messages"][1]["content"]

        # Extract falcon code
        falcon_code = extract_falcon_code(original_assistant)

        # Get corresponding steps
        step_content = steps[i] if i < len(steps) else ""

        # Build new assistant content: steps first, then program
        new_assistant = ""

        if step_content:
            new_assistant += step_content + "\n\n"

        new_assistant += "```falcon\n" + falcon_code + "\n```"

        # Create result entry
        result_entry = {
            "messages": [
                {"role": "user", "content": user_content},
                {"role": "assistant", "content": new_assistant}
            ]
        }
        result_entries.append(result_entry)

        if (i + 1) % 100 == 0:
            print(f"Processed {i + 1}/{len(master_entries)} entries...")

    # Write result.jsonl
    with open(OUTPUT_FILE, 'w', encoding='utf-8') as f:
        for entry in result_entries:
            f.write(json.dumps(entry, ensure_ascii=False) + "\n")

    print(f"\nDone! Wrote {len(result_entries)} entries to {OUTPUT_FILE}")


if __name__ == "__main__":
    main()