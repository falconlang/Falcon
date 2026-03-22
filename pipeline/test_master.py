#!/usr/bin/env python3
"""Test all solutions in answers/MASTER.json against the Falcon compiler.

Usage:
    python3 pipeline/test_master.py [--binary lang/Falcon] [--timeout 10]
"""

import argparse
import json
import os
import subprocess
import sys
import tempfile
from pathlib import Path

SKIP_KEYWORDS = [
    "@Button", "@Label", "@TextBox", "@Notifier", "@Web", "@Camera",
    "when ", "openScreen", "closeScreen", "getStartValue",
    "openScreenWithValue", "closeScreenWithValue", "closeApp",
    "getPlainStartText", "any Button", "any Label",
]


def should_skip(code: str) -> bool:
    return any(kw.lower() in code.lower() for kw in SKIP_KEYWORDS)


def run_code(binary: str, code: str, timeout: int) -> tuple[bool, str]:
    with tempfile.NamedTemporaryFile(suffix=".mist", mode="w", delete=False, encoding="utf-8") as f:
        f.write(code)
        tmp = f.name
    try:
        result = subprocess.run(
            [binary, "run", tmp],
            capture_output=True, text=True, timeout=timeout,
        )
        return result.returncode == 0, result.stderr.strip()
    except subprocess.TimeoutExpired:
        return False, "timeout"
    finally:
        os.unlink(tmp)


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--binary", default="lang/Falcon")
    parser.add_argument("--master", default="answers/MASTER.json")
    parser.add_argument("--timeout", type=int, default=10)
    args = parser.parse_args()

    binary = os.path.abspath(args.binary)
    if not os.path.isfile(binary):
        sys.exit(f"Falcon binary not found: {binary}")

    with open(args.master, encoding="utf-8") as f:
        data = json.load(f)

    total = len(data)
    passed = 0
    failed = 0
    skipped = 0
    failures = []  # (id, error)

    for i, (pid, entry) in enumerate(data.items(), 1):
        code = entry.get("answer", "")
        if not code.strip():
            skipped += 1
            continue
        if should_skip(code):
            skipped += 1
            continue

        ok, err = run_code(binary, code, args.timeout)
        status = "PASS" if ok else "FAIL"
        print(f"[{i}/{total}] #{pid}: {status}" + (f"  — {err[:80]}" if not ok else ""))

        if ok:
            passed += 1
        else:
            failed += 1
            failures.append((pid, err))

    testable = passed + failed
    print(f"\n{'='*50}")
    print(f"Total entries : {total}")
    print(f"Skipped (UI)  : {skipped}")
    print(f"Tested        : {testable}")
    print(f"Passed        : {passed}  ({passed/testable*100:.1f}% of tested)" if testable else "Passed: 0")
    print(f"Failed        : {failed}  ({failed/testable*100:.1f}% of tested)" if testable else "Failed: 0")

    if failures:
        print(f"\nFailed IDs: {[pid for pid, _ in failures]}")


if __name__ == "__main__":
    main()