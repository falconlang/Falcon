#!/usr/bin/env python3
"""pipeline/generate_problems.py — Generate new, distinct Falcon problem statements.

Loads the existing problem set, then asks Claude to generate NEW problems that
don't overlap. Works in batches of BATCH_SIZE, deduplicates against existing,
and accumulates until the target count is reached.

Usage:
    python3 pipeline/generate_problems.py [--target 500] [--output problems_clean/PROBLEM_NEW.json]
"""

import argparse
import json
import math
import re
from pathlib import Path

import anthropic

# ---------------------------------------------------------------------------
# Topics to target — spread across underrepresented and completely new angles
# ---------------------------------------------------------------------------

TOPICS = [
    # Underrepresented in existing set
    "while loops with complex exit conditions",
    "for-each loops over dictionaries with key-value destructuring",
    "local variable scoping and shadowing",
    "global variable mutation inside nested functions",
    "type-checking using the ? operator (text, number, list, dict, bin, hexa, emptyList, emptyText)",
    "the pair operator `:` to build dictionary entries",
    "text join operator `_` in complex string construction",
    "color values: makeColor, splitColor, hex color manipulation",
    "CSV helpers: toCsvRow, toCsvTable, csvRowToList, csvTableToList",
    "formatDecimal for fixed-point output",
    "base conversion: bin, hexa, octal, decToHex, decToBin, hexToDec, binToDec",
    "bitwise operators: &, |, ~ (xor)",
    "lexicographic operators: ===, !==, <<, >>",
    "list slice, allButFirst, allButLast",
    "list lookupInPairs and pairsToDict round-trips",
    "dict walkTree and getAtPath / setAtPath for nested paths",
    "dict toPairs, keys(), values() and iterating them",
    "string segment, splitAtFirst, splitAtAny, splitAtFirstOfAny",
    "replaceFrom and replaceFromLongestFirst with dictionary maps",
    "list reduce with a non-zero initial value",
    "chained lambdas: filter then map then reduce in one pipeline",
    "sort lambda with a custom comparator",
    "min/max lambda to find the extremal element by a computed key",
    # Completely new application domains
    "unit conversion utilities (km↔miles, °C↔°F, kg↔lbs)",
    "calendar and date arithmetic (day-of-week, leap year, days between)",
    "simple stack implemented with a list (push, pop, peek, isEmpty)",
    "simple queue implemented with a list (enqueue, dequeue, size)",
    "priority queue simulation using a sorted list",
    "two-pointer and sliding-window patterns on lists",
    "run-length encoding and decoding",
    "frequency map (histogram) of a list",
    "anagram detection and character frequency comparison",
    "palindrome checking for strings and lists",
    "Roman numeral conversion (integer ↔ Roman)",
    "fizzbuzz variants and divisibility patterns",
    "number formatting: thousands separator, zero-padding, sign display",
    "simple expression evaluator for +/- on a list of tokens",
    "Luhn algorithm for credit-card checksum",
    "grade-letter assignment from a numeric score with weighted average",
    "inventory system: restock, sell, low-stock alert",
    "time-series running average and exponential moving average",
    "sparse matrix as a dict of dicts",
    "adjacency list graph: add edge, neighbours, degree",
    "topological sort on a small DAG",
    "Levenshtein (edit distance) between two short strings",
    "Huffman frequency count (not full tree) — just character frequencies",
    "simple state machine with transitions stored in a dict",
    "memoisation pattern using a global dict cache",
    "co-routine-style generator simulation with a global index",
    "piping data through a list of single-argument functions",
    "matrix transpose and dot product",
    "polynomial evaluation using Horner's method",
]

BATCH_SIZE = 20   # problems per API call

SYSTEM = """You write problem statements for the Falcon programming language, designed to fine-tune small language models.

Falcon key facts:
- 1-based indexing
- No return keyword — last expression in a body is the return value
- Global vars accessed via `this.varName`; declared with `global x = …`
- Local vars declared with `local x = …`
- For loop: `for (i: 1 .. 10)` or `for (item in list)` or `for (k, v in dict)`
- While loop: `while (cond) { … }`
- Void func: `func name(args) { … }` — Result func: `func name(args) = { … }`
- List lambdas: `.map { x -> … }`, `.filter { x -> … }`, `.reduce(init) { x, acc -> … }`, `.sort { a, b -> bool }`, `.min { … }`, `.max { … }`
- Operators: `_` (text join), `?` (type check), `:` (pair), `~` (xor), `===`/`!==`/`<<`/`>>` (text compare)

Each problem statement must:
1. Be a single self-contained instruction (1–3 sentences, 80–220 chars).
2. Specify exact variable names, function signatures, sample inputs, and what to print.
3. Be solvable with 5–30 lines of Falcon code.
4. NOT reference App Inventor components (no @Button, no `when`, no openScreen).
5. Be meaningfully different from the others in this batch and from the examples shown.
"""


def load_existing(problems_dir: str) -> set[str]:
    """Return normalised texts of all existing problems."""
    existing: set[str] = set()
    for p in Path(problems_dir).glob("PROBLEM*.json"):
        with open(p, encoding="utf-8") as f:
            data = json.load(f)
        for text in data.values():
            existing.add(" ".join(text.lower().split()))
    return existing


def sample_existing(existing_texts: list[str], n: int = 12) -> list[str]:
    """Return n evenly-spaced examples to show Claude as 'already covered'."""
    if len(existing_texts) <= n:
        return existing_texts
    step = math.ceil(len(existing_texts) / n)
    return existing_texts[::step][:n]


def generate_batch(client: anthropic.Anthropic, model: str,
                   topic: str, examples: list[str], count: int) -> list[str]:
    """Ask Claude to produce `count` new problem statements on `topic`."""
    examples_block = "\n".join(f"- {e}" for e in examples)
    prompt = (
        f"Topic to focus on: **{topic}**\n\n"
        f"These problem statements already exist — do NOT repeat or paraphrase them:\n"
        f"{examples_block}\n\n"
        f"Write exactly {count} NEW, distinct Falcon problem statements on the topic above. "
        f"Number them 1 through {count}. Output only the numbered list, no extra text."
    )
    response = client.messages.create(
        model=model,
        max_tokens=2048,
        system=SYSTEM,
        messages=[{"role": "user", "content": prompt}],
    )
    raw = response.content[0].text.strip()
    # Parse numbered list
    problems: list[str] = []
    for line in raw.splitlines():
        line = line.strip()
        m = re.match(r"^\d+[\.\)]\s+(.+)$", line)
        if m:
            problems.append(m.group(1).strip())
    return problems


def normalise(text: str) -> str:
    return " ".join(text.lower().split())


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--target",   type=int, default=500,
                        help="Number of new problems to generate (default: 500)")
    parser.add_argument("--problems", default="problems_clean/",
                        help="Existing problems directory to check against")
    parser.add_argument("--output",   default="problems_clean/PROBLEM_NEW.json",
                        help="Output JSON file for new problems")
    parser.add_argument("--model",    default="claude-sonnet-4-6",
                        help="Claude model ID")
    args = parser.parse_args()

    client = anthropic.Anthropic()

    print(f"Loading existing problems from '{args.problems}'…")
    existing_norm = load_existing(args.problems)
    existing_list = list(existing_norm)
    print(f"  {len(existing_norm)} existing problems loaded.\n")

    new_problems: dict[str, str] = {}   # id → text
    next_id = 20001                      # start IDs well away from existing range

    # Round-robin over topics until we hit the target
    topic_cycle = TOPICS * math.ceil(args.target / (len(TOPICS) * BATCH_SIZE) + 1)
    topic_idx = 0

    while len(new_problems) < args.target:
        remaining = args.target - len(new_problems)
        batch_count = min(BATCH_SIZE, remaining + 5)  # ask a few extra to absorb dupes
        topic = topic_cycle[topic_idx % len(topic_cycle)]
        topic_idx += 1

        examples = sample_existing(existing_list, n=12)
        print(f"[{len(new_problems):>4}/{args.target}] Topic: {topic[:60]}…", end=" ", flush=True)

        try:
            batch = generate_batch(client, args.model, topic, examples, batch_count)
        except Exception as e:
            print(f"ERROR: {e}")
            continue

        added = 0
        for text in batch:
            if len(new_problems) >= args.target:
                break
            norm = normalise(text)
            if norm not in existing_norm and len(text) > 40:
                existing_norm.add(norm)   # prevent intra-run dupes too
                new_problems[str(next_id)] = text
                next_id += 1
                added += 1

        print(f"got {len(batch)}, kept {added}")

    # Write output
    Path(args.output).parent.mkdir(parents=True, exist_ok=True)
    with open(args.output, "w", encoding="utf-8") as f:
        json.dump(new_problems, f, indent=2, ensure_ascii=False)

    print(f"\nDone — {len(new_problems)} new problems written to '{args.output}'.")


if __name__ == "__main__":
    main()
