#!/usr/bin/env python3
"""pipeline/generate_reasons.py — Generate reasoning traces for Falcon problems using extended thinking.

For each problem in MASTER.jsonl, calls claude-3-7-sonnet with extended thinking, extracts
the generated code, validates it by compiling and running it locally, uses Claude as a judge
to verify the output is correct, and retries on failure. Only validated entries are written.

Output format (each line of MASTER_REASON.jsonl):
    {"messages": [
        {"role": "user", "content": "<problem>"},
        {"role": "assistant", "content": "<think>\\n...\\n</think>\\n```falcon\\n...\\n```"}
    ]}

Usage:
    python3 pipeline/generate_reasons.py [options]

Options:
    --input   PATH    Source JSONL (default: answers/MASTER.jsonl)
    --output  PATH    Output JSONL (default: answers/MASTER_REASON.jsonl)
    --binary  PATH    Falcon binary (default: lang/Falcon)
    --workers N       Parallel threads (default: 5)
    --chunk   N       Entries per worker chunk (default: 100)
    --model   ID      Thinking model (default: claude-3-7-sonnet-20250219)
    --verify-model ID Model used for output verification (default: claude-haiku-4-5-20251001)
    --budget  N       Thinking token budget per attempt (default: 8000)
    --retries N       Max retries per entry on failure (default: 2)
    --limit   N       Max entries to process (default: unlimited)
"""

import argparse
import json
import os
import re
import subprocess
import sys
import tempfile
import threading
from concurrent.futures import ThreadPoolExecutor, as_completed

import anthropic

# ---------------------------------------------------------------------------
# Falcon language reference — copied from solve.py
# ---------------------------------------------------------------------------

FALCON_GUIDE = """
## Falcon Language Reference

### Quirks
1. 1-based indexing.
2. Variables are dynamically typed — no type declarations.
3. Lists and dicts are passed as references.
4. The last expression in any body is the return value (no `return` keyword).
5. No try-catch or throw.
6. Single-line comments only: `// comment`
7. Never use `_` for unused variables.
8. Variables can never be uninitialized.
9. String contents that are numeric can be numerically operated on: `"2" + "3.14"` is valid.

### Data types
- String: `"Hello"`
- Boolean: `true` / `false`
- Number: `123`, `3.14`
- List: `[1, 2, 3]`
- Dictionary: `{"key": "value"}`
- Colour: `#FFFFFF`

### Operators
- Arithmetic: `+`, `-`, `*`, `/`, `%` (remainder), `^` (power)
- Logical: `&&`, `||`
- Bitwise: `&`, `|`, `~` (xor)
- Equality: `==`, `!=`
- Relational: `<`, `<=`, `>`, `>=`
- Text lexicographic: `===` (equals), `!==` (not equals), `<<` (less), `>>` (greater)
- Unary: `!` (not), `-` (negate)
- Text join: `"Hello " _ "World!"`
- Pair: `"Fruit": "Mango"`
- Question `?`: type check — `"Hello" ? text`, `"1010" ? bin`, `[] ? emptyList`

### Variables
```
global score = 0        // global, access via this.score
local name = "Falcon"   // local, access via name
```

### If / else
```
if (x > y) {
  println("X is greater")
} else if (y > x) {
  println("Y is greater")
} else {
  println("Equal")
}
// As expression:
local msg = if (x > y) "X" else "Y"
```

### While loop
```
local x = 0
while (x < 5) {
  x = x + 1
}
```

### For n loop
```
for (i: 1 .. 10 step 2) { println(i) }   // step is optional, defaults to 1
```

### Each loop (list)
```
for (item in myList) { println(item) }
```

### Each loop (dict)
```
for (key, value in myDict) { println(key _ ": " _ value) }
```

### Functions
```
// Void function (no = before body):
func greet(name) {
  println("Hello " _ name)
}

// Result function (= before body, last expr is returned):
func double(n) = { n * 2 }

func fib(n) = {
  if (n < 2) { n } else { fib(n - 1) + fib(n - 2) }
}
```

### List access & mutation
```
local nums = [10, 20, 30]
println(nums[1])    // 10  (1-based)
nums[1] = 99
```

### Dict access
```
local d = {"a": 1, "b": 2}
println(d.get("a", 0))
d.set("c", 3)
```

### Built-in functions
Math: `sqrt`, `abs`, `neg`, `log`, `exp`, `round`, `ceil`, `floor`,
      `sin`, `cos`, `tan`, `asin`, `acos`, `atan`, `degrees`, `radians`,
      `decToHex`, `decToBin`, `hexToDec`, `binToDec`, `dec`, `bin`, `octal`, `hexa`,
      `randInt(from, to)`, `randFloat()`, `setRandSeed(n)`,
      `min(...)`, `max(...)`, `avgOf(list)`, `maxOf(list)`, `minOf(list)`,
      `geoMeanOf(list)`, `stdDevOf(list)`, `stdErrOf(list)`, `modeOf(list)`,
      `mod(x,y)`, `rem(x,y)`, `quot(x,y)`, `atan2(a,b)`, `formatDecimal(n, places)`
Output: `println(any)`
Values: `copyList(list)`, `copyDict(dict)`, `makeColor([r,g,b])`, `splitColor(colour)`

### Text methods
`textLen()`, `trim()`, `uppercase()`, `lowercase()`, `startsWith(s)`, `contains(s)`,
`containsAny(list)`, `containsAll(list)`, `split(at)`, `splitAtFirst(at)`,
`splitAtAny(list)`, `splitAtFirstOfAny(list)`, `splitAtSpaces()`, `reverse()`,
`csvRowToList()`, `csvTableToList()`, `segment(from, length)`,
`replace(target, replacement)`, `replaceFrom(dict)`, `replaceFromLongestFirst(dict)`

### List methods
`listLen()`, `add(any...)`, `containsItem(any)`, `indexOf(any)`, `insert(at, any)`,
`remove(at)`, `appendList(list)`, `lookupInPairs(key, notfound)`, `join(sep)`,
`slice(i1, i2)`, `random()`, `reverseList()`, `toCsvRow()`, `toCsvTable()`,
`sort()`, `allButFirst()`, `allButLast()`, `pairsToDict()`

### Dict methods
`dictLen()`, `get(key, notfound)`, `set(key, value)`, `delete(key)`,
`getAtPath(pathList, notfound)`, `setAtPath(pathList, value)`,
`containsKey(key)`, `mergeInto(dict)`, `walkTree(path)`, `keys()`, `values()`, `toPairs()`

### List lambdas
```
local doubled = nums.map { n -> n * 2 }
local evens   = nums.filter { n -> n % 2 == 0 }
local sorted  = names.sort { m, n -> m.textLen() > n.textLen() }
local longest = names.max { m, n -> n.textLen() > m.textLen() }
local total   = nums.reduce(0) { x, acc -> x + acc }
```
"""

SYSTEM_PROMPT = f"""You are an expert Falcon programmer. Solve the given problem in Falcon.

{FALCON_GUIDE}

Rules:
- Use `println()` for all output.
- Do NOT use App Inventor components (`@Button`, `when ...`, `openScreen`, etc.) — this is a CLI runtime.
- Follow all quirks listed above, especially 1-based indexing and no `return` keyword.
- Wrap your solution in a ```falcon code block. No other text outside the code block.
"""

VERIFY_SYSTEM = (
    "You are a strict judge checking whether a Falcon program's output correctly solves a problem statement. "
    "Reply with YES on the first line if the output is correct, or NO on the first line if it is wrong or incomplete. "
    "Follow with one short sentence explaining your verdict."
)

# ---------------------------------------------------------------------------
# App Inventor skip logic — copied from solve.py
# ---------------------------------------------------------------------------

SKIP_KEYWORDS = ["@Button", "@Label", "@TextBox", "@Notifier", "@Web", "@Camera",
                 "when ", "openScreen", "closeScreen", "getStartValue",
                 "openScreenWithValue", "closeScreenWithValue", "closeApp",
                 "getPlainStartText", "any Button", "any Label"]


def should_skip(problem_text: str) -> bool:
    return any(kw.lower() in problem_text.lower() for kw in SKIP_KEYWORDS)


# ---------------------------------------------------------------------------
# Progress helpers
# ---------------------------------------------------------------------------

def load_progress(progress_path: str) -> set[int]:
    if os.path.exists(progress_path):
        with open(progress_path, encoding="utf-8") as f:
            return set(json.load(f))
    return set()


def save_progress(progress_path: str, done: set[int], lock: threading.Lock) -> None:
    with lock:
        with open(progress_path, "w", encoding="utf-8") as f:
            json.dump(sorted(done), f)


# ---------------------------------------------------------------------------
# Code extraction
# ---------------------------------------------------------------------------

def extract_code(text: str) -> str | None:
    """Extract Falcon code from a ```falcon ... ``` block. Returns None if not found."""
    match = re.search(r"```falcon\s*\n(.*?)\n```", text, re.DOTALL)
    if match:
        return match.group(1).strip()
    # Fallback: any fenced block
    match = re.search(r"```\w*\s*\n(.*?)\n```", text, re.DOTALL)
    if match:
        return match.group(1).strip()
    return None


# ---------------------------------------------------------------------------
# Run & verify — copied/adapted from solve.py
# ---------------------------------------------------------------------------

def run_code(binary: str, code: str, timeout: int = 10) -> tuple[bool, str, str]:
    """Run Falcon code; returns (success, stdout, stderr)."""
    with tempfile.NamedTemporaryFile(suffix=".mist", mode="w", delete=False, encoding="utf-8") as f:
        f.write(code)
        tmp_path = f.name
    try:
        result = subprocess.run(
            [binary, "run", tmp_path],
            capture_output=True, text=True, timeout=timeout,
        )
        return result.returncode == 0, result.stdout, result.stderr
    except subprocess.TimeoutExpired:
        return False, "", "Error: execution timed out"
    finally:
        os.unlink(tmp_path)


def verify_output(client: anthropic.Anthropic, model: str,
                  problem_text: str, code: str, output: str) -> tuple[bool, str]:
    """Ask Claude whether the program output satisfies the problem. Returns (ok, reason)."""
    content = (
        f"Problem:\n{problem_text}\n\n"
        f"Falcon code written:\n```\n{code}\n```\n\n"
        f"Program output:\n{output if output.strip() else '(no output)'}"
    )
    response = client.messages.create(
        model=model,
        max_tokens=128,
        system=VERIFY_SYSTEM,
        messages=[{"role": "user", "content": content}],
    )
    reply = response.content[0].text.strip()
    first_line = reply.splitlines()[0].strip().upper()
    return first_line.startswith("YES"), reply


# ---------------------------------------------------------------------------
# Core: generate one attempt (extended thinking call)
# ---------------------------------------------------------------------------

def generate_attempt(client: anthropic.Anthropic, model: str, budget: int,
                     problem_text: str, prev_error: str | None = None) -> tuple[str, str] | None:
    """
    Call Claude with extended thinking for one attempt.
    Returns (thinking_text, code_text) or None if no code block found.
    prev_error: if set, embed the previous failure in the user message for context.
    """
    user_content = problem_text
    if prev_error:
        user_content = (
            f"{problem_text}\n\n"
            f"Note: a previous attempt failed with this error:\n{prev_error}\n"
            f"Please fix the issue and produce a correct solution."
        )

    response = client.messages.create(
        model=model,
        max_tokens=16000,
        thinking={"type": "enabled", "budget_tokens": budget},
        system=SYSTEM_PROMPT,
        messages=[{"role": "user", "content": user_content}],
    )

    thinking_text = next(
        (b.thinking for b in response.content if b.type == "thinking"), ""
    )
    text_block = next(
        (b.text for b in response.content if b.type == "text"), ""
    ).strip()

    code = extract_code(text_block)
    if code is None:
        return None

    return thinking_text, code


# ---------------------------------------------------------------------------
# Worker: process a chunk of entries
# ---------------------------------------------------------------------------

def process_entry(
    idx: int,
    entry: dict,
    think_client: anthropic.Anthropic,
    verify_client: anthropic.Anthropic,
    args: argparse.Namespace,
) -> tuple[bool, dict | None, str]:
    """
    Process one entry. Returns (success, record_or_None, status_message).
    """
    problem_text = entry["messages"][0]["content"]
    prev_error: str | None = None

    for attempt in range(args.retries + 1):
        # 1. Generate reasoning + code via extended thinking
        result = generate_attempt(think_client, args.model, args.budget, problem_text, prev_error)
        if result is None:
            prev_error = "No ```falcon code block found in response."
            continue

        thinking_text, code = result

        # 2. Compile & run locally
        ok, stdout, stderr = run_code(args.binary, code)
        if not ok:
            error_detail = stderr.strip() or stdout.strip() or "non-zero exit"
            prev_error = f"Compilation/runtime error:\n{error_detail}"
            continue

        # 3. Verify output with Claude-as-judge
        correct, reason = verify_output(verify_client, args.verify_model, problem_text, code, stdout)
        if not correct:
            prev_error = (
                f"Code ran but output was wrong.\n"
                f"Output was:\n{stdout.strip() or '(none)'}\n"
                f"Judge said: {reason}"
            )
            continue

        # Success — build the final assistant message
        assistant_content = f"<think>\n{thinking_text}\n</think>\n```falcon\n{code}\n```"
        record = {
            "messages": [
                {"role": "user", "content": problem_text},
                {"role": "assistant", "content": assistant_content},
            ]
        }
        return True, record, f"OK (attempt {attempt + 1})"

    return False, None, f"FAILED after {args.retries + 1} attempts. Last error: {prev_error}"


def process_chunk(
    chunk: list[tuple[int, dict]],
    args: argparse.Namespace,
    done: set[int],
    write_lock: threading.Lock,
    progress_lock: threading.Lock,
    counters: dict,
    counters_lock: threading.Lock,
    progress_path: str,
) -> None:
    think_client = anthropic.Anthropic()
    verify_client = anthropic.Anthropic()

    for idx, entry in chunk:
        success, record, status = process_entry(idx, entry, think_client, verify_client, args)

        if success and record is not None:
            with write_lock:
                with open(args.output, "a", encoding="utf-8") as f:
                    f.write(json.dumps(record, ensure_ascii=False) + "\n")
                    f.flush()

        with progress_lock:
            done.add(idx)
        save_progress(progress_path, done, progress_lock)

        with counters_lock:
            if success:
                counters["done"] += 1
            else:
                counters["failed"] += 1
            total = counters["total"]
            tag = "OK" if success else "FAIL"
            print(f"  [{counters['done'] + counters['failed']}/{total}] idx={idx} {tag} — {status}", flush=True)


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

def main() -> None:
    parser = argparse.ArgumentParser(description="Generate reasoning traces for MASTER.jsonl")
    parser.add_argument("--input",        default="answers/MASTER.jsonl")
    parser.add_argument("--output",       default="answers/MASTER_REASON.jsonl")
    parser.add_argument("--binary",       default="lang/Falcon", help="Falcon binary path")
    parser.add_argument("--workers",      type=int, default=5)
    parser.add_argument("--chunk",        type=int, default=100)
    parser.add_argument("--model",        default="claude-3-7-sonnet-20250219",
                                          help="Extended thinking model")
    parser.add_argument("--verify-model", default="claude-haiku-4-5-20251001",
                                          dest="verify_model",
                                          help="Model for output verification judge")
    parser.add_argument("--budget",       type=int, default=8000,
                                          help="Thinking token budget per attempt")
    parser.add_argument("--retries",      type=int, default=2,
                                          help="Max retries per entry on failure")
    parser.add_argument("--limit",        type=int, default=0,
                                          help="Max entries to process (0 = unlimited)")
    args = parser.parse_args()

    if not os.path.isfile(args.binary):
        sys.exit(f"Error: Falcon binary not found at '{args.binary}'. Build it first.")

    # Load input
    with open(args.input, encoding="utf-8") as f:
        all_entries = [json.loads(line) for line in f if line.strip()]
    print(f"Loaded {len(all_entries)} entries from {args.input}")

    progress_path = args.output + ".progress.json"
    done: set[int] = load_progress(progress_path)
    print(f"Already done: {len(done)} entries")

    # Filter: skip App Inventor + already completed
    pending: list[tuple[int, dict]] = []
    skipped_app = 0
    for idx, entry in enumerate(all_entries):
        if idx in done:
            continue
        problem_text = entry["messages"][0]["content"]
        if should_skip(problem_text):
            skipped_app += 1
            continue
        pending.append((idx, entry))

    if args.limit > 0:
        pending = pending[: args.limit]

    print(f"Skipped (App Inventor): {skipped_app}")
    print(f"To process: {len(pending)} entries with {args.workers} workers")
    print(f"Model: {args.model} | Verify: {args.verify_model} | Budget: {args.budget} | Retries: {args.retries}")

    if not pending:
        print("Nothing to do.")
        return

    # Split into chunks
    chunks: list[list[tuple[int, dict]]] = [
        pending[i : i + args.chunk] for i in range(0, len(pending), args.chunk)
    ]
    print(f"Chunks: {len(chunks)} x ~{args.chunk}\n")

    write_lock = threading.Lock()
    progress_lock = threading.Lock()
    counters_lock = threading.Lock()
    counters = {"done": 0, "failed": 0, "total": len(pending)}

    with ThreadPoolExecutor(max_workers=args.workers) as executor:
        futures = [
            executor.submit(
                process_chunk,
                chunk, args, done,
                write_lock, progress_lock,
                counters, counters_lock,
                progress_path,
            )
            for chunk in chunks
        ]
        for fut in as_completed(futures):
            exc = fut.exception()
            if exc:
                print(f"Chunk raised: {exc}", file=sys.stderr)

    print(f"\nDone. Written: {counters['done']}, Failed/Skipped: {counters['failed']}")
    print(f"Output: {args.output}")


if __name__ == "__main__":
    main()
