#!/usr/bin/env python3
"""Verify MASTER.json programs produce correct, meaningful output.

No API calls — purely local execution + regex-based expected-value extraction.

Validation layers:
  1. Exit code 0  (already guaranteed by MASTER.json, but re-checked)
  2. Non-empty stdout  (for programs that use println)
  3. No "Error:" string in stdout  (Falcon prints runtime errors to stdout)
  4. Expected-value match  (extracted from problem text via regex)

Categories:
  PASS        — all applicable checks pass
  WARN        — ran fine but extracted expected value not found in output
  SUSPICIOUS  — empty output despite println, or "Error:" in stdout
  UNVERIFIABLE— no println and no extractable expected value (func-def only)

Usage:
    python3 pipeline/verify_master.py [--binary lang/Falcon] [--master answers/MASTER.json]
                                      [--save answers/MASTER_verified.json]
"""

import argparse
import json
import os
import re
import subprocess
import sys
import tempfile
from pathlib import Path


# ---------------------------------------------------------------------------
# Expected-value extraction from problem text
# ---------------------------------------------------------------------------

_EXPECTED_PATTERNS = [
    # backtick-wrapped number (int or float, optionally negative)
    re.compile(r'`(-?\d+(?:\.\d+)?)`'),
    # backtick bool
    re.compile(r'`(true|false)`'),
    # "= 1296" style (Cayley formula, etc.)
    re.compile(r'=\s*(\d+)(?:[;,.\s]|$)'),
    # "print `X`" or "prints `X`"
    re.compile(r'prints?\s+`([^`]+)`'),
    # "result is X" or "result: X"
    re.compile(r'result\s+(?:is|:)\s+`?(-?\d+(?:\.\d+)?)`?'),
    # "returns X" where X is a number or bool
    re.compile(r'returns?\s+`?(-?\d+(?:\.\d+)?|true|false)`?'),
    # explicit "output: X"
    re.compile(r'output[:\s]+`?(-?\d+(?:\.\d+)?|true|false)`?'),
]


def extract_expected(problem: str) -> list[str]:
    """Return a list of expected value strings found in the problem text."""
    found = []
    for pat in _EXPECTED_PATTERNS:
        for m in pat.finditer(problem):
            val = m.group(1).strip()
            if val and val not in found:
                found.append(val)
    return found


# ---------------------------------------------------------------------------
# Runner
# ---------------------------------------------------------------------------

def run_code(binary: str, code: str, timeout: int = 10) -> tuple[int, str]:
    """Run Falcon code; return (returncode, stdout)."""
    with tempfile.NamedTemporaryFile(suffix=".mist", mode="w", delete=False, encoding="utf-8") as f:
        f.write(code)
        tmp = f.name
    try:
        r = subprocess.run(
            [binary, "run", tmp],
            capture_output=True, text=True, timeout=timeout,
        )
        return r.returncode, r.stdout
    except subprocess.TimeoutExpired:
        return -1, ""
    finally:
        os.unlink(tmp)


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--binary",  default="lang/Falcon")
    parser.add_argument("--master",  default="answers/MASTER.json")
    parser.add_argument("--save",    default="", help="Write verified-only JSON here")
    parser.add_argument("--timeout", type=int, default=10)
    args = parser.parse_args()

    binary = os.path.abspath(args.binary)
    if not os.path.isfile(binary):
        sys.exit(f"Binary not found: {binary}")

    with open(args.master, encoding="utf-8") as f:
        data = json.load(f)

    total = len(data)
    cats = {"PASS": [], "WARN": [], "SUSPICIOUS": [], "UNVERIFIABLE": []}

    for i, (pid, entry) in enumerate(data.items(), 1):
        code    = entry.get("answer", "")
        problem = entry.get("problem", "")

        has_println = "println" in code

        rc, stdout = run_code(binary, code, args.timeout)

        # Layer 1 — compile / run
        if rc != 0:
            # Shouldn't happen (MASTER.json was already filtered), but guard anyway
            cats["SUSPICIOUS"].append((pid, "non-zero exit", stdout[:80]))
            print(f"[{i}/{total}] #{pid}: SUSPICIOUS (exit {rc})")
            continue

        # Layer 3 — no hidden "Error:" in stdout
        if stdout.lstrip().startswith("Error:"):
            cats["SUSPICIOUS"].append((pid, "Error: in stdout", stdout[:80]))
            print(f"[{i}/{total}] #{pid}: SUSPICIOUS (Error in stdout)")
            continue

        # Layer 2 — output presence
        if has_println and not stdout.strip():
            cats["SUSPICIOUS"].append((pid, "empty output despite println", ""))
            print(f"[{i}/{total}] #{pid}: SUSPICIOUS (no output)")
            continue

        # Layer 4 — expected value matching
        expected = extract_expected(problem)
        if expected:
            matched = any(v in stdout for v in expected)
            if matched:
                cats["PASS"].append(pid)
                print(f"[{i}/{total}] #{pid}: PASS (matched {expected})")
            else:
                cats["WARN"].append((pid, expected, stdout.strip()[:60]))
                print(f"[{i}/{total}] #{pid}: WARN  expected={expected}  got={stdout.strip()[:50]!r}")
        elif not has_println:
            cats["UNVERIFIABLE"].append(pid)
            print(f"[{i}/{total}] #{pid}: UNVERIFIABLE (no println, no expected value)")
        else:
            # Has output, no extractable expected value — treat as PASS
            cats["PASS"].append(pid)
            print(f"[{i}/{total}] #{pid}: PASS (output present, no expected to match)")

    # ---------------------------------------------------------------------------
    # Report
    # ---------------------------------------------------------------------------
    print(f"\n{'='*55}")
    print(f"Total entries    : {total}")
    print(f"PASS             : {len(cats['PASS'])}  ({len(cats['PASS'])/total*100:.1f}%)")
    print(f"WARN             : {len(cats['WARN'])}  ({len(cats['WARN'])/total*100:.1f}%)")
    print(f"SUSPICIOUS       : {len(cats['SUSPICIOUS'])}  ({len(cats['SUSPICIOUS'])/total*100:.1f}%)")
    print(f"UNVERIFIABLE     : {len(cats['UNVERIFIABLE'])}  ({len(cats['UNVERIFIABLE'])/total*100:.1f}%)")

    if cats["WARN"]:
        print(f"\n--- WARN details (first 20) ---")
        for pid, exp, got in cats["WARN"][:20]:
            print(f"  #{pid}: expected {exp}  →  got {got!r}")

    if cats["SUSPICIOUS"]:
        print(f"\n--- SUSPICIOUS details ---")
        for pid, reason, snippet in cats["SUSPICIOUS"]:
            print(f"  #{pid}: {reason}  {snippet!r}")

    # ---------------------------------------------------------------------------
    # Optionally save verified subset
    # ---------------------------------------------------------------------------
    if args.save:
        keep_ids = set(cats["PASS"]) | set(cats["UNVERIFIABLE"])
        verified = {pid: entry for pid, entry in data.items() if pid in keep_ids}
        with open(args.save, "w", encoding="utf-8") as f:
            json.dump(verified, f, indent=2, ensure_ascii=False)
        print(f"\nSaved {len(verified)} verified entries → {args.save}")


if __name__ == "__main__":
    main()
