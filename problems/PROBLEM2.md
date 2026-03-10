# Falcon Language — Problem Statements (1001–1200)

---

## Section 1: Variables (Problems 1001–1015)

1001. Declare a global `pi` set to `3.14159265`. Declare a local `tau` set to `this.pi * 2`. Print both.
1002. Declare globals `minVal` and `maxVal` both set to `0`. Write `observe(n)` that updates both — keeping the running minimum and maximum. Seed with `observe(5)` first, then call with `[3, 9, 1, 7, 2]`. Print both.
1003. Declare a local `matrix` as a list of three lists, each a row of three zeros. Print element at row 2, column 3.
1004. Declare globals `step` at `1` and `value` at `0`. Write `next()` that adds `this.step` to `this.value` and doubles `this.step`. Call 6 times. Print the sequence of values.
1005. Declare a local `flags` dict with keys `"verbose"`, `"dryRun"`, `"strict"` all set to `false`. Toggle `"verbose"` and `"strict"` to `true`. Print the dict.
1006. Declare a local `prev` and `curr` both set to `1`. Advance the pair 8 times using Fibonacci stepping (no function). Print each `curr`.
1007. Declare a global `registry` as an empty dict and a global `nextId` at `1`. Write `register(name)` that stores `name` with `this.nextId` as key, then increments. Register 4 names and print the registry.
1008. Declare a local `acc` as an empty list. Use a for loop to push squares of `1..10` into it, then assign `acc` to a new local `squares`. Print `squares`.
1009. Declare a global `dirty` at `false`. Write `markDirty()` and `markClean()` that flip it. Write `ifDirty(msg)` that prints `msg` only when dirty. Simulate.
1010. Declare locals `x0`, `y0`, `x1`, `y1` as two 2D points. Compute midpoint and distance. Print both.
1011. Declare a global `tick` at `0`. Write `advance(n)` that adds `n` to `this.tick`. Write `reset()` that zeroes it. Advance by `[3, 7, 2]`, print, reset, print.
1012. Declare a local `memo` as an empty dict. Without a function, manually compute and cache the first 6 Fibonacci values into `memo`. Print `memo`.
1013. Declare a local `chain` as a list of strings `["start"]`. Append each processed result: uppercase of previous item. Do this 5 times. Print `chain`.
1014. Declare globals `alpha` and `beta` each set to `2`. Write `swap()` that swaps them using XOR-style arithmetic (`alpha = alpha ~ beta` etc.). Call twice and print both times.
1015. Declare a local `history` as an empty list. In a for loop from `1` to `10`, push each value only if it is prime. Print `history`.

---

## Section 2: Math (Problems 1016–1045)

1016. Write `isAbundant(n)` — sum of proper divisors exceeds `n`. Collect all abundant numbers below `100` into a list and print.
1017. Write `aliquotSum(n)` — sum of proper divisors. Find all pairs `(a, b)` where `aliquotSum(a) == b` and `aliquotSum(b) == a` and `a != b` below `300` (amicable pairs).
1018. Write `digitfactorial(n)` — sum of factorials of each digit. Check whether `145` and `40585` equal their own digit-factorial sum (factorions).
1019. Write `kaprekarStep(n)` — arrange digits descending minus ascending. Iterate until reaching `6174` (Kaprekar constant). Count steps for `n=1234`.
1020. Compute `pi` via the Monte Carlo method: generate `10000` random `(x, y)` pairs in `[0,1]` and count how many fall inside the unit circle. Print the approximation.
1021. Write `isSmith(n)` — a composite whose digit sum equals the sum of digits of its prime factors. Test `4`, `22`, `27`, `121`.
1022. Write `abundantSumPairs(limit)` — find all integers below `limit` that can be expressed as the sum of two abundant numbers. Count them for `limit=100`.
1023. Write `narcissistic(n)` — Armstrong number generalised to `k` digits. Collect all 4-digit narcissistic numbers.
1024. Compute the sum `1/1^2 + 1/2^2 + ... + 1/n^2` for `n=1000`. Compare to `pi^2/6` and print the error.
1025. Write `continued_fraction_sqrt2(terms)` — approximate `sqrt(2)` using a continued-fraction expansion for `terms` iterations. Print for `terms=20`.
1026. Write `jacobiSymbol(a, n)` — simplified Jacobi symbol computation for odd `n`. Test `(2, 15)`.
1027. Write `totientSum(n)` — sum of `euler_phi(k)` for `k` from `1` to `n`. Print for `n=20`.
1028. Write `mobiusFn(n)` — return `0` if `n` has a squared prime factor, `1` if even number of prime factors, `-1` if odd. Test `1..10`.
1029. Compute the number of carry operations when adding `123` and `456` digit by digit. Generalise for two arbitrary numbers.
1030. Write `digitalPersistence(n)` — repeatedly multiply digits until single digit. Count steps. Test `77` (persistence 4) and `679` (persistence 5).
1031. Write `collatzMax(n)` — the maximum value reached in the Collatz sequence starting at `n`. Test `n=27`.
1032. Write `isAutomorphic(n)` — `n^2` ends in the digits of `n`. Test `5`, `6`, `25`, `76`.
1033. Write `sumOfFactors(n)` and use it to find all perfect numbers below `10000`.
1034. Write `primorial(n)` — product of all primes up to `n`. Test `n=13`.
1035. Write `sopfr(n)` — sum of prime factors with repetition. Test `12` (2+2+3=7), `60`.
1036. Write `goldbach(n)` — express even `n` as sum of two primes (Goldbach). Print pairs for `n=28, 36, 100`.
1037. Write `harshad(n)` — divisible by its digit sum. List all Harshad numbers below `100`.
1038. Write `odious(n)` — odd number of `1` bits. Collect first 10 odious numbers.
1039. Write `evil(n)` — even number of `1` bits. Collect first 10 evil numbers.
1040. Write `polygonal(n, s)` — nth s-gonal number. Print first 8 hexagonal, heptagonal, and octagonal numbers.
1041. Write `isEmirp(n)` — prime whose digit-reversal is a different prime. Test `13`, `17`, `31`, `11`.
1042. Compute the Stern-Brocot sequence for the first 30 terms.
1043. Write `intToBijective(n)` — bijective base-10 (digits 1–9 only, no zero). Test `9`, `10`, `19`, `100`.
1044. Write `selfDescribing(n)` — a number where digit `i` describes how many times `i` appears. Test `6210001000` (10-digit).
1045. Compute the first 15 terms of the look-and-say sequence starting from `"1"`.

---

## Section 3: Text (Problems 1046–1075)

1046. Write `longestCommonSubstring(a, b)` — not subsequence, must be contiguous. Test `("abcdef", "zcdemf")`.
1047. Write `smallestWindow(s, t)` — shortest substring of `s` containing all characters of `t`. Test `("ADOBECODEBANC", "ABC")`.
1048. Write `countDistinctSubstrings(s)` — count unique substrings. Test `"abab"`.
1049. Write `zigzagEncode(s, rows)` — encode string in zigzag pattern across rows and read off row by row. Test `("PAYPALISHIRING", 3)`.
1050. Write `atbashCipher(s)` — substitute each letter with its mirror in the alphabet (a→z, b→y). Test `"Hello"`.
1051. Write `vigenereEncrypt(text, key)` — Vigenère cipher. Test `("ATTACKATDAWN", "LEMON")`.
1052. Write `runLengthDecompress(s)` — expand `"3a2b1c"` to `"aaabbc"`. Test.
1053. Write `longestWordNoRepeats(sentence)` — word with no repeated characters. Test.
1054. Write `reverseVowels(s)` — swap only the vowel characters, keeping consonants in place. Test `"hello world"`.
1055. Write `removeConsecutiveDuplicates(s)` — remove consecutive duplicate characters. Test `"aabbccddee"`.
1056. Write `groupAnagrams(words)` — group list of words into anagram families (dict of sorted-key → list). Test `["eat","tea","tan","ate","nat","bat"]`.
1057. Write `isInterleaving(a, b, c)` — check if `c` is an interleaving of `a` and `b`. Test `("aab", "axy", "aaxaby")`.
1058. Write `longestNonRepeatSubstring(s)` — length of longest substring with no repeating characters. Test `"abcabcbb"`.
1059. Write `wordWrap(text, width)` — greedy word wrap into lines. Test `("the quick brown fox", 10)`.
1060. Write `decodeRunLength(s)` — pairs of `(digit, char)`. Test `"2A3B1C"`.
1061. Write `isomorphic(a, b)` — each character in `a` maps consistently to one in `b`. Test `("egg","add")`, `("foo","bar")`.
1062. Write `textJustify(words, width)` — distribute spaces evenly between words to fill exact `width`. Test.
1063. Write `countBracketDepth(s)` — maximum nesting depth of `(`. Test `"((a)(b(c)))"`
1064. Write `extractEmails(text)` — find all `word@word.word` patterns. Test a paragraph.
1065. Write `camelToWords(s)` — convert `"camelCaseString"` to `"camel case string"`. Test.
1066. Write `wordsToSentence(list)` — join list, capitalize first, append period. Test.
1067. Write `toOrdinal(n)` — format number as ordinal string (`1st`, `2nd`, `3rd`, `4th`...). Test `1..20`.
1068. Write `expandContractions(s)` using `replaceFrom` with a contraction dict. Test `"I can't don't won't"`.
1069. Write `obfuscate(s, n)` — shift each character forward by `n` positions in a circular alphabet. Test `("Hello", 3)`.
1070. Write `stripAccents(s)` — replace accented characters using `replaceFrom` dict. Test `"café naïve résumé"`.
1071. Write `splitCamelCase(s)` — insert space before each uppercase letter. Test `"camelCaseString"`.
1072. Write `toCurrencyString(amount)` — format `1234567.89` as `"1,234,567.89"`. Test.
1073. Write `removeDiacritics(s)` — remove combining-mark-like characters (simplified). Test.
1074. Write `detectLanguage(s)` — count frequency of letters and guess English vs other based on `e` frequency. Test.
1075. Write `readabilityScore(text)` — count words, sentences (ends in `.`), and syllables. Compute Flesch score approximation.

---

## Section 4: Lists (Problems 1076–1110)

1076. Write `slidingWindowMax(list, k)` — max in each sliding window of size `k`. Test `([1,3,-1,-3,5,3,6,7], 3)`.
1077. Write `longestBitonicSubarray(list)` — length of longest subarray that first increases then decreases. Test.
1078. Write `trapWater(heights)` — amount of water trapped. Test `[4,2,0,3,2,5]`.
1079. Write `stockBuySell(prices)` — maximum profit from one buy and one sell. Test `[7,1,5,3,6,4]`.
1080. Write `stockBuySellMultiple(prices)` — maximum profit with unlimited transactions. Test.
1081. Write `zigzagSort(list)` — arrange so `a < b > c < d`. Test `[4,3,1,2]`.
1082. Write `waveSort(list)` — same as zigzag but using sort-then-swap approach. Test.
1083. Write `countInversions(list)` — pairs `(i,j)` where `i<j` but `list[i]>list[j]`. Test `[2,4,1,3,5]`.
1084. Write `medianOfMedians(list)` — simplified: find median of list using sort. Test.
1085. Write `threeWayPartition(list, low, high)` — Dutch National Flag. Test `([1,2,0,1,2,0,2,1], 0, 2)`.
1086. Write `longestSubarrayWithSum(list, target)` — longest contiguous subarray summing to `target`. Test `([1,2,3,2,1,3], 6)`.
1087. Write `subarrayCount(list, target)` — number of subarrays summing to `target`. Test.
1088. Write `productArray(list)` — each element is product of all others without division. Test.
1089. Write `circularArrayMax(list, k)` — max sum of `k` consecutive elements in a circular array. Test.
1090. Write `mergekSortedLists(lists)` — merge `k` sorted lists into one sorted list. Test 3 lists.
1091. Write `findKthSmallest(list, k)` using sort. Test.
1092. Write `findKthLargest(list, k)`.
1093. Write `groupByModulo(list, m)` — group numbers by their remainder mod `m`. Test `([1..15], 4)`.
1094. Write `longestSublistWithUniqueElems(list)` — length of longest contiguous section with all unique elements. Test.
1095. Write `allSubarrays(list)` — return all contiguous subarrays. Test `[1,2,3]`.
1096. Write `prefixSums(list)` — list where `result[i] = sum(list[1..i])`. Test.
1097. Write `suffixSums(list)` — list where `result[i] = sum(list[i..end])`. Test.
1098. Write `rangeSum(prefixSums, l, r)` — sum of elements from `l` to `r` using prefix sums. Test.
1099. Write `longestFlatSubarray(list)` — longest run with the same value. Test `[1,1,1,2,2,3,3,3,3]`.
1100. Write `matrixTrace(m)` — sum of diagonal of square matrix. Test 4x4.
1101. Write `rotateListRight(list, k)`. Test `([1,2,3,4,5], 2)`.
1102. Write `interweave(a, b, c)` — interleave three lists. Test.
1103. Write `splitEvens(list)` — separate into two lists: even indices and odd indices. Test.
1104. Write `buildPyramid(list)` — each level is pairwise sums of previous. Return all levels. Test `[1,2,3,4]`.
1105. Write `chainMap(list, fns)` — apply a sequence of named transforms in order. Test `([1,2,3], ["double","square"])`.
1106. Write `listDiff(a, b)` — element-wise differences. Test.
1107. Write `listRatio(a, b)` — element-wise ratios. Test.
1108. Write `listMax(a, b)` — element-wise maximum. Test.
1109. Write `listMin(a, b)` — element-wise minimum. Test.
1110. Write `rollingStdDev(list, k)` — standard deviation of each window of size `k`. Test.

---

## Section 5: Dictionaries (Problems 1111–1135)

1111. Write `mapInPlace(d, fn)` — apply string-named function to every value. Test doubling numeric values.
1112. Write `accumDict(list, keyFn, valFn)` — build dict by accumulating values per key. Test grouping word lengths by first letter.
1113. Write `countByRange(d, lo, hi)` — count values within `[lo, hi]`. Test.
1114. Write `invertMultimap(d)` — dict maps key → list; invert to value → list of keys. Test.
1115. Write `diffKeys(a, b)` — keys in `a` not in `b`. Test.
1116. Write `commonKeys(a, b)` — keys in both dicts. Test.
1117. Write `mergeWithConflict(a, b, strategy)` — on conflict, pick `"first"`, `"second"`, or `"sum"`. Test all three.
1118. Write `transformKeyNames(d, map)` — rename keys according to a mapping dict. Test `{"fname":"name", "lname":"surname"}`.
1119. Write `nestedGet(d, keys)` — get nested value from list of keys. Test missing path returning `""`.
1120. Write `histogramToFreq(hist, total)` — convert count dict to percentage dict. Test.
1121. Write `rankByValue(d)` — return dict mapping key → rank (1 = highest value). Test.
1122. Write `clusterByValue(d, n)` — group keys into `n` buckets by value range. Test.
1123. Write `dictToCSV(headers, rows)` — list of dicts to CSV string. Test 3 rows.
1124. Write `csvToDict(headers, csv)` — parse CSV string into list of dicts. Test.
1125. Write `buildAdjList(edges)` — from list of `[from, to]` pairs, build adjacency dict. Test.
1126. Write `dfsDict(graph, start)` — DFS traversal using adjacency dict. Test.
1127. Write `bfsDict(graph, start)` — BFS traversal. Test.
1128. Write `shortestPathDict(graph, start, end)` — BFS-based shortest path. Test.
1129. Write `detectCycleDict(graph)` — detect if directed graph has a cycle. Test.
1130. Write `inDegree(graph)` — count incoming edges for each node. Test.
1131. Write `outDegree(graph)` — count outgoing edges. Test.
1132. Write `topSortDict(graph)` — topological sort using Kahn's algorithm on adjacency dict. Test.
1133. Write `SCCKosaraju(graph)` — find strongly connected components using two-pass DFS. Test.
1134. Write `invertGraph(graph)` — reverse all edges in adjacency dict. Test.
1135. Write `isDAG(graph)` — true if directed graph has no cycle. Test.

---

## Section 6: Colors (Problems 1136–1150)

1136. Write `colorFromWavelength(nm)` — map wavelength (380–750 nm) to approximate RGB. Test `450`, `550`, `650`.
1137. Write `heatmap(value, lo, hi)` — return a color from blue (cold) to red (hot). Test `[lo, mid, hi]`.
1138. Write `neonify(color)` — boost saturation by pushing channels toward their extreme. Test.
1139. Write `splitTone(shadows, highlights, color)` — blend shadow color into darks and highlight color into lights. Test.
1140. Write `colorDistance3D(a, b)` — Euclidean distance in RGBA space using 4 components. Test.
1141. Write `colorScheme(base, scheme)` — return list of colors for `"complementary"`, `"analogous"`, `"triadic"`. Test each.
1142. Write `buildGradientTable(colorA, colorB, n)` — list of `n` evenly blended colors. Print as RGB triples. Test.
1143. Write `colorToHex(color)` — convert using `decToHex` for each channel. Pad to 2 digits. Test.
1144. Write `hexToColor(hex)` — parse 6-char hex to `makeColor`. Test `"FF8000"`.
1145. Write `dominantHue(palette)` — return the color in the list that is most saturated (max channel minus min channel). Test.
1146. Write `colorQuantize(palette, n)` — reduce to `n` representative colors by bucketing. Test.
1147. Write `applyLUT(color, lut)` — look up each channel in a 256-entry list. Test with simple gamma LUT.
1148. Write `colorDelta(a, b)` — return new color representing absolute channel-wise difference. Test.
1149. Write `colorAverage(palette)` — channel-wise mean of all colors. Test 4 colors.
1150. Write `makeRainbow(n)` — list of `n` colors cycling through hues using channel rotation math. Test `n=7`.

---

## Section 7: Controls (Problems 1151–1175)

1151. Use a while loop to find the smallest `n` such that `sum of 1/k` for `k=1..n` exceeds `3`.
1152. Use nested for loops to print all prime pairs `(p, p+2)` (twin primes) below `100`.
1153. Use a while loop to compute integer square root without `sqrt` — repeated subtraction of odd numbers.
1154. Write a loop that reads a list `[3, -1, 4, -1, 5, -9, 2, 6]` and prints the index of the first negative followed by the index of the last negative.
1155. Use nested loops to fill a 5×5 grid with numbers where `grid[i][j] = i * j`. Print each row.
1156. Simulate a simple queue processor: enqueue 8 tasks (strings), dequeue and `println` one at a time in a while loop.
1157. Use a for loop to find the two closest numbers in a sorted list `[1, 5, 9, 14, 20, 27]`.
1158. Write a loop that collapses consecutive equal elements: `[1,1,2,2,2,3,1,1]` → `[1,2,3,1]`. Print result.
1159. Use a for-each loop to print each key-value pair from a dict only if the value is a number.
1160. Write a simulation loop: `n` frogs on a log, each second one random frog jumps off. Loop until empty. Print how many seconds.
1161. Use a while loop to compute Babylonian square root of `50` — `x = (x + 50/x)/2` starting from `x=25`. Stop when stable to 6 decimals.
1162. Use nested loops to compute the number of lattice points inside a circle of radius `10`.
1163. Simulate a rain gauge: a list of hourly readings. Use a loop to find the longest dry spell (consecutive zeros). Test.
1164. Use a for loop to build a spiral sequence: alternately append from front and back of list `[1..20]`.
1165. Write a loop that counts how many numbers from `1` to `1000` are simultaneously divisible by `3`, `7`, but not by `5`.
1166. Use nested loops to print a right-aligned triangle of numbers — row `i` shows numbers `1` to `i`.
1167. Simulate a simple calculator loop: process a list of `[operand, operator]` pairs applied sequentially to an accumulator. Test `[(5,"+"), (3,"*"), (2,"-")]`.
1168. Write a for-each loop over a nested list, printing the index and sum of each inner list.
1169. Use a while loop to generate the Padovan sequence (each term = sum of two terms before the previous). Print first 15.
1170. Use nested for loops to list all ways to make change for `15` cents using coins `[1, 5, 10]`.
1171. Write a loop to detect whether a list forms a palindrome. Print result and do not use `reverseList`.
1172. Simulate a school bell: every 45 minutes a bell rings. Loop through a 480-minute day, printing bell times.
1173. Use a for range loop to sum all integers from `1` to `n` that share no common factor with `n` (`gcd == 1`). Test `n=30`.
1174. Write a loop that finds the mode of a list without a dict — using a nested count loop. Test.
1175. Use a for-each loop over a list of strings and collect only those that are palindromes. Print the result list.

---

## Section 8: Procedures (Problems 1176–1200)

1176. Write `memoize(cache, key, computation)` where computation is expressed as a string-dispatched function call. Test `fib` with cache.
1177. Write `retry(maxAttempts, successThreshold)` — loop calling `randInt(1,10)` until it exceeds threshold or max attempts. Return success or failure.
1178. Write `parseAndEval(expr)` — parse `"a + b"` style expressions from a dict of variables. Test `{"a":3,"b":5}` with `"a + b"`.
1179. Write `buildFSM(transitions, initial)` — dict-based FSM. Write `runFSM(fsm, inputs)` that transitions through states. Test a simple traffic light FSM.
1180. Write `coroutineSim(steps)` — simulate cooperative multitasking with a list of `[name, action]` steps. Print each task's turn.
1181. Write `eventLoop(events, handlers)` — dispatch list of events to handler functions by name. Test 5 events.
1182. Write `pipeline(value, stages)` — each stage is a string name of a transform. Apply sequentially. Return result. Test `(10, ["double","increment","square"])`.
1183. Write `optionalChain(d, path)` — safely navigate nested dict, return `""` at any missing step. Test.
1184. Write `matchPattern(value, cases)` — simplified pattern matching: list of `[pred, result]` pairs, return first matching result. Test.
1185. Write `generate(seed, fn, n)` — produce a list of `n` values where each is derived from the previous by named function dispatch. Test.
1186. Write `scan(list, init, fn)` — like reduce but collects all intermediate accumulators into a list. Test with sum.
1187. Write `slidingReduce(list, k, fn)` — apply reduce to each window of size `k`. Test sum-of-window.
1188. Write `deepFlatten(list)` recursively. Test `[1,[2,[3,[4,5]]],6]`.
1189. Write `treeMap(tree, fn)` — apply string-named function to every `val` in a nested `{val, left, right}` dict tree. Test.
1190. Write `treeFold(tree, init, fn)` — fold over all node values. Test computing sum.
1191. Write `buildHeap(list)` — simulate a min-heap using a list. Write `heapPush(heap, val)` and `heapPop(heap)`. Test.
1192. Write `huffmanCodes(freq)` — given letter frequency dict, build greedy Huffman-like codes as a dict. Test `{"a":5,"b":2,"c":1,"d":3}`.
1193. Write `topologicalLevels(graph)` — return list of levels (nodes with same depth from sources). Test.
1194. Write `bellmanFord(graph, start)` — shortest paths on weighted graph with negative edges. Test.
1195. Write `floydWarshall(matrix)` — all-pairs shortest path on adjacency matrix. Test 4-node graph.
1196. Write `kruskal(edges, nodes)` — minimum spanning tree using union-find. Test.
1197. Write `prim(graph, start)` — MST using greedy adjacency approach. Test.
1198. Write `aStarSearch(grid, start, end)` — simplified A* on a 2D grid using Manhattan heuristic. Test 5×5 grid.
1199. Write `simulatedAnnealing(fn, start, temp, cooling)` — optimize a single-variable function by random perturbation. Test minimizing a quadratic.
1200. Write `geneticStep(population, fitnessFn, mutationRate)` — one generation of selection, crossover, and mutation on lists of bits. Test on a simple target bit string.
