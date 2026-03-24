#!/usr/bin/env python3
"""
Normalise Falcon code in MASTER_REASON.jsonl by round-tripping every code block
through `Falcon format` (lex → parse → AST.String()).

For each entry:
  1. Extract the ```falcon...``` block from the assistant message.
  2. Pipe it through `lang/Falcon format`.
  3. If the formatter succeeds, replace the code block with the normalised output.
  4. If it fails (already broken entry or binary error), skip and count.

The <think> block and everything else in the assistant message are left untouched.
The file is written atomically at the end.

Usage:
    python3 pipeline/normalise_syntax.py [--workers N]
"""

import argparse
import json
import os
import re
import subprocess
import sys
import threading
from concurrent.futures import ThreadPoolExecutor, as_completed
from pathlib import Path

FALCON_BIN   = Path(__file__).parent.parent / "lang" / "Falcon"
MASTER_REASON = Path(__file__).parent.parent / "answer_reasoning" / "MASTER_REASON.jsonl"


# ─── helpers ──────────────────────────────────────────────────────────────────

def extract_code(content: str) -> str | None:
    """Extract code from ```falcon...``` after </think>."""
    c = re.sub(r"<think>.*?</think>", "", content, flags=re.DOTALL).strip()
    m = re.search(r"```falcon\s*(.*?)\s*```", c, re.DOTALL)
    return m.group(1) if m else None


def replace_code(content: str, new_code: str) -> str:
    """Replace the ```falcon...``` block after </think>, preserving the think block.
    Uses a lambda replacement to avoid re.sub backslash processing."""
    think_end = content.find("</think>")
    if think_end == -1:
        suffix = content
        prefix = ""
    else:
        prefix = content[: think_end + len("</think>")]
        suffix = content[think_end + len("</think>") :]

    replacement = f"```falcon\n{new_code}\n```"
    new_suffix = re.sub(
        r"```falcon\s*.*?\s*```",
        lambda _: replacement,
        suffix,
        flags=re.DOTALL,
    )
    return prefix + new_suffix


def format_code(code: str) -> tuple[bool, str]:
    """Run `Falcon format` on code. Returns (ok, formatted_or_error)."""
    try:
        r = subprocess.run(
            [str(FALCON_BIN), "format"],
            input=code,
            capture_output=True,
            text=True,
            timeout=10,
        )
        if r.returncode == 0:
            return True, r.stdout
        return False, r.stderr.strip()
    except subprocess.TimeoutExpired:
        return False, "timeout"


# ─── worker ───────────────────────────────────────────────────────────────────

def process_line(idx: int, line: str) -> tuple[int, str | None, str]:
    """
    Returns (idx, new_line_or_None, status).
    new_line is None if unchanged or failed.
    """
    rec = json.loads(line)
    asst = next((m for m in rec["messages"] if m["role"] == "assistant"), None)
    if not asst:
        return idx, None, "no-asst"

    code = extract_code(asst["content"])
    if code is None:
        return idx, None, "no-code"

    ok, result = format_code(code)
    if not ok:
        return idx, None, f"fail: {result[:60]}"

    # Strip trailing newline that fmt adds but keep interior newlines intact
    normalised = result.rstrip("\n")

    if normalised == code.rstrip("\n"):
        return idx, None, "unchanged"

    asst["content"] = replace_code(asst["content"], normalised)
    return idx, json.dumps(rec, ensure_ascii=False), "ok"


# ─── main ─────────────────────────────────────────────────────────────────────

def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--workers", type=int, default=8)
    args = parser.parse_args()

    if not FALCON_BIN.exists():
        sys.exit(f"Falcon binary not found at {FALCON_BIN}")

    lines = MASTER_REASON.read_text(encoding="utf-8").splitlines(keepends=False)
    while lines and not lines[-1].strip():
        lines.pop()
    total = len(lines)
    print(f"Loaded {total} entries. Running with {args.workers} workers…")

    out_lines: list[str] = list(lines)
    counters = {"ok": 0, "unchanged": 0, "fail": 0, "skip": 0}
    lock = threading.Lock()

    with ThreadPoolExecutor(max_workers=args.workers) as ex:
        futures = {ex.submit(process_line, i, line): i for i, line in enumerate(lines)}
        done = 0
        for fut in as_completed(futures):
            idx, new_line, status = fut.result()
            done += 1
            with lock:
                if new_line is not None:
                    out_lines[idx] = new_line
                if status == "ok":
                    counters["ok"] += 1
                elif status == "unchanged":
                    counters["unchanged"] += 1
                elif status.startswith("fail"):
                    counters["fail"] += 1
                else:
                    counters["skip"] += 1
                if done % 200 == 0 or done == total:
                    print(f"  {done}/{total}  ok={counters['ok']}  unchanged={counters['unchanged']}  fail={counters['fail']}  skip={counters['skip']}", flush=True)

    # Atomic write via temp file
    tmp = MASTER_REASON.with_suffix(".jsonl.norm_tmp")
    tmp.write_text("\n".join(out_lines) + "\n", encoding="utf-8")
    os.replace(tmp, MASTER_REASON)

    print()
    print("=" * 55)
    print(f"Total entries  : {total}")
    print(f"Normalised     : {counters['ok']}")
    print(f"Already normal : {counters['unchanged']}")
    print(f"Failed (kept)  : {counters['fail']}")
    print(f"Skipped        : {counters['skip']}")
    print(f"Saved → {MASTER_REASON}")


if __name__ == "__main__":
    main()
