#!/usr/bin/env python3
"""Merge all chunk files (chunk_01.txt to chunk_10.txt) into a single JSON array."""

import json
import re
from pathlib import Path

INPUT_DIR = Path("/home/kumaraswamy/Documents/falcon/finetune/dataset/1.8k-codes-steps-output/")
OUTPUT_FILE = INPUT_DIR / "merged_chunks.json"


def parse_chunk_file(filepath: Path) -> list[str]:
    """Parse a chunk file and return a list of entry contents."""
    content = filepath.read_text()

    # Split by entry markers
    # Pattern matches === ENTRY <number> ===
    pattern = r'=== ENTRY \d+ ==='

    # Find all entry markers and their positions
    parts = re.split(pattern, content)

    # parts[0] is content before first entry marker (usually empty or whitespace)
    # Each subsequent part is the content of an entry
    entries = []
    for part in parts[1:]:  # Skip the first empty part
        # Strip trailing/leading whitespace
        entry = part.strip()
        if entry:  # Only add non-empty entries
            entries.append(entry)

    return entries


def main():
    all_entries = []

    for i in range(1, 11):  # chunk_01.txt to chunk_10.txt
        chunk_file = INPUT_DIR / f"chunk_{i:02d}.txt"

        if not chunk_file.exists():
            print(f"Warning: {chunk_file} not found, skipping...")
            continue

        entries = parse_chunk_file(chunk_file)
        print(f"Processed {chunk_file.name}: {len(entries)} entries")
        all_entries.extend(entries)

    # Write to JSON file
    with open(OUTPUT_FILE, 'w', encoding='utf-8') as f:
        json.dump(all_entries, f, indent=2, ensure_ascii=False)

    print(f"\nTotal entries: {len(all_entries)}")
    print(f"Output saved to: {OUTPUT_FILE}")


if __name__ == "__main__":
    main()