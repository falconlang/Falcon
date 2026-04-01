import json
import re
from pathlib import Path

INPUT_FILE = Path("/home/kumaraswamy/Documents/falcon/finetune/dataset/1.8k-codes-steps-output/MASTER.jsonl")
OUTPUT_DIR = Path("/home/kumaraswamy/Documents/falcon/finetune/dataset/1.8k-codes-steps-output/curriculum")

CODE_RE = re.compile(r'```falcon\n(.*?)```', re.DOTALL)


def extract_falcon_code(content: str) -> str:
    match = CODE_RE.search(content)
    if match:
        return match.group(1)
    return content.strip()


def extract_func_body(code: str, after_pos: int) -> str:
    eq_pos = code.find('=', after_pos)
    brace_pos = code.find('{', after_pos)

    if brace_pos == -1 and eq_pos == -1:
        return ''
    if brace_pos == -1:
        nl = code.find('\n', eq_pos + 1)
        return code[eq_pos + 1:nl] if nl != -1 else code[eq_pos + 1:]
    if eq_pos != -1 and eq_pos < brace_pos:
        brace_after_eq = code.find('{', eq_pos + 1)
        if brace_after_eq != -1:
            brace_pos = brace_after_eq

    depth = 0
    for i in range(brace_pos, len(code)):
        if code[i] == '{':
            depth += 1
        elif code[i] == '}':
            depth -= 1
            if depth == 0:
                return code[brace_pos:i + 1]
    return code[brace_pos:]


def detect_nested_loops(code: str) -> bool:
    for m in re.finditer(r'\b(for|while)\s*\(', code):
        brace_pos = code.find('{', m.end())
        if brace_pos == -1:
            continue
        depth = 0
        body_end = brace_pos
        for i in range(brace_pos, len(code)):
            if code[i] == '{':
                depth += 1
            elif code[i] == '}':
                depth -= 1
                if depth == 0:
                    body_end = i
                    break
        body = code[brace_pos:body_end]
        if re.search(r'\b(for|while)\s*\(', body):
            return True
    return False


def is_self_recursive(code: str) -> bool:
    for m in re.finditer(r'\bfunc\s+(\w+)\s*\(', code):
        name = m.group(1)
        body = extract_func_body(code, m.end())
        if re.search(r'\b' + re.escape(name) + r'\s*\(', body):
            return True
    return False


def has_nested_func_def(code: str) -> bool:
    for m in re.finditer(r'\bfunc\s+(\w+)\s*\(', code):
        body = extract_func_body(code, m.end())
        if re.search(r'\bfunc\s+\w+\s*\(', body):
            return True
    return False


def classify(code: str) -> int:
    has_loop = bool(re.search(r'\b(for|while)\s*\(', code))
    has_if = bool(re.search(r'\bif\s*\(', code))
    has_lambda = bool(re.search(r'\{[^{}]*->', code))
    has_ho = bool(re.search(r'\.(map|filter|sort|reduce)\s*[\({]', code))
    func_count = len(re.findall(r'\bfunc\s+\w+\s*\(', code))
    multi_func = func_count >= 2

    if (is_self_recursive(code)
            or detect_nested_loops(code)
            or multi_func
            or (has_loop and has_lambda)
            or has_nested_func_def(code)):
        return 6

    if has_lambda or has_ho:
        return 5

    if has_loop:
        return 4

    if has_if:
        return 3

    non_blank = [l for l in code.strip().splitlines() if l.strip()]
    local_count = len(re.findall(r'\blocal\s+\w+', code))
    if len(non_blank) > 3 or local_count >= 2 or func_count >= 1:
        return 2

    return 1


def main():
    OUTPUT_DIR.mkdir(parents=True, exist_ok=True)

    entries = []
    with open(INPUT_FILE, 'r', encoding='utf-8') as f:
        for line in f:
            line = line.strip()
            if line:
                entries.append(json.loads(line))

    print(f"Loaded {len(entries)} entries from MASTER.jsonl")

    classified = []
    level_counts = {i: 0 for i in range(1, 7)}

    for i, entry in enumerate(entries):
        assistant_content = entry["messages"][1]["content"]
        code = extract_falcon_code(assistant_content)
        level = classify(code)
        classified.append((level, entry))
        level_counts[level] += 1
        if (i + 1) % 200 == 0:
            print(f"  Classified {i + 1}/{len(entries)}...")

    level_files = {}
    for lvl in range(1, 7):
        path = OUTPUT_DIR / f"level_{lvl}.jsonl"
        level_files[lvl] = open(path, 'w', encoding='utf-8')

    for level, entry in classified:
        level_files[level].write(json.dumps(entry, ensure_ascii=False) + "\n")

    for f in level_files.values():
        f.close()

    curriculum_path = OUTPUT_DIR / "curriculum.jsonl"
    sorted_entries = sorted(classified, key=lambda x: x[0])
    with open(curriculum_path, 'w', encoding='utf-8') as f:
        for level, entry in sorted_entries:
            f.write(json.dumps(entry, ensure_ascii=False) + "\n")

    total = len(entries)
    print(f"\nClassification complete. {total} entries total.\n")
    print(f"{'Level':<8} {'Count':>6}  {'Pct':>6}  File")
    print("-" * 50)
    for lvl in range(1, 7):
        cnt = level_counts[lvl]
        pct = cnt / total * 100
        print(f"Level {lvl}   {cnt:>6}   {pct:>5.1f}%  level_{lvl}.jsonl")
    print("-" * 50)
    print(f"{'Total':<8} {total:>6}")
    print(f"\nCombined: curriculum/curriculum.jsonl ({total} entries, sorted L1->L6)")


if __name__ == "__main__":
    main()
