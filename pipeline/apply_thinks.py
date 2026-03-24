#!/usr/bin/env python3
"""
Apply prewritten think blocks to MASTER_REASON.jsonl for the 99 swapped entries.
Replaces only the <think>...</think> content; code block is left untouched.
"""

import json
import re
from pathlib import Path

MASTER_REASON = Path(__file__).parent.parent / "answer_reasoning" / "MASTER_REASON.jsonl"

# ─── Think blocks keyed by line index ─────────────────────────────────────────

THINKS = {
    112: "2×2 matrix multiplication: compute each output element as the dot product of a row of A and a column of B. r00 = a[1][1]*b[1][1] + a[1][2]*b[2][1], and so on for r01, r10, r11. Assemble and return as [[r00,r01],[r10,r11]].",
    120: "Filter the list with a lambda keeping elements greater than the threshold, then find the maximum with .max using a comparator lambda. Both are list lambda operations.",
    134: "Copy the list with copyList, then loop through odd 1-based indices (1, 3, 5… via step 2 range loop), multiplying each element by the factor in place.",
    151: "Copy dict a with copyDict, then iterate over b's key-value pairs. On conflict, add the two values; otherwise just set the key from b. Return the merged copy.",
    196: "Check cache.containsKey(key): if present, return cache.get(key). Otherwise compute the pre-evaluated calc value, store with cache.set, and return it.",
    197: "Use a for range loop 1..a.listLen() step 1, adding a[i] + b[i] to a result list at each position. 1-based indexing throughout.",
    206: "Copy the grid with copyList. Nested loops over rows and columns. For each cell, iterate neighbor deltas [-1,0,1]×[-1,0,1], skipping (0,0), to count live neighbors. Apply Conway's rules: alive with <2 or >3 neighbors → 0; dead with exactly 3 → 1.",
    223: "Recursive merge sort counting inversions. Base case: return [arr, 0] wrapped as a local list (bare list literal ambiguity). Split at floor(n/2), recurse on both halves. When merging, if right[j] < left[i], add remaining left-half elements to inv count. Return [merged, inv] as a local list.",
    227: "Build [1..n] list via for range. Fisher-Yates shuffle: loop i from n down to 2 (step -1), pick j = randInt(1,i), swap arr[i] and arr[j]. Verify by checking containsItem for each 1..8.",
    271: "First Gray bit is unchanged. For each subsequent bit, compare with the previous binary bit: same → '0', different → '1', appended via if-else. Convert the resulting binary string to decimal with binToDec.",
    280: "Count inversions with nested for loops: for each pair (i,j) with j>i, increment if perm[i]>perm[j]. Return 1 if mod(inv,2)==0 (even permutation), -1 otherwise.",
    301: "100 trials, each simulating 100 random steps: randInt(0,1)==1 gives +1, else -1. Accumulate abs(final position) per trial, divide total by 100. Output with formatDecimal for 2 decimal places.",
    312: "Define helper applyTextOp dispatching on 'upper', 'lower', 'trim', 'reverse', 'len'. composeTextOps iterates the ops list with for-each, threading the intermediate result through each operation.",
    320: "Split string on spaces with splitAtSpaces(). Slide a window of size n: outer loop i from 1 to listLen-n+1, inner loop j from i to i+n-1 collecting words, joined with space.",
    326: "Split on spaces with splitAtSpaces(), call reverseList() to reverse in place, then chain join(' ') to produce the reversed-word string.",
    353: "Initialize the LZW dict with single characters from a fixed alphabet string using segment. Greedy compression: extend current string while dict contains current+ch; otherwise output current's code, add combined to dict with nextCode, reset current to ch.",
    359: "Classic DP for matrix chain order. Build n×n dp table (n=dims.listLen()-1). Outer loop on chain length 2..n, inner loop i with j=i+len-1. Try all splits k; cost = dp[i][k] + dp[k+1][j] + dims[i]*dims[k+1]*dims[j+1]. Return dp[1][n].",
    360: "While loop with index i. Grab value at i, count the run with an inner while advancing while list[i+count]==val. Add [val, count] pair, advance i by count.",
    493: "Use .contains(' ') to check for a multi-word city. If found, print the full destination; otherwise split on ' ' and print the first segment.",
    571: "Check .contains('_') on the feature name. Split on '_' with .split(), then print parts[2] (1-based index 2 is the numeric part '03').",
    576: "A* using a sorted list as a min-priority queue. Each entry is [f, g, row, col, path]. Sort open list each step to get min-f entry. Use visited dict keyed by 'r,c' to avoid revisiting. Expand 4 cardinal neighbors with Manhattan heuristic. Record path when goal is reached.",
    603: "Declare global activeBuffs as []. addBuff guards with !this.activeBuffs.containsItem(buff) before adding. hasBuff returns this.activeBuffs.containsItem(buff). Access the global via this.",
    604: "BFS with a queue list (remove(1) for dequeue). Track visited and cameFrom dicts keyed by 'r,c' strings. On reaching goal, walk cameFrom backward from end to start, then reverseList to get correct order.",
    668: "O(n²): for each start index i scan forward with j, maintaining a seen list. Break when the character is already in seen via containsItem. Track the maximum window length.",
    675: "Same O(n²) approach as substring uniqueness: for each start i scan forward with j, maintaining a seen list with containsItem. Break on first duplicate, update maximum length.",
    680: "Copy dict a with copyDict. For each key-value in b: on conflict, dispatch on strategy — 'first' keeps a's value, 'second' uses b's, 'sum' adds them. Otherwise set b's value directly.",
    692: "Copy and sort the list with copyList + sort(). Iterate sorted copy: add each element to result only if !result.containsItem(b) — deduplication is easy since equal values are adjacent after sorting.",
    694: "For each plaintext character: find its 1-based index in plainWheel, output cipherWheel[idx]. Rotate both wheels: remove at idx and re-insert at position 2 for the standard Chaocipher step.",
    700: "Build a context dict with algo, key, iv. Define separate ctxEncrypt (XOR-style: mod(byte ~ key[1], 256)) and ctxDecrypt (same since XOR is symmetric). Return the context dict.",
    701: "Use global secretStore dict keyed by namespace. buildSecretManager initializes the entry. secretSet encodes each character as mod(charLen+13, 256) per byte. secretGet retrieves. secretRotate increments each byte mod 256 with .map lambda.",
    720: "IDEA multiplication treats 0 as 65536. a==0 → mod(65537-b,65537); b==0 → mod(65537-a,65537); otherwise mod(a*b,65537).",
    733: "Simple DER parse: tag=data[1], length=data[2], value=data.slice(3, 2+length). Return as a local list [tag, length, value].",
    755: "Generate n-1 random shares mod p, compute the last as (key - sum_of_others) mod p so all shares sum to key. Recover with .reduce using mod-addition.",
    775: "Romberg tableau for sin(x) from 0 to π. R[i][1] is the trapezoidal rule with step h=(b-a)/2^(i-1) and midpoints at (2k-1)*h. R[i][j] applies Richardson extrapolation: (4^(j-1)*R[i][j-1] - R[i-1][j-1]) / (4^(j-1)-1). Print each diagonal R[i][i].",
    776: "Generate 10 random particles at [randFloat()*10, randFloat()*10]. Nested loops i,j skipping i==j compute Euclidean distance sqrt(dx²+dy²). Track min distance and the two particle indices.",
    785: "Declare globals wavePrev, waveCurr, waveNext, each filled with 50 zeros. Initialize Gaussian pulse on waveCurr with exp(-0.5*x²/4) where x=i-25. waveStep computes r=(c*dt/dx)², then nextG[i] = 2*curr[i]-prev[i]+r*(curr[i+1]-2*curr[i]+curr[i-1]), with fixed boundaries at i=1 and 50.",
    790: "Secant method with f hardcoded as x³-x-2. At each step compute fa and fb, update root = b - fb*(b-a)/(fb-fa), shift a=b, b=root. Stop when |b-a| < tol or max iterations hit.",
    791: "Müller's method: fit a quadratic through p0,p1,p2 using divided differences d1,d2 and coefficient a. Compute the quadratic discriminant, choose the sign giving the larger denominator for stability. Update p0=p1, p1=p2, p2=root.",
    792: "Build augmented matrix [A|b]. For each column: find the row with max absolute value for partial pivoting, swap rows, eliminate below by subtracting factor*pivot_row from each row. Back-substitute from n to 1.",
    793: "Gram-Schmidt QR: extract columns of A. For each column j, subtract projections onto previous orthonormal q-vectors (storing dot products as R[k][j]), normalize to get q_j with R[j][j]=norm. Assemble Q from orthonormal column vectors, return {Q, R}.",
    802: "Thomas algorithm forward sweep: c'[1]=c[1]/b[1], d'[1]=d[1]/b[1]; for i>1: denom=b[i]-a[i]*c'[i-1], c'[i]=c[i]/denom, d'[i]=(d[i]-a[i]*d'[i-1])/denom. Back substitution: x[n]=d'[n], x[i]=d'[i]-c'[i]*x[i+1].",
    803: "Compute exponent as floor(log(abs(x))/log(10)), mantissa = x/10^exp. Round mantissa to precision decimal places. Concatenate mantissa _ 'e+' _ exp as a string.",
    822: "Compute 3×3 determinant with cofactor expansion along row 1. Build the adjugate matrix (transpose of cofactors), divide each element by det. Return the 3×3 inverse.",
    829: "Take absolute value of each signal sample with .map lambda. Apply exponential moving average: prev = alpha*x + (1-alpha)*prev for each sample. Collect smoothed values in env list.",
    845: "Normalize t = magnitude/max_mag. Interpolate r from 0→255 (t*255) and b from 136→0 ((1-t)*136) to blend violet to red. Use makeColor([r,0,b]).",
    846: "Normalize t = (value-lo)/(hi-lo). 'viridis': blue-to-yellow-green channel mapping. 'plasma': purple-to-yellow. 'hot': black-to-red-to-white with progressively activated green and blue channels. Use makeColor.",
    852: "Define rk4_step helper (hardcoded f(y)=-y). solveODE loops from t0 to tf, dispatching on method: 'euler' uses Euler step, 'rk4' calls rk4_step, 'leapfrog' uses half-step velocity. Return {ts, ys} dict.",
    857: "Recursive Cooley-Tukey FFT: base case n==1 returns [[signal[1],0]]. Split into even/odd indexed sublists, recurse. For each k, compute twiddle (angle=-2π(k-1)/n), butterfly: result[k]=E[k]+twiddle*O[k], result[k+half]=E[k]-twiddle*O[k]. Return [re,im] pairs.",
    858: "Compute mean with .reduce sum/n. Compute variance via .map (x-mean)² then .reduce sum/n. Normalize each sample to (x-mean)/sqrt(variance) with a final .map lambda.",
    865: "Declare global trampolineTable dict. needsTrampoline returns containsKey. addTrampoline computes an address (dictLen*16+65536) and stores {addr, target} dict. Access global via this.",
    867: "Get variable names with .keys(). Nested loops i,j (j>i). Intervals [s1,e1] and [s2,e2] interfere if s1≤e2 && s2≤e1. Add [vars[i], vars[j]] edge pairs to the result list.",
    872: "Iterate closures with for-each. For each closure, iterate its 'free' variable list, look up each in env_outer dict. Collect captured vars in a new dict, print with closure name.",
    873: "LL(1) simulation with explicit stack and input list. Loop until accepted or error. If top==input: pop stack, advance input. If top=='$' and input=='$': accept. Otherwise look up table[top+input] key, replace top with production symbols pushed in reverse order.",
    879: "Repeated integer halving: while x>1, set x=floor(x/2) and increment result. The count of halvings is the floor of log₂(n).",
    897: "Use chained .replace() calls to substitute two-character escape sequences: '\\\\n' with '\\\\n', '\\\\t' with '\\\\t', and '\"' with '\\\\\"'. The input already contains literal backslash-n/t sequences.",
    915: "Build a result dict mapping each block's id to a list of expressions from its instructions. Collect only instructions that contain an 'expr' key.",
    916: "Collect all store ops and all load ops in separate lists. For every store-load pair, add a may-alias entry {store_addr, load_addr, may_alias: true} — conservative analysis assumes any store may alias any load.",
    927: "Build escape info dict mapping function names to {escapes: []}. markEscape appends obj to the function's escapes list. doesEscape checks containsItem on that list.",
    929: "buildExceptionTable wraps fn and handlers list into a dict. findHandler iterates handlers checking pc within [start, end], returning the matching handler label or 'no handler'.",
    935: "Build refinement type as {base, predicate} dict. checkRefine evaluates base_ok (? type check operator) and pred_ok (value predicate string dispatch). Return {base_ok, pred_ok, valid} dict.",
    952: "Collect accesses per address into a dict (addr → [{tid,op}]). For each address, if any access is a write and multiple threads accessed it, report a race with the address and access list.",
    954: "Iterative selection sort: copy list, outer loop i from 1 to n-1 finding minIdx in i+1..n with inner loop. Swap res[i] with res[minIdx].",
    983: "Split on spaces with splitAtSpaces(), reverse the list in place with reverseList(), then join with join(' '). Return from an inner block.",
    1031: "Iterate words in the sentence using for-each over splitAtSpaces(). For each word, set dict[word] = word.reverse().",
    1032: "Passthrough stub: return the intervals list unchanged. A full merge would sort then merge overlapping ranges, but this simplified version returns as-is.",
    1039: "If condition is true, call fn(x) and return the result; otherwise return x. A simple conditional function application guard.",
    1046: "Chain six .replace() calls removing each punctuation character in sequence, replacing each with empty string.",
    1051: "Distribute floats into B buckets by computing index = floor(x*B)+1, clamped to B. Sort each bucket with a comparator lambda. Concatenate all sorted buckets in order.",
    1052: "Sort with a multi-field comparator lambda. For each spec in specs, compare the field values; on first difference, return the comparison based on the 'desc' flag. Continue to next spec if values are equal.",
    1064: "Validate required fields (containsKey check), uniqueness (track seen values per field in uniqueVals dict), and range bounds (min/max comparison). Collect all violation dicts with row index and error message.",
    1077: "For each index i, sum five values using an inner loop j from -2 to 2, clamping idx=i+j to [1,n]. Divide the sum by 5 for the centered moving average.",
    1090: "Loop through encoded string in steps of 2: segment(i,1) is the count as a number, segment(i+1,1) is the character. Append count copies of the character to result.",
    1101: "Stub implementation returning an empty dict. A full version would iterate all pairs of time series, compute Pearson correlation between each pair, and collect those exceeding threshold T.",
    1112: "For each text, split on ':' to get parts. Strip quotes from each part with .replace('\"',''). Set the cleaned key → value pair in the result dict.",
    1116: "Slide a window of size m (pattern length) across text with a range loop. Extract segment(i, m) and compare === to pattern. Return first match index i, or -1 if none found.",
    1126: "BFS from (px,py) using a queue list. Each entry is [row, col, distance]. Track visited positions by 'x,y' key strings. Expand 4 cardinal directions, checking bounds and non-wall cells. Return distance when goal is reached.",
    1128: "Global recipeGraph dict. crafting_tree populates it by iterating the recipes dict. dependencies does a DFS with an explicit stack, tracking a visited list, to collect all transitive ingredient dependencies.",
    1144: "One RK4 step for harmonic oscillator (f=-x). Compute four k-values for position and velocity. New state: x + dt/6*(k1x+2*k2x+2*k3x+k4x), v + dt/6*(k1v+2*k2v+2*k3v+k4v). Return [nx, nv] as a local list.",
    1153: "Compute filled = round(current/max * width). Loop i from 1 to width: append fillChar if i≤filled, emptyChar otherwise. Concatenate and return the bar string.",
    1155: "State machine: inQuote flag and current accumulated string. On '\"': if closing, save current to result and reset; if opening, set flag. Otherwise append to current if inside quotes. Return all extracted strings.",
    1164: "For each column j (new row in rotated output), iterate original rows from n down to 1: grid[i][j] → new row element. This transpose-then-reverse-rows achieves 90° clockwise rotation.",
    1173: "Bottom-up DP. Base: dp[rows][cols] = max(1, 1-dungeon[rows][cols]). Fill last row right-to-left and last col bottom-to-top. Interior: best = min(dp[i][j+1], dp[i+1][j]) - dungeon[i][j], clamped to ≥1. Return dp[1][1].",
    1174: "While loop encoding runs: count consecutive equal elements from i with inner while. Add [count, val] pair (count first, value second). Decoder iterates pairs and repeats each val count times.",
    1178: "Global collObjects dict set via collision_registry. check_all iterates all pairs (i,j with j>i) from .keys() list. Compute Euclidean distance; report pair if dist < sum of radii.",
    1187: "Nested loops over all pairs i,j (j>i). Compute distance and minimum separation. If overlap exists and dist>0, compute normal vector, push each entity apart by overlap/2 along the normal.",
    1191: "Build 5×5 grid. For each cell (r,c), iterate rules: apply if rule's row and col match (or are -1 for wildcard). Last matching rule wins.",
    1195: "Project point onto segment: t = dot(p-a, b-a) / dot(b-a, b-a), clamped to [0,1]. Closest point = [ax+t*dx, ay+t*dy]. Return as a local list.",
    1220: "Iteratively remove zero-indegree and zero-outdegree nodes from remaining graph until no change. Nodes left form cycles; report all their edges as the feedback arc set.",
    1269: "State machine with inQuotes flag. Iterate characters: toggle on '\"'; split on ',' only when !inQuotes; otherwise append to current field. Add the last field after the loop ends.",
    1320: "while loop with seen dict for cycle detection. Convert x to string, split to digits, sum digit squares. Repeat until x==1 (happy) or a seen cycle is detected (unhappy). Return 1 or 0.",
    1382: "Initialize digitFreq dict with keys 0..9 each set to 0. Iterate piStr character by character with segment(i,1), incrementing the matching key's count.",
    1384: "Loop n from 10 to 99. Convert n to string, sum each digit segment as a number. If n%dsum==0, increment harshadCount. Print the final count.",
    1424: "Direct coordinate transformations: x-axis reflection negates y as [px, -py]; y-axis reflection negates x as [-px, py].",
    1445: "Iterate step vectors with for-each. Build direction label: dy<0 → append 'U', dy>0 → append 'D'; dx<0 → append 'L', dx>0 → append 'R'. Concatenate all labels for the full path string.",
    1452: "Recursive Welzl algorithm. Base cases: 0 boundary → trivial circle; 1 → point circle; 2 → diameter circle; 3 → circumcircle via determinant formula. Recursive case: try without last point; if it's outside the result, add it to boundary and recurse.",
    1470: "Nested loops 1..4 for rows and columns. mod(row+col, 2)==0 → white makeColor([255,255,255]); otherwise black makeColor([0,0,0]). Build and collect color rows.",
    1475: "Find the point in pts[2..n-1] with max perpendicular distance to the first-to-last line segment. If maxDist > epsilon, recursively simplify both sub-polylines and merge (skip duplicate at split point). Otherwise return just the two endpoints.",
    1489: "Check b using Falcon's type operator (b ? number) and b!=0 as compound condition. Return a/b if both true, else 0.",
    1553: "Compute mask = 2^width - 1. Left shift: n*2^bits. Wrap the shifted-out bits back: floor(n / 2^(width-bits)). OR them together and apply the mask to keep only width bits.",
    1793: "Recursive dict diff: for each key in v2 — if new, emit 'set'; if both dicts, recurse with path prefix; if value differs, emit 'set'. For each key in v1 not in v2, emit 'delete'. Accumulate all change operations.",
}


def replace_think(content: str, new_think: str) -> str:
    """Replace the content inside <think>...</think> in the assistant message."""
    return re.sub(
        r'<think>.*?</think>',
        f'<think>\n{new_think}\n</think>',
        content,
        flags=re.DOTALL
    )


def main():
    lines = MASTER_REASON.read_text(encoding="utf-8").splitlines(keepends=False)
    while lines and not lines[-1].strip():
        lines.pop()

    patched = 0
    for idx, think_text in THINKS.items():
        if idx >= len(lines):
            print(f"  SKIP idx {idx}: out of range")
            continue
        rec = json.loads(lines[idx])
        asst = next((m for m in rec["messages"] if m["role"] == "assistant"), None)
        if not asst:
            print(f"  SKIP idx {idx}: no assistant message")
            continue
        if "<think>" not in asst["content"]:
            print(f"  SKIP idx {idx}: no think block found")
            continue
        asst["content"] = replace_think(asst["content"], think_text)
        lines[idx] = json.dumps(rec, ensure_ascii=False)
        patched += 1

    MASTER_REASON.write_text("\n".join(lines) + "\n", encoding="utf-8")
    print(f"Patched {patched} think blocks.")


if __name__ == "__main__":
    main()
