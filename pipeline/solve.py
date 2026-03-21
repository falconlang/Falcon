#!/usr/bin/env python3
"""pipeline/solve.py — Solve Falcon problems using Claude and test with the runtime.

Usage:
    python3 pipeline/solve.py [options]

Options:
    --problems DIR    Path to problems directory (default: problems/)
    --binary PATH     Path to Falcon binary (default: lang/Falcon)
    --output PATH     Output JSONL path (default: dataset/pairs.jsonl)
    --retries N       Max retries per problem on failure (default: 2)
    --model ID        Claude model ID (default: claude-sonnet-4-6)
    --limit N         Process at most N problems (default: unlimited)
"""

import argparse
import json
import os
import subprocess
import sys
import tempfile
from pathlib import Path

import anthropic

# ---------------------------------------------------------------------------
# Falcon language reference — embedded into the system prompt
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

SYSTEM_PROMPT = f"""You are an expert Falcon programmer. Write correct, clean Falcon code to solve the given problem.

{FALCON_GUIDE}

Rules:
- Output ONLY the raw Falcon code — no explanation, no markdown fences, no comments unless needed.
- Use `println()` for all output.
- Do NOT use App Inventor components (`@Button`, `when ...`, `openScreen`, etc.) — this is a CLI runtime.
- Follow all quirks listed above, especially 1-based indexing and no `return` keyword.
- Keep solutions concise and correct.
"""

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

SKIP_KEYWORDS = ["@Button", "@Label", "@TextBox", "@Notifier", "@Web", "@Camera",
                 "when ", "openScreen", "closeScreen", "getStartValue",
                 "openScreenWithValue", "closeScreenWithValue", "closeApp",
                 "getPlainStartText", "any Button", "any Label"]


def should_skip(problem_text: str) -> bool:
    """Skip problems that require App Inventor components."""
    return any(kw.lower() in problem_text.lower() for kw in SKIP_KEYWORDS)


def solve(client: anthropic.Anthropic, model: str, problem_text: str,
          prev_attempts: list[dict] | None = None) -> str:
    """Ask Claude to generate Falcon code for the problem."""
    messages: list[dict] = [{"role": "user", "content": problem_text}]

    if prev_attempts:
        for attempt in prev_attempts:
            messages.append({"role": "assistant", "content": attempt["code"]})
            feedback = f"That code failed.\nError: {attempt['error']}"
            if attempt.get("output"):
                feedback += f"\nProgram output was:\n{attempt['output']}"
            feedback += "\nPlease fix it and output only the corrected Falcon code."
            messages.append({"role": "user", "content": feedback})

    response = client.messages.create(
        model=model,
        max_tokens=1024,
        system=SYSTEM_PROMPT,
        messages=messages,
    )
    return response.content[0].text.strip()


VERIFY_SYSTEM = (
    "You are a strict judge checking whether a Falcon program's output correctly solves a problem statement. "
    "Reply with YES on the first line if the output is correct, or NO on the first line if it is wrong or incomplete. "
    "Follow with one short sentence explaining your verdict."
)


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


def run_code(binary: str, code: str, timeout: int = 10) -> tuple[bool, str, str]:
    """Run Falcon code; returns (success, stdout, stderr)."""
    with tempfile.NamedTemporaryFile(suffix=".mist", mode="w", delete=False, encoding="utf-8") as f:
        f.write(code)
        tmp_path = f.name
    try:
        result = subprocess.run(
            [binary, "run", tmp_path],
            capture_output=True, text=True, timeout=timeout
        )
        return result.returncode == 0, result.stdout, result.stderr
    except subprocess.TimeoutExpired:
        return False, "", "Error: execution timed out"
    finally:
        os.unlink(tmp_path)


def load_progress(progress_path: str) -> set[str]:
    if os.path.exists(progress_path):
        with open(progress_path, encoding="utf-8") as f:
            return set(json.load(f))
    return set()


def save_progress(progress_path: str, done: set[str]) -> None:
    with open(progress_path, "w", encoding="utf-8") as f:
        json.dump(sorted(done), f)


def load_problems(problems_dir: str) -> list[tuple[str, str]]:
    """Return list of (id, problem_text) sorted by numeric id."""
    problems: list[tuple[str, str]] = []
    for p in sorted(Path(problems_dir).glob("PROBLEM*.json")):
        with open(p, encoding="utf-8") as f:
            data = json.load(f)
        for pid, text in data.items():
            problems.append((str(pid), str(text)))
    # Sort numerically
    problems.sort(key=lambda x: int(x[0]) if x[0].isdigit() else 0)
    return problems


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

def main() -> None:
    parser = argparse.ArgumentParser(description="Solve Falcon problems with Claude and test them.")
    parser.add_argument("--problems", default="problems/", help="Problems directory")
    parser.add_argument("--binary", default="lang/Falcon", help="Falcon binary path")
    parser.add_argument("--output", default="dataset/pairs.jsonl", help="Output JSONL path")
    parser.add_argument("--retries", type=int, default=1, help="Max retries per problem (default: 1 = 2 total attempts)")
    parser.add_argument("--model", default="claude-sonnet-4-6", help="Claude model ID")
    parser.add_argument("--limit", type=int, default=0, help="Max problems to process (0 = all)")
    args = parser.parse_args()

    # Validate binary
    if not os.path.isfile(args.binary):
        sys.exit(f"Error: Falcon binary not found at '{args.binary}'. Build it first:\n  cd lang && go build -o Falcon .")

    # Ensure output directory exists
    Path(args.output).parent.mkdir(parents=True, exist_ok=True)
    progress_path = str(Path(args.output).parent / "progress.json")

    client = anthropic.Anthropic()  # reads ANTHROPIC_API_KEY from env

    problems = load_problems(args.problems)
    done = load_progress(progress_path)

    print(f"Loaded {len(problems)} problems. {len(done)} already solved.")

    processed = 0
    solved = 0
    skipped = 0
    failed = 0

    with open(args.output, "a", encoding="utf-8") as out_f:
        for pid, problem_text in problems:
            if pid in done:
                continue
            if args.limit and processed >= args.limit:
                break

            processed += 1

            if should_skip(problem_text):
                print(f"[{pid}] SKIP (App Inventor component)")
                skipped += 1
                done.add(pid)
                save_progress(progress_path, done)
                continue

            print(f"[{pid}] Solving…", end="", flush=True)

            attempts: list[dict] = []
            success = False
            final_code = ""
            final_output = ""

            for attempt_num in range(args.retries + 1):
                code = solve(client, args.model, problem_text,
                             prev_attempts=attempts if attempt_num > 0 else None)
                ok, stdout, stderr = run_code(args.binary, code)

                if ok:
                    correct, reason = verify_output(client, args.model, problem_text, code, stdout)
                    if correct:
                        success = True
                        final_code = code
                        final_output = stdout
                        break
                    else:
                        # Ran fine but output is wrong — feed back output + verdict
                        attempts.append({
                            "code": code,
                            "error": f"wrong output — {reason}",
                            "output": stdout,
                        })
                        print(f" wrong_output({attempt_num + 1})", end="", flush=True)
                else:
                    error_msg = stderr.strip() or "non-zero exit"
                    attempts.append({"code": code, "error": error_msg, "output": stdout})
                    print(f" error({attempt_num + 1})", end="", flush=True)

            if success:
                record = {
                    "id": pid,
                    "problem": problem_text,
                    "code": final_code,
                    "output": final_output,
                }
                out_f.write(json.dumps(record, ensure_ascii=False) + "\n")
                out_f.flush()
                done.add(pid)
                save_progress(progress_path, done)
                solved += 1
                print(f" OK")
            else:
                failed += 1
                print(f" FAIL")

    print(f"\nDone. Processed={processed}, solved={solved}, skipped={skipped}, failed={failed}")
    print(f"Output: {args.output}")


if __name__ == "__main__":
    main()
