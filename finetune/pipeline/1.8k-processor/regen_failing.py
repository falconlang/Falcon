#!/usr/bin/env python3
"""
Regenerate code for failing MASTER.jsonl entries using Claude.
Verifies each regenerated entry passes before committing.
"""

import json, re, subprocess, tempfile, time
from pathlib import Path
import anthropic

MASTER = Path("finetune/dataset/1.8k-codes-steps-output/MASTER.jsonl")
FALCON = "lang/Falcon"

SYSTEM_PROMPT = """You write code in the Falcon programming language. Output ONLY a fenced code block with no explanation.

Falcon syntax rules:
- Returning function:   func name(args) = { ... last_expr_is_return }
- Void function:        func name(args) { ... }
- Local variable:       local x = expr
- Global variable:      global x = expr   (access with this.x)
- Print:                println(expr)
- Range for loop:       for (i: 1 .. n step 1) { }
- For-each loop:        for (item in list) { }
- Dict iteration:       for (k, v in dict) { }
- While loop:           while (cond) { }
- If expression:        if (cond) { a } else { b }   returns a value
- If statement:         if (cond) { } (no else, void)
- String concat:        "hello" _ name
- List methods:         .add(v), .remove(i), .listLen(), .slice(start,end), .sort { a,b -> a<b }, .map { x -> expr }, .filter { x -> bool }, .reduce(init) { x, acc -> expr }, .appendList(lst), .containsItem(v), .join(sep), .reverseList()
- Dict methods:         .set(k,v), .get(k,default), .containsKey(k), .delete(k), .keys(), .values(), .toPairs(), .copyDict(d)
- String methods:       .textLen(), .segment(start,len), .split(sep), .replace(a,b), .join(sep)
- Math:                 floor(), ceil(), round(), sqrt(), abs(), min(a,b), max(a,b), pow(b,e), randFloat(), randInt(lo,hi)
- Type checks:          expr ? list, expr ? dict, expr ? text, expr ? number, expr ? emptyList, expr ? base10
- Coerce string->number: str * 1  (NOT dec(var))
- Copy list:            copyList(lst)
- Format decimal:       formatDecimal(n, places)
- Color:                makeColor([r,g,b]), splitColor(c) returns [_,r,g,b]
- Break out of loop:    break

CRITICAL RULES:
1. .sort{}, .reverseList() return NEW lists — always assign back: x = x.sort { a,b -> a<b }
2. No bare variable in statement position inside loops/if-bodies (use break or restructure)
3. Lambda bodies for .map/.filter/.reduce/.sort end with the return expression
4. Functions called in void context (result not used) must be declared void (no = before {)
5. A returning function's last expression is its return value — no return keyword
6. Calling a returning function without using the result is an error
7. For early-exit patterns in loops, use break instead of bare variable expressions
8. Local variables inside a lambda body: the LAST expression is the return value of the lambda
9. dict.set() is void — does not return a value, so it's fine in statement position
10. Recursive calls whose results are discarded must be in void functions

Output format — exactly:
```falcon
<code here>
```"""

client = anthropic.Anthropic()

def extract_code(content):
    m = re.search(r'```falcon\n(.*?)```', content, re.DOTALL)
    return m.group(1).strip() if m else content.strip()

def replace_code_in_content(content, new_code):
    return re.sub(
        r'(```falcon\n).*?(```)',
        lambda m: m.group(1) + new_code + '\n' + m.group(2),
        content, flags=re.DOTALL
    )

def run_falcon(code):
    with tempfile.NamedTemporaryFile(mode='w', suffix='.mist', delete=False) as f:
        f.write(code); path = f.name
    r = subprocess.run([FALCON, "run", path], capture_output=True, text=True, timeout=10)
    Path(path).unlink(missing_ok=True)
    return r.returncode, r.stdout.strip(), r.stderr.strip()

def regenerate(prompt, current_code, err, attempt=1):
    user_msg = prompt
    if attempt > 1:
        user_msg += f"\n\nPrevious attempt failed with: {err}\nFix the issue."
    msg = client.messages.create(
        model="claude-opus-4-6",
        max_tokens=2048,
        system=SYSTEM_PROMPT,
        messages=[{"role": "user", "content": user_msg}]
    )
    return msg.content[0].text

def main():
    lines = MASTER.read_text(encoding="utf-8").splitlines()
    patched_lines = list(lines)

    fixed = skipped = 0

    for i, raw in enumerate(lines):
        try: entry = json.loads(raw)
        except: continue
        msgs = entry.get("messages", [])
        if len(msgs) < 2: continue

        code = extract_code(msgs[1]["content"])
        rc, out, err = run_falcon(code)
        if rc == 0:
            continue

        prompt = msgs[0]["content"]
        print(f"\n[{i}] {prompt[:70]!r}")
        print(f"  Error: {err.splitlines()[-1][:80]}")

        new_code = None
        for attempt in range(1, 4):
            response = regenerate(prompt, code, err, attempt)
            candidate = extract_code(response)
            rc2, out2, err2 = run_falcon(candidate)
            if rc2 == 0:
                new_code = candidate
                print(f"  PASS (attempt {attempt}) out={out2[:60]!r}")
                break
            else:
                print(f"  attempt {attempt} failed: {err2.splitlines()[-1][:70]}")
                err = err2
                time.sleep(1)

        if new_code is None:
            print(f"  SKIP — could not fix after 3 attempts")
            skipped += 1
            continue

        msgs[1]["content"] = replace_code_in_content(msgs[1]["content"], new_code)
        new_line = json.dumps(entry, ensure_ascii=False)
        try: json.loads(new_line)
        except json.JSONDecodeError as e:
            print(f"  SKIP — JSON round-trip failed: {e}"); skipped += 1; continue

        patched_lines[i] = new_line
        fixed += 1

    MASTER.write_text("\n".join(patched_lines), encoding="utf-8")
    print(f"\nDone. Fixed: {fixed}, Skipped: {skipped}")

if __name__ == "__main__":
    main()
