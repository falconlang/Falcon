#!/usr/bin/env python3
"""
Produce result.jsonl from MASTER.jsonl and merged_chunks.json.

For each entry, combines:
- Falcon code from MASTER.jsonl
- Steps from merged_chunks.json
- Output from executing the falcon code (if any)
"""

import json
import subprocess
import tempfile
import re
from pathlib import Path

INPUT_DIR = Path("/home/kumaraswamy/Documents/falcon/finetune/dataset/1.8k-codes-steps-output/")
MASTER_FILE = INPUT_DIR / "MASTER.jsonl"
CHUNKS_FILE = INPUT_DIR / "merged_chunks.json"
OUTPUT_FILE = INPUT_DIR / "result.jsonl"
FALCON_BIN = "/home/kumaraswamy/Documents/falcon/lang/Falcon"


def extract_falcon_code(content: str) -> str:
    """Extract falcon code from ```falcon ... ``` block."""
    match = re.search(r'```falcon\n(.*?)```', content, re.DOTALL)
    if match:
        return match.group(1).strip()
    return content.strip()


def execute_falcon(code: str) -> str | None:
    """Execute falcon code and return output (if any)."""
    with tempfile.NamedTemporaryFile(mode='w', suffix='.fal', delete=False) as f:
        f.write(code)
        temp_path = f.name

    try:
        result = subprocess.run(
            [FALCON_BIN, "run", temp_path],
            capture_output=True,
            text=True,
            timeout=30
        )
        output = result.stdout.strip()
        return output if output else None
    except subprocess.TimeoutExpired:
        return None
    except Exception as e:
        print(f"Error executing: {e}")
        return None
    finally:
        Path(temp_path).unlink(missing_ok=True)


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

        # Execute falcon code to get output
        output = execute_falcon(falcon_code)

        # Build new assistant content
        new_assistant = falcon_code

        if step_content:
            new_assistant += "\n\n" + step_content

        if output:
            new_assistant += "\n\nOutput\n```\n" + output + "\n```"

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