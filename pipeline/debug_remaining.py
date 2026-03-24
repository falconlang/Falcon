#!/usr/bin/env python3
"""Show failing entries with their actual error and code."""
import json, subprocess, re
from pathlib import Path

FALCON_BIN = Path(__file__).parent.parent / "lang" / "Falcon"
MASTER_REASON = Path(__file__).parent.parent / "answer_reasoning" / "MASTER_REASON.jsonl"

def extract_code(content):
    c = re.sub(r"<think>.*?</think>", "", content, flags=re.DOTALL).strip()
    m = re.search(r"```falcon\s*(.*?)\s*```", c, re.DOTALL)
    return m.group(1).strip() if m else None

def check(code):
    r = subprocess.run([str(FALCON_BIN), "format"], input=code, capture_output=True, text=True, timeout=10)
    return r.returncode == 0, r.stderr.strip()

lines = MASTER_REASON.read_text().splitlines()
failing = []
for i, line in enumerate(lines):
    rec = json.loads(line)
    asst = next((m for m in rec["messages"] if m["role"] == "assistant"), None)
    if not asst: continue
    code = extract_code(asst["content"])
    if not code: continue
    ok, err = check(code)
    if not ok:
        failing.append((i, err, code))

print(f"Total failing: {len(failing)}\n")
for i, err, code in failing[:30]:
    print(f"=== idx {i} ===")
    print(f"ERROR: {err[:120]}")
    print(f"CODE:\n{code[:400]}")
    print()
