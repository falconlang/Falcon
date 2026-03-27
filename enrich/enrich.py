import groq
import json
import time
import os
import subprocess
import tempfile
from tqdm import tqdm

client = groq.Groq(api_key="")

FALCON_BIN  = "/home/kumaraswamy/Documents/falcon/lang/Falcon"
input_file    = "MASTER.jsonl"
output_file   = "MASTER_enriched.jsonl"
progress_file = "progress.txt"

FALCON_SPEC = """
Falcon is a DSL designed for App Inventor. Key facts:
- 1-based indexing
- Dynamically typed, no variable declarations
- No return statement — last expression in body is returned
- No try-catch or throw
- === (text equals), !== (text not equals), << (text less than), >> (text greater than)
- _ is the string join operator e.g. "Hello " _ "World!"
- ? operator checks type: "Hello" ? text, [] ? emptyList
- Lists passed by reference, use copyList() to avoid mutation
- func name(args) = { body } for result functions
- func name(args) { body } for void functions
- for (i: 1 .. n step 1) for range loops
- for (x in list) for each loops
- List lambdas: .map{}, .filter{}, .sort{}, .reduce(), .min{}, .max{}
- .sort { m, n -> bool } where bool means "m precedes n"
- textLen(), segment(from, length), trim(), split(), join(), etc.
- listLen(), add(), remove(), insert(), indexOf(), slice(), etc.
- copyList(list) creates a shallow copy
"""

EXPLAIN_PROMPT = """You are a code explainer for the Falcon DSL language.

Here is the Falcon DSL specification for reference:
{spec}

Explain the following Falcon code step by step in technical detail.
Cover loop logic, variable tracking and edge cases.
Do NOT write any code. Do NOT use code blocks. Use numbered steps only.

Code:
{code}"""


def run_falcon(code_block):
    """Extract code from ```falcon block and run it, return stdout or None."""
    # Strip the ```falcon ... ``` wrapper
    code = code_block.strip()
    if code.startswith("```falcon"):
        code = code[len("```falcon"):].strip()
    if code.endswith("```"):
        code = code[:-3].strip()

    # Write to a temp file and execute
    with tempfile.NamedTemporaryFile(
        mode="w", suffix=".fal", delete=False
    ) as tmp:
        tmp.write(code)
        tmp_path = tmp.name

    try:
        result = subprocess.run(
            [FALCON_BIN, "run", tmp_path],
            capture_output=True,
            text=True,
            timeout=10
        )
        output = result.stdout.strip()
        return output if output else None
    except subprocess.TimeoutExpired:
        return "[Timeout]"
    finally:
        os.unlink(tmp_path)  # always clean up temp file


# Load already processed indices
processed = set()
if os.path.exists(progress_file):
    with open(progress_file, "r") as f:
        processed = set(int(x.strip()) for x in f.readlines())
    print(f"Resuming — {len(processed)} already done\n")

# Load all samples
samples = []
with open(input_file, "r") as f:
    for line in f:
        samples.append(json.loads(line.strip()))

total         = len(samples)
remaining     = total - len(processed)
last_completed = max(processed) if processed else -1
session_count  = 0

print(f"Total    : {total}")
print(f"Done     : {len(processed)}")
print(f"Remaining: {remaining}\n")

with open(output_file, "a") as fout, open(progress_file, "a") as prog:

    pbar = tqdm(
        total=total,
        initial=len(processed),
        unit="sample",
        dynamic_ncols=True,
        bar_format="{l_bar}{bar}| {n}/{total} [{elapsed}<{remaining}, {rate_fmt}]"
    )

    for i, sample in enumerate(samples):
        if i in processed:
            continue

        code        = sample["messages"][1]["content"]
        user_prompt = sample["messages"][0]["content"]

        pbar.set_description(f"Entry {i+1} | {user_prompt[:35]}")

        # Step 1 — run the Falcon code
        falcon_output = run_falcon(code)
        if falcon_output:
            output_section = f"\n\nOutput:\n{falcon_output}"
        else:
            output_section = ""
        pbar.write(f"\n{'='*60}")
        pbar.write(f"  ENTRY [{i+1}/{total}]")
        pbar.write(f"  PROMPT : {user_prompt[:60]}")
        pbar.write(f"  OUTPUT : {falcon_output if falcon_output else '[No output]'}")
        pbar.write(f"{'='*60}")
        # Step 2 — get explanation from Groq
        while True:
            try:
                response = client.chat.completions.create(
                    model="llama-3.3-70b-versatile",
                    messages=[
                        {
                            "role": "user",
                            "content": EXPLAIN_PROMPT.format(
                                spec=FALCON_SPEC,
                                code=code
                            )
                        }
                    ],
                    max_tokens=512
                )

                explanation = response.choices[0].message.content

                # Step 3 — combine code + explanation + output
                enriched_content = (
                    f"{code}"
                    f"\n\nHere's how it works:\n{explanation}"
                    f"{output_section}"
                )

                enriched = {
                    "messages": [
                        {"role": "user",      "content": user_prompt},
                        {"role": "assistant", "content": enriched_content}
                    ]
                }

                fout.write(json.dumps(enriched) + "\n")
                fout.flush()

                prog.write(f"{i}\n")
                prog.flush()

                last_completed  = i
                session_count  += 1
                processed.add(i)

                pbar.set_postfix(
                    words=len(explanation.split()),
                    output="✓" if falcon_output else "∅",
                    last=i+1
                )
                pbar.update(1)

                time.sleep(3)
                break

            except groq.RateLimitError:
                pbar.write(f"\n  ⚠ Rate limit hit at entry {i+1}/{total}")
                pbar.write(f"  ✓ Last completed : entry {last_completed+1} (index {last_completed})")
                pbar.write(f"  ✗ Remaining      : {total - len(processed)} entries")
                pbar.write(f"  → Re-run tomorrow, script will resume from entry {last_completed+2}")
                pbar.close()
                exit(0)

            except Exception as e:
                pbar.write(f"  ✗ Error at entry {i+1}: {e} — skipping")
                break

    pbar.close()

done = len(processed)
print(f"\n{'='*60}")
print(f"  SESSION COMPLETE")
print(f"  Processed this run : {session_count}")
print(f"  Total done so far  : {done}")
print(f"  Remaining          : {total - done}")
if done < total:
    print(f"  → Re-run tomorrow to continue from entry {last_completed+2}")
else:
    print(f"  → All {total} samples enriched! ✓")
print(f"{'='*60}")
