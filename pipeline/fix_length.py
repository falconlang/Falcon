#!/usr/bin/env python3
"""
Fix the most common error in MASTER_REASON.jsonl:
  .length() → .textLen() or .listLen()

Strategy per entry:
  1. Try replacing ALL .length() with .textLen() → check parser
  2. If still fails, try ALL .length() with .listLen() → check parser
  3. If still fails, try heuristic per-call replacement based on variable name
  4. If still fails, leave unchanged (will be counted)
"""

import json
import subprocess
import re
import shutil
from pathlib import Path

FALCON_BIN = Path(__file__).parent.parent / "lang" / "Falcon"
MASTER_REASON = Path(__file__).parent.parent / "answer_reasoning" / "MASTER_REASON.jsonl"


# ---------- helpers ----------

def extract_code(content: str) -> str | None:
    content_no_think = re.sub(r"<think>.*?</think>", "", content, flags=re.DOTALL).strip()
    m = re.search(r"```falcon\s*(.*?)\s*```", content_no_think, re.DOTALL)
    return m.group(1).strip() if m else None


def replace_code_in_content(content: str, new_code: str) -> str:
    """Replace only the falcon code block (after </think>) with new_code."""
    # Split at </think>
    think_end = content.find("</think>")
    if think_end == -1:
        # No think block, just replace the code block
        return re.sub(r"```falcon\s*.*?\s*```", f"```falcon\n{new_code}\n```", content, flags=re.DOTALL)
    prefix = content[:think_end + len("</think>")]
    suffix = content[think_end + len("</think>"):]
    new_suffix = re.sub(r"```falcon\s*.*?\s*```", f"```falcon\n{new_code}\n```", suffix, flags=re.DOTALL)
    return prefix + new_suffix


def check(code: str) -> tuple[bool, str]:
    result = subprocess.run(
        [str(FALCON_BIN), "format"],
        input=code, capture_output=True, text=True, timeout=10,
    )
    return result.returncode == 0, result.stderr.strip()


def has_length_error(stderr: str) -> bool:
    return "Cannot find method .length()" in stderr


# ---------- fix strategies ----------

LIST_NAME_HINTS = re.compile(
    r"^(list|arr|items|nums|numbers|elems|elements|result|parts|tokens|"
    r"words|rows|cols|stack|queue|heap|entries|records|keys|values|pairs|"
    r"data|path|perm|choices|pool|bucket|counts|scores|bits|bytes|digits|"
    r"chars|ops|instrs|instructions|signals|samples|neighbors|children|"
    r"nodes|edges|seq|sequence|matrix|grid|row|col|board|line|lines|"
    r"sorted|filtered|mapped|chunk|chunks|subset|groups|table|pages|"
    r"memo|cache|hist|histogram|freq|freqs|weights|probs|deck|hand|"
    r"run|runs|cluster|clusters|candidates)\b",
    re.IGNORECASE
)


def is_list_var(varname: str) -> bool:
    return bool(LIST_NAME_HINTS.match(varname.strip()))


def smart_replace_length(code: str) -> str:
    """Replace each .length() based on the preceding variable name heuristic."""
    def repl(m):
        # Look backwards in the original string for the variable name
        prefix = code[:m.start()]
        # Find the last identifier or ] or ) before .length()
        var_match = re.search(r'(\w+)\s*$', prefix)
        if var_match:
            varname = var_match.group(1)
            if is_list_var(varname):
                return ".listLen()"
        return ".textLen()"

    return re.sub(r'\.length\(\)', repl, code)


def try_fix(code: str) -> str | None:
    """Try various fixes for .length(). Return fixed code or None."""
    # Strategy 1: all textLen
    c1 = code.replace(".length()", ".textLen()")
    ok, _ = check(c1)
    if ok:
        return c1

    # Strategy 2: all listLen
    c2 = code.replace(".length()", ".listLen()")
    ok, _ = check(c2)
    if ok:
        return c2

    # Strategy 3: heuristic per-call
    c3 = smart_replace_length(code)
    if c3 != code:
        ok, _ = check(c3)
        if ok:
            return c3

    # Strategy 4: heuristic + then try flipping remaining failures
    # Sometimes code has mixed string/list — try textLen for string-named, listLen for rest
    # Already covered by strategy 3, so nothing more here
    return None


# ---------- main ----------

def main():
    lines = MASTER_REASON.read_text(encoding="utf-8").splitlines()
    total = len(lines)

    fixed = 0
    skipped_no_length = 0
    still_failing = 0
    unchanged = 0

    out_lines = []
    for i, line in enumerate(lines):
        if (i + 1) % 200 == 0:
            print(f"  [{i+1}/{total}] fixed={fixed} still_failing={still_failing}", flush=True)

        record = json.loads(line)
        messages = record["messages"]
        assistant_msg = next((m for m in messages if m["role"] == "assistant"), None)
        if not assistant_msg:
            out_lines.append(line)
            unchanged += 1
            continue

        code = extract_code(assistant_msg["content"])
        if not code:
            out_lines.append(line)
            unchanged += 1
            continue

        ok, stderr = check(code)
        if ok:
            out_lines.append(line)
            unchanged += 1
            continue

        if not has_length_error(stderr):
            out_lines.append(line)
            skipped_no_length += 1
            continue

        # Try to fix
        fixed_code = try_fix(code)
        if fixed_code:
            new_content = replace_code_in_content(assistant_msg["content"], fixed_code)
            assistant_msg["content"] = new_content
            record["messages"] = messages
            out_lines.append(json.dumps(record, ensure_ascii=False))
            fixed += 1
        else:
            out_lines.append(line)
            still_failing += 1

    # Backup original
    backup = MASTER_REASON.with_suffix(".jsonl.bak_length")
    if not backup.exists():
        shutil.copy(MASTER_REASON, backup)
        print(f"Backup saved to {backup.name}")

    MASTER_REASON.write_text("\n".join(out_lines) + "\n", encoding="utf-8")

    print()
    print("=" * 50)
    print(f"Total entries     : {total}")
    print(f"Fixed (.length()) : {fixed}")
    print(f"Still failing     : {still_failing}  (had .length() but fix didn't help)")
    print(f"Not .length() err : {skipped_no_length}  (other errors, left unchanged)")
    print(f"Already passing   : {unchanged - skipped_no_length}")


if __name__ == "__main__":
    main()
