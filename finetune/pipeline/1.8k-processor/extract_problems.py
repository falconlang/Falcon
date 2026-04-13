import json
from pathlib import Path

INPUT_FILE = Path("/home/kumaraswamy/Documents/falcon/finetune/dataset/1.8k-codes-steps-output/MASTER.jsonl")
OUTPUT_FILE = Path("/home/kumaraswamy/Documents/falcon/finetune/dataset/1.8k-codes-steps-output/problems.txt")

with open(INPUT_FILE, 'r', encoding='utf-8') as fin, \
     open(OUTPUT_FILE, 'w', encoding='utf-8') as fout:
    for line in fin:
        line = line.strip()
        if not line:
            continue
        entry = json.loads(line)
        problem = entry["messages"][0]["content"]
        fout.write(problem + "\n")

print(f"Wrote problems to {OUTPUT_FILE}")
