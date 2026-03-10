# Falcon Language — 1000 Problem Statements

---

## Section 1: Variables (Problems 1–80)

1. Declare a global variable `score` initialized to `0`, then print it using `this.score`.
2. Declare a local variable `name` set to `"Falcon"` and print it.
3. Declare two global variables `x` and `y` set to `10` and `20`. Print their sum using `this.x + this.y`.
4. Declare a local variable `count` initialized to `5`, add `3` to it, then print the result.
5. Declare a global variable `message` set to `"Hello"`. Reassign it to `"World"` and print it.
6. Declare a local boolean variable `isReady` set to `true` and print it.
7. Declare two local variables `a` and `b`, swap their values using a temp variable, and print both.
8. Declare a global variable `total` starting at `0`. Write a void function `addToTotal(n)` that increases `this.total` by `n`. Call it three times and print `this.total`.
9. Declare a local variable `pi` set to `3.14159`. Multiply it by `2` and print the result.
10. Declare a local variable `greeting` set to `"Hello"`. Concatenate `", World!"` to it and print.
11. Declare a global `counter` at `0`. Write a function `increment()` that adds `1` to it. Call it `5` times and print.
12. Declare a local variable `result` set to `100`. Subtract `37` from it. Print the final value.
13. Declare two globals `firstName` and `lastName`. Print the full name joined by a space.
14. Declare a local variable `n` set to `7`. Print whether it is greater than `5` using an if expression.
15. Declare a global list `items` initialized to an empty list `[]`. Add three elements and print it.
16. Declare a local variable `flag` set to `false`. Toggle it to `true` with `!flag` and print.
17. Declare a local `temperature` set to `37.5`. Print whether it represents a fever (above `37.2`).
18. Declare a global `balance` set to `1000`. Write a `withdraw(amount)` function that subtracts from `this.balance` only if sufficient funds exist, otherwise prints `"Insufficient funds"`.
19. Declare three local variables `r`, `g`, `b` set to `255`, `128`, `0`. Print each on its own line.
20. Declare a local variable `word` set to `"racecar"`. Print both the word and its length.
21. Declare a global `callCount` at `0`. Write any function that increments it on every call. Call it `10` times and print.
22. Declare a local variable `x` set to `-15`. Print its absolute value without using `abs()` — use an if expression instead.
23. Declare a local `speed` set to `60`. Multiply it by `1.60934` to convert mph to km/h and print.
24. Declare two locals `p` and `q` both set to `true`. Print the result of `p && q`, `p || q`, and `!p`.
25. Declare a global `history` as an empty list. Write a `record(val)` function that appends `val` to `this.history`. Call it 5 times and print.
26. Declare a local `age` set to `17`. Print `"Minor"` if under 18, else `"Adult"`.
27. Declare a local `discount` set to `0.15`. Compute the discounted price of an item costing `80` and print it.
28. Declare a global `maxSeen` set to `0`. Write `observe(n)` that updates `this.maxSeen` if `n` is larger. Test with values `3, 7, 2, 9, 4` and print final max.
29. Declare a local `binary` set to `"1101"`. Print the statement `binary _ " is a binary string: " _ (binary ? bin)`.
30. Declare a local list `primes` set to `[2, 3, 5, 7, 11]`. Reassign the third element to `99` and print the list.
31. Declare a global `isLoggedIn` set to `false` and `username` set to `""`. Write `login(name)` that sets both. Call it and print.
32. Declare a local `x` set to `3.7`. Print the floor, ceiling, and rounded values.
33. Declare a local `sentence` set to `"The quick brown fox"`. Print the number of characters.
34. Declare a global `attempts` at `0` and `maxAttempts` at `3`. Write `tryAction()` that increments attempts and prints `"Blocked"` if over max.
35. Declare a local `ratio` as `7 / 4` and print it formatted to 2 decimal places.
36. Declare a local `hexVal` set to `"1F"`. Print `hexToDec(hexVal)`.
37. Declare locals `start` and `end` set to `1` and `100`. Print their average.
38. Declare a global `version` set to `"1.0.0"`. Print whether it starts with `"1"`.
39. Declare a local `items` list with 5 numbers. Print the sum using `reduce`.
40. Declare a local `label` set to `"  Falcon  "`. Trim it and print.
41. Declare a global `config` as a dictionary `{"debug": true, "version": "2.0"}`. Print the version.
42. Declare a local `val` and assign the result of `if (10 > 5) "yes" else "no"` to it. Print.
43. Declare a global `log` as an empty list and a `logEntry(msg)` function. Log 4 messages and print the log.
44. Declare a local `x` set to `256`. Print `decToBin(x)` and `decToHex(x)`.
45. Declare a local `words` list set to `["apple", "banana", "cherry"]`. Reassign index 2 and print.
46. Declare a global `isRunning` set to `true`. Write a `stop()` function that sets it to `false`. Call it and print.
47. Declare a local `n` set to `0`. Increment it inside a while loop 10 times. Print `n`.
48. Declare a global `buffer` as `""`. Write `append(s)` that concatenates to the buffer. Append 4 strings and print.
49. Declare a local `angle` set to `45`. Compute `sin(radians(angle))` and print formatted to 4 decimal places.
50. Declare locals `a`, `b`, `c` as `3`, `4`, `0`. Assign `c = a * b` and print all three.
51. Declare a global `streak` at `0` and `bestStreak` at `0`. Write `win()` and `lose()` functions. Simulate a sequence and print best streak.
52. Declare a local variable `n` set to `1000`. Divide it by `3` using integer division (use `floor`) and print quotient and remainder.
53. Declare a local `names` list. Add five names via `add()`. Print the list sorted.
54. Declare a global `palette` as an empty list. Write `addColor(c)` that pushes colors. Add 3 colors and print.
55. Declare a local `text` set to `"12345"`. Print whether it is a number type using `?`.
56. Declare a local `x` set to `2`. Print `x`, `x^2`, `x^3`, `x^4`, `x^5` each on a new line.
57. Declare a global `map` as empty dict. Write `put(k, v)` that sets a key. Add 3 pairs and print.
58. Declare a local `threshold` set to `50`. Print `"Pass"` or `"Fail"` for each of `[45, 60, 50, 33, 78]`.
59. Declare a local `running` set to `true`. Simulate one loop iteration that sets it to `false` and prints `"Stopped"`.
60. Declare globals `wins` `losses` `draws` all at `0`. Write `result(r)` that increments the right counter. Simulate 6 results and print all three.
61. Declare a local `n` set to `5`. Print the sum of integers from `1` to `n` using the formula `n*(n+1)/2`.
62. Declare a local `phrase` set to `"Hello World"`. Print it in uppercase and lowercase.
63. Declare a global `maxCapacity` set to `10` and `current` at `0`. Write `addItem()` that increments only if under capacity.
64. Declare a local `data` set to `"name:Alice:age:30"`. Split on `":"` and print each field.
65. Declare a global `cache` as an empty dict. Write `store(key, val)` and `fetch(key)` functions.
66. Declare a local `price` set to `9.99`. Multiply by `1.08` for tax and print formatted to 2 places.
67. Declare a local `bits` set to `"10110101"`. Print `binToDec(bits)`.
68. Declare a global `sessionTokens` as an empty list. Write `addToken(t)` and `revokeToken(t)` functions. Simulate and print.
69. Declare a local `x` set to `10`. Print the result of `x ~ 6` (XOR) and `x & 3` (AND) and `x | 5` (OR).
70. Declare locals `monday` through `friday` each set to hours worked (e.g. `8, 7, 9, 6, 8`). Print total weekly hours.
71. Declare a local `flag` using `5 > 3 && 2 < 4`. Print the flag.
72. Declare a global `errors` list. Write `reportError(msg)` that appends and prints `"Error #" _ errors.listLen() _ ": " _ msg`.
73. Declare a local `x` set to `3.999`. Print the result of casting to integer via `floor(x)`.
74. Declare a local `pair` as a list `["Alice", 95]`. Print `"Name: " _ pair[1] _ ", Score: " _ pair[2]`.
75. Declare a global `settings` dict with keys `"volume"`, `"brightness"`, `"theme"`. Write `updateSetting(k, v)` and call it twice.
76. Declare a local `n` set to `25`. Print `sqrt(n)` and check whether `n` is a perfect square.
77. Declare a local `sentence` set to `"one two three four five"`. Split at spaces and print the 3rd word.
78. Declare a global `eventQueue` as an empty list. Write `enqueue(e)` and `dequeue()` that removes the first item. Simulate and print.
79. Declare locals `lat` and `lon` set to `40.7128` and `-74.0060`. Print `"Location: " _ lat _ ", " _ lon`.
80. Declare a global `initialized` set to `false`. Write `init()` that sets it to `true` and prints `"Ready"` only if not already initialized.

---

## Section 2: Math (Problems 81–230)

81. Print the result of `17 + 28 * 3 - 6 / 2`.
82. Compute and print the area of a circle with radius `7` using `pi = 3.14159`.
83. Compute and print the hypotenuse of a right triangle with legs `3` and `4`.
84. Print `sqrt(2)` formatted to 6 decimal places.
85. Print the result of `2 ^ 32`.
86. Compute and print `floor(7 / 2)` and `ceil(7 / 2)`.
87. Compute `log(1000)` and print it formatted to 4 decimal places.
88. Print the sine, cosine, and tangent of `60` degrees.
89. Compute `exp(2)` and print it formatted to 4 decimal places.
90. Print the decimal value of binary string `"11001010"`.
91. Print the binary string of decimal `200`.
92. Print the hexadecimal string of `4095`.
93. Print the decimal value of hex `"DEAD"`.
94. Generate and print 5 random integers between `1` and `100`.
95. Print a random float. If it is above `0.75`, print `"Lucky"`, else `"Try again"`.
96. Use `setRandSeed(99)` then print three consecutive `randInt(1, 10)` results.
97. Compute `mod(123, 7)`, `rem(123, 7)`, and `quot(123, 7)` and print all three.
98. Compute the average of `[55, 82, 91, 73, 67, 88, 76]` using `avgOf`.
99. Print `stdDevOf([4, 8, 15, 16, 23, 42])`.
100. Print `geoMeanOf([1, 2, 4, 8, 16])`.
101. Find and print the mode of `[3, 1, 4, 1, 5, 9, 2, 6, 1, 3]`.
102. Print `min(7, 3, 9, 1, 5)` and `max(7, 3, 9, 1, 5)`.
103. Format `3.14159265358979` to 2, 4, and 8 decimal places and print each.
104. Compute the area of a rectangle with width `13.5` and height `7.2` and print.
105. Compute compound interest: principal `1000`, rate `5%`, `10` years. Print the final amount.
106. Write a function `hypotenuse(a, b) = { sqrt(a^2 + b^2) }`. Test with `(5, 12)` and `(8, 15)`.
107. Write a function `celsiusToKelvin(c) = { c + 273.15 }`. Test with `0`, `100`, `-273.15`.
108. Write a function `circleArea(r) = { 3.14159 * r ^ 2 }`. Test with radii `1`, `5`, `10`.
109. Write a function `sphereVolume(r) = { (4/3) * 3.14159 * r ^ 3 }`. Test with `r = 3`.
110. Write a function `triangleArea(base, height) = { 0.5 * base * height }`. Test with `(10, 6)`.
111. Print `degrees(asin(0.5))` — the angle whose sine is 0.5.
112. Print `degrees(acos(0.5))`.
113. Print `degrees(atan(1))`.
114. Compute the distance between points `(3, 4)` and `(0, 0)` using `sqrt`.
115. Compute the distance between `(1, 2)` and `(4, 6)` and print.
116. Write `isPerfectSquare(n)` using `floor` and `sqrt`. Test `16`, `20`, `81`, `100`.
117. Write `isPerfectCube(n)` using `round` and `^`. Test `8`, `27`, `30`, `64`.
118. Print the sum of squares from `1` to `10` using a for loop.
119. Print the sum of cubes from `1` to `5`.
120. Print all multiples of `7` from `1` to `100`.
121. Compute the value of the arithmetic series `sum of i from 1 to 100` using the formula `n*(n+1)/2`.
122. Compute `n!` for `n = 12` using a while loop.
123. Write `nCr(n, r)` using factorials. Print `nCr(10, 3)`.
124. Write `nPr(n, r) = { nCr(n,r) * factorial(r) }`. Print `nPr(5, 2)`.
125. Print the first 15 triangular numbers `T(n) = n*(n+1)/2`.
126. Print the first 10 pentagonal numbers `P(n) = n*(3n-1)/2`.
127. Print the first 10 square numbers.
128. Print the sum of all even numbers from `2` to `50`.
129. Print the sum of all odd numbers from `1` to `99`.
130. Write `digitSum(n)` using modulo and floor. Test `9875`, `12345`, `999`.
131. Write `reverseDigits(n)` using modulo arithmetic. Test `1234`, `9870`.
132. Write `isPalindromicNumber(n)` by reversing digits. Test `121`, `123`, `1331`.
133. Write `countDivisors(n)`. Print how many divisors `60` has.
134. Write `sumDivisors(n)` that sums all divisors. Print for `28` (perfect number).
135. Write `isPerfectNumber(n)`. Test `6`, `28`, `12`.
136. Write `isAbundant(n)` where sum of proper divisors > n. Test `12`, `15`.
137. Write `collatzLength(n)` that counts steps to reach 1. Test `27`, `6`, `1`.
138. Print the Collatz sequence starting from `19`.
139. Write `digitalRoot(n)` — repeatedly sum digits until single digit. Test `493`, `942`.
140. Write `toBinary(n)` using repeated division and string concatenation. Test `10`, `255`.
141. Write `toHex(n)` using `decToHex`. Test `255`, `4096`, `65535`.
142. Compute `sin^2(x) + cos^2(x)` for `x = 37` degrees and confirm it equals `1`.
143. Compute the surface area of a cylinder: `2*pi*r*(r+h)` for `r=4`, `h=10`.
144. Compute the volume of a cone: `(1/3)*pi*r^2*h` for `r=3`, `h=9`.
145. Write `harmonicSum(n)` = sum of `1/i` for `i` from 1 to n. Print for `n=10`.
146. Write `geometricSum(a, r, n)` = `a * (r^n - 1) / (r - 1)`. Test `(2, 3, 5)`.
147. Print whether `abs(-5) + abs(-3) == abs(-5 + -3)` is always true (test a few cases).
148. Compute and print the golden ratio using `(1 + sqrt(5)) / 2`.
149. Write `clamp(x, lo, hi)` that restricts `x` to the range. Test `(-5, 0, 10)`, `(15, 0, 10)`, `(5, 0, 10)`.
150. Write `lerp(a, b, t)` = `a + (b - a) * t`. Print values for `t` in `[0, 0.25, 0.5, 0.75, 1]`.
151. Write `normalize(x, min, max)` = `(x - min) / (max - min)`. Test with `(75, 0, 100)`.
152. Write `map(x, inMin, inMax, outMin, outMax)` for range remapping. Convert `50` from `[0,100]` to `[0,255]`.
153. Compute `atan2(1, 1)` in degrees and print.
154. Print the first 20 Fibonacci numbers using a loop.
155. Write `tribonacci(n)` — each term is sum of previous 3. Print first 10 terms.
156. Write `lucas(n)` — Lucas sequence starting `2, 1`. Print first 10.
157. Write `catalan(n) = (2n)! / ((n+1)! * n!)`. Print `catalan(5)`.
158. Write `powerMod(base, exp, mod)` using iterative squaring simulation. Test `(2, 10, 1000)`.
159. Compute the sum of the first `n` squares: `n*(n+1)*(2n+1)/6`. Test `n=10`.
160. Compute the sum of the first `n` cubes: `(n*(n+1)/2)^2`. Test `n=5`.
161. Write `countPrimes(n)` that counts all primes up to `n`. Print for `n=100`.
162. Write `nthPrime(n)` that returns the `n`th prime number. Print 10th and 20th primes.
163. Compute `pi` approximation using Leibniz formula for 1000 terms.
164. Write `isPrime(n)` and use it to collect all primes between `100` and `200` into a list.
165. Write `primeFactors(n)` that returns a list of prime factors. Test `60`, `84`, `360`.
166. Write `gcd(a, b)` recursively and `lcm(a, b)` using it. Test `(48, 18)`, `(100, 75)`.
167. Print the first 5 pairs of twin primes (primes differing by 2).
168. Compute Euler's number `e` as `sum of 1/k!` for `k` from 0 to 20.
169. Write `sigmoid(x) = 1 / (1 + exp(-x))`. Print for `x` in `[-2, -1, 0, 1, 2]`.
170. Write `relu(x) = max(0, x)` equivalent using if expression. Test `[-3, 0, 2, 5]`.
171. Write `softmax(list)` that normalizes a list to sum to 1. Test `[1.0, 2.0, 3.0]`.
172. Compute BMI from weight `70` kg and height `1.75` m. Print category.
173. Compute compound interest with monthly compounding: `P*(1 + r/12)^(12*t)`. Use `P=5000`, `r=0.04`, `t=5`.
174. Write `degreesToRadians(d)` and `radiansToDegrees(r)`. Confirm round-trip.
175. Compute the `n`th term of an arithmetic sequence: `a + (n-1)*d`. Test `a=3`, `d=7`, `n=10`.
176. Compute the `n`th term of a geometric sequence: `a * r^(n-1)`. Test `a=2`, `r=3`, `n=8`.
177. Write `sumArithmetic(a, d, n)` and print the sum of first 10 terms of `(5, 3, 10)`.
178. Write `average(list)` without using `avgOf`. Test with `[10, 20, 30, 40, 50]`.
179. Write `variance(list)` = average of squared deviations from mean. Test `[2, 4, 4, 4, 5, 5, 7, 9]`.
180. Write `stdDev(list)` as `sqrt(variance(list))`. Verify against `stdDevOf`.
181. Print whether `17` is a prime using `isPrime`. Then find the next prime after `17`.
182. Compute `floor(log(n) / log(2))` for `n=64, 128, 256` to get bit length.
183. Compute the number of digits in `n` as `floor(log(n)) + 1` for `n = 12345`.
184. Write `sumOfSquaredDigits(n)`. Test `89` (enters cycle) and `1` (happy number).
185. Write `isHappy(n)` using a loop that detects if sum of squared digits reaches `1` or cycles.
186. Write `isMersennePrime(p)` checking if `2^p - 1` is prime. Test `p=2,3,5,7`.
187. Compute the area of a regular hexagon with side `s=6` using `(3*sqrt(3)/2)*s^2`.
188. Write `roundTo(x, decimals)` using `round(x * 10^decimals) / 10^decimals`. Test `(3.14159, 2)`.
189. Print a multiplication table of `7` from `1` to `12` in one line per row.
190. Compute `sum of abs(i - 5)` for `i` from `1` to `10`.
191. Compute the dot product of vectors `[1, 2, 3]` and `[4, 5, 6]`.
192. Compute the magnitude of vector `[3, 4, 5]` using `sqrt`.
193. Normalize vector `[3, 4]` to unit length and print components.
194. Write `cross2D(ax, ay, bx, by) = ax*by - ay*bx`. Test `(1,0,0,1)`.
195. Write `quadratic(a, b, c)` that returns list of roots. Test `(1, -5, 6)`.
196. Compute `sum of 1/n^2` for `n` from 1 to 1000 (approximation of `pi^2/6`).
197. Write `bernoulli(p, n)` = `p^n * (1-p)^(1-n)` for single trial. Not standard — compute P(X=1) and P(X=0).
198. Print `sin(x)` for `x` from `0` to `360` degrees in steps of `30`.
199. Write `interpolate(points, x)` that does linear interpolation between two nearest points in a list of `[x, y]` pairs.
200. Compute the perimeter and area of a regular polygon with `n=6` sides and side length `s=5`.
201. Write `digitalRoot(n)` and verify that `digitalRoot(493)` equals `7`.
202. Compute the Hamming distance between two integers by XOR-ing and counting bits.
203. Write `intLog2(n)` using a while loop that shifts right. Test `16`, `32`, `255`.
204. Write `popcount(n)` that counts the number of `1` bits. Test `255`, `256`, `170`.
205. Write `isPowerOfTwo(n)` using the fact `n & (n-1) == 0`. Test `8`, `9`, `16`, `15`.
206. Compute the sum of all proper divisors of `220` and `284` to confirm amicable pair.
207. Write `isArmstrong(n)` — a number equal to sum of its digits each raised to power = digit count. Test `153`, `9474`, `100`.
208. Compute and print the average of all primes below `50`.
209. Print whether any number in `[4, 6, 8, 9, 10, 14, 15]` is prime (print each result).
210. Write a function that takes a list of measurements and returns `{min, max, range, mean, stddev}` as a dictionary.
211. Write `binomial(n, k)` using `nCr`. Print Pascal's triangle for `n=0..6`.
212. Print the sum of all integers from `1` to `n` that are divisible by `3` or `5`, for `n=100`.
213. Write `euler_phi(n)` counting integers from `1` to `n` coprime to `n`. Test `n=12`.
214. Print the decimal expansion (first 10 digits) of `sqrt(2)` using `formatDecimal`.
215. Write `truncate(x)` using `floor` for positive and `ceil` for negative. Test `3.7`, `-3.7`.
216. Compute the sum of all two-digit numbers whose digits sum to `9`.
217. Write `cubeRoot(n) = n ^ (1/3)`. Test `27`, `64`, `125`.
218. Compute the Maclaurin series for `sin(x)` using first 5 odd terms. Test `x = pi/4`.
219. Write `heron(a, b, c)` for triangle area using Heron's formula. Test `(3, 4, 5)`.
220. Write `isIsosceles(a, b, c)`, `isEquilateral(a, b, c)`, `isScalene(a, b, c)`.
221. Compute and print `log(x)` for `x` in `[1, 10, 100, 1000, 10000]`.
222. Write `percentile(list, p)` that returns the `p`th percentile of a sorted list.
223. Compute the moving average of `[3, 5, 7, 2, 8, 10, 11, 65, 72, 81, 99, 100, 150]` with window `3`.
224. Write `formatScientific(x)` that prints a number like `3.14e5`.
225. Compute cumulative sums of `[1, 2, 3, 4, 5]` and print the resulting list.
226. Write `entropy(probList)` = `-sum(p * log(p))`. Test `[0.5, 0.25, 0.25]`.
227. Write `fahrenheitToCelsius(f)` and test with `32`, `98.6`, `212`.
228. Compute simple interest `P*R*T/100` for `P=5000`, `R=7`, `T=3`.
229. Write `digitProduct(n)` that multiplies all digits together. Test `2345` and `999`.
230. Write `nthRoot(x, n) = x ^ (1/n)`. Test `(81, 4)`, `(32, 5)`, `(64, 6)`.

---

## Section 3: Text (Problems 231–380)

231. Declare a local `s` set to `"  Hello, Falcon!  "`. Print the trimmed version.
232. Print the length of the string `"Supercalifragilistic"`.
233. Convert `"hello world"` to uppercase and print.
234. Convert `"THE QUICK BROWN FOX"` to lowercase and print.
235. Print whether `"javascript"` contains the substring `"java"`.
236. Print whether `"https://example.com"` starts with `"https"`.
237. Split `"red,green,blue,yellow"` by `","` and print each item on its own line.
238. Split `"   spaced   out   words   "` at spaces and print the resulting list.
239. Split `"key=value"` at first `"="` and print the key and value separately.
240. Reverse the string `"Hello, World!"` and print.
241. Replace every `"cat"` in `"The cat sat on the cat mat"` with `"dog"`.
242. Use `replaceFrom` to replace `"CITY"` and `"COUNTRY"` in `"Welcome to CITY, COUNTRY!"`.
243. Extract the substring from index `2` to length `4` from `"abcdefgh"`.
244. Print `"apple" === "apple"`, `"apple" !== "banana"`, `"apple" << "banana"`, `"banana" >> "apple"`.
245. Print whether `"sky"` is contained in `["sun", "sky", "sea"]` using `containsAny`.
246. Print whether `"The quick fox"` contains all words in `["quick", "fox"]` using `containsAll`.
247. Split `"one;two|three.four"` at any of `[";", "|", "."]` and print the parts.
248. Convert `"Alice,30,Paris"` using `csvRowToList` and print each field.
249. Create a list `["Bob", "25", "Engineer"]` and convert to CSV row with `toCsvRow`.
250. Write a function `countChar(s, ch)` that counts occurrences of a character. Test `("banana", "a")`.
251. Write a function `isPalindrome(s)` using `lowercase` and `reverse`. Test `"Racecar"`, `"hello"`.
252. Write a function `wordCount(s)` using `trim` and `splitAtSpaces`. Test with various sentences.
253. Write `capitalize(s)` that uppercases the first letter. Test `"hello"`, `"world"`.
254. Write `titleCase(sentence)` that capitalizes every word. Test `"the quick brown fox"`.
255. Write `repeatStr(s, n)` using a while loop. Test `("abc", 4)`.
256. Write `truncate(s, maxLen)` that appends `"..."` if too long. Test `("Hello World", 7)`.
257. Write `padLeft(s, width, ch)` that pads the string on the left. Test `("42", 6, "0")`.
258. Write `padRight(s, width, ch)` that pads the string on the right. Test `("Hi", 10, "-")`.
259. Write `center(s, width, ch)` that centers the string. Test `("cat", 9, "*")`.
260. Write `startsWith(s, prefix)` manually using `segment`. Confirm with `("foobar", "foo")`.
261. Write `endsWith(s, suffix)` using `segment` and `textLen`. Test `("foobar", "bar")`.
262. Write `removePrefix(s, prefix)` that strips a prefix if present. Test `("https://x.com", "https://")`.
263. Write `removeSuffix(s, suffix)` that strips a suffix if present. Test `("index.html", ".html")`.
264. Write `countWords(s)` that returns number of words. Test with `"  one  two   three  "`.
265. Write `longestWord(sentence)` that returns the longest word. Test `"the quick brown fox"`.
266. Write `shortestWord(sentence)` that returns the shortest word.
267. Write `reverseWords(sentence)` that reverses word order. Test `"hello world foo"`.
268. Write `isAnagram(a, b)` by comparing sorted character lists. Test `("listen", "silent")`.
269. Write `removeDuplicateChars(s)` that removes repeated characters. Test `"aabbccdd"`.
270. Write `compressSpaces(s)` that collapses multiple spaces to one. Test `"hello   world"`.
271. Write `extractDigits(s)` that returns all digit characters as a string. Test `"a1b2c3"`.
272. Write `removeDigits(s)` that strips numeric characters. Test `"h3ll0 w0rld"`.
273. Write `countVowels(s)` counting `a,e,i,o,u`. Test `"Hello World"`.
274. Write `countConsonants(s)` as `letters - vowels`. Test `"Hello World"`.
275. Write `isAllUppercase(s)` checking `s === s.uppercase()`. Test.
276. Write `isAllLowercase(s)`. Test.
277. Write `alternatingCase(s)` that alternates upper/lower for each character. Test `"hello"`.
278. Write `slugify(s)` that lowercases and replaces spaces with `"-"`. Test `"Hello World"`.
279. Write `camelToSnake(s)` that naively converts by replacing uppercase chars. Describe the logic.
280. Write `rotateString(s, n)` that rotates characters left by `n`. Test `("abcde", 2)`.
281. Write `longestCommonPrefix(list)` of strings. Test `["flower", "flow", "flight"]`.
282. Write `allStartWith(list, prefix)` — true if every string starts with prefix. Test.
283. Write `sortStringsByLength(list)` — ascending. Test `["cat", "elephant", "ox"]`.
284. Write `filterByLength(list, minLen, maxLen)`. Test `["a", "bb", "ccc", "dddd"]` with `(2, 3)`.
285. Write `joinWithAnd(list)` — joins like `"a, b, and c"`. Test `["red", "green", "blue"]`.
286. Write `abbreviate(fullName)` that returns initials. Test `"John Michael Doe"` → `"J.M.D."`.
287. Write `maskEmail(email)` that shows first 2 chars and domain. Test `"alice@example.com"`.
288. Write `maskNumber(s, show)` that masks all but last `show` digits with `*`. Test `("1234567890", 4)`.
289. Write `parseCSVLine(line)` that splits on `","` and trims each field. Test `" Alice , 30 , Paris "`.
290. Write `wrapText(s, maxWidth)` that inserts `"\n"` to wrap at word boundaries.
291. Write `isValidEmail(s)` checking for `@` and `.` using `contains`. Simplistic test.
292. Write `isValidUrl(s)` that checks `startsWith("http")` and `contains(".")`.
293. Write `firstNWords(sentence, n)` that returns the first `n` words. Test `("one two three four", 2)`.
294. Write `lastNWords(sentence, n)`.
295. Write `nthWord(sentence, n)` that returns the nth word. Test `("the quick fox", 2)`.
296. Write `removeStopWords(sentence, stopWords)`. Test with `["the", "a", "is"]`.
297. Write `frequency(s, ch)` that returns frequency ratio of character in string. Test `("banana", "a")`.
298. Write `mostFrequentChar(s)` using a dict. Test `"mississippi"`.
299. Write `longestRun(s)` that finds the longest run of the same character. Test `"aaabbbccccdd"`.
300. Write `encodeRLE(s)` (Run-Length Encoding): `"aaabbc"` → `"3a2b1c"`. Test.
301. Write `reverseEachWord(sentence)`. Test `"hello world"` → `"olleh dlrow"`.
302. Write `interleave(a, b)` that zips two strings character by character. Test `("abc", "123")`.
303. Write `containsNumber(s)` — true if any character is a digit. Test `"hello3world"`, `"hello"`.
304. Write `stripPunctuation(s)` that removes common punctuation characters.
305. Write `levenshtein(a, b)` distance using iterative DP. Test `("kitten", "sitting")`.
306. Build a simple template engine: `render(template, vars_dict)` replaces `{{key}}` patterns. Test.
307. Write `indent(s, n)` that prepends `n` spaces to each line of a multiline string.
308. Write `countLines(s)` by splitting on `"\n"`. Test.
309. Write `trimLines(s)` that trims each line individually. Test.
310. Write `removeBlankLines(s)` that filters out empty lines. Test.
311. Write `toUpperFirstWord(sentence)` — only the first word is uppercased. Test.
312. Write `swapCase(s)` — lowercase becomes uppercase and vice versa. Test `"Hello World"`.
313. Write `isSubstring(s, sub)` using `contains`. Confirm with `("abcdef", "cde")`.
314. Write `indexOfChar(s, ch)` — index of first occurrence. Test `("abcabc", "b")`.
315. Write `lastIndexOf(s, sub)` — last occurrence index. Test `("abcabc", "b")`.
316. Write `replaceFirst(s, old, new)` — replace only the first occurrence. Test.
317. Write `countOccurrences(s, sub)` — count non-overlapping occurrences. Test `("abababab", "ab")`.
318. Write `insertAt(s, pos, insert)` — inserts string at position. Test `("helloworld", 6, " ")`.
319. Write `deleteAt(s, pos, len)` — removes `len` characters starting at `pos`. Test.
320. Write `splitIntoChunks(s, size)` — split string into chunks of given size. Test `("abcdefgh", 3)`.
321. Print the string `"Falcon"` in a diamond shape using loops and text join.
322. Write `toBinaryString(s)` that converts each character to its ASCII code then binary. Test `"Hi"`.
323. Write `isRotation(a, b)` — checks if `b` is a rotation of `a`. Test `("abcde", "cdeab")`.
324. Write `longestPalindromicSubstring(s)` by checking all substrings. Test `"babad"`.
325. Write `generateAcronym(phrase)` — takes first letter of each word. Test `"Artificial Intelligence"`.
326. Write `parseQueryString(qs)` like `"a=1&b=2"` into a dictionary. Test.
327. Write `buildQueryString(dict)` that serializes dictionary into `"k=v&k=v"` format.
328. Write `normalizeWhitespace(s)` that trims and collapses internal spaces.
329. Write `toCamelCase(phrase)` from space-separated words. Test `"hello world test"`.
330. Write `toKebabCase(phrase)` — lowercase, hyphen-separated. Test `"Hello World"`.
331. Write `stripHTMLTags(s)` by detecting `<` and `>` brackets and removing content inside.
332. Print the character frequencies of `"programming"` sorted by frequency descending.
333. Write `isValidIPv4(s)` — splits by `"."` and checks 4 parts, each 0–255.
334. Write `isPangram(s)` — checks if all letters a-z appear. Test `"the quick brown fox jumps over the lazy dog"`.
335. Write `removeDuplicateWords(sentence)`. Test `"the cat sat on the cat mat"`.
336. Write `longestWordLength(sentence)` using `maxOf` on the word-length list.
337. Write `capitalizeEveryOtherWord(sentence)` — alternates capitalization.
338. Write `reverseWordOrder(sentence)` using `splitAtSpaces` and `reverseList`.
339. Write `insertWordAt(sentence, word, pos)` — inserts word at position `pos`. Test.
340. Write `replaceWordAt(sentence, newWord, pos)` — replace the word at position `pos`.
341. Write `removeWordAt(sentence, pos)` — remove the word at position.
342. Write `countUniqueWords(sentence)` by building a set-like list. Test.
343. Write `longestIncreasingSubstringByAlphabet(s)` — find longest run where each char >= previous.
344. Write `buildTable(headers, rows)` that returns a text table as a string. Test simple data.
345. Write `leftPad(s, n)` and `rightPad(s, n)` with spaces.
346. Write `isAlphanumeric(s)` — all chars are letters or digits.
347. Write `isAlpha(s)` — all characters are letters only.
348. Write `isDigitOnly(s)` using `? number`. Test `"1234"`, `"12a4"`.
349. Write `toWords(n)` for numbers 1–19 only (one, two, ..., nineteen). Test `7`, `13`, `19`.
350. Write `pluralize(word, count)` — appends `"s"` if count != 1. Test `(1, "cat")`, `(3, "cat")`.
351. Concatenate a list of 100 strings using reduce with `_`. Print result length.
352. Write `extractBetween(s, open, close)` — extracts text between two delimiters. Test `("(hello)", "(", ")")`.
353. Write `wrapInQuotes(s)` — surrounds string with `"`. Test `"hello"`.
354. Write `repeatChar(ch, n)` using a while loop. Test `("=", 20)`.
355. Write `isValidHex(s)` — only digits `0-9` and letters `A-F`. Use `containsAll` approach.
356. Write `removeChar(s, ch)` — removes all occurrences of a character. Test `("banana", "a")`.
357. Write `countUppercase(s)` — count uppercase letters.
358. Write `countLowercase(s)`.
359. Write `countSpaces(s)` — count space characters.
360. Write `formatName(first, last)` → `"Last, First"` format.
361. Write `initials(name)` — takes full name and returns dot-separated initials.
362. Write `expandTabs(s, tabSize)` — replaces `"\t"` with spaces.
363. Write `toSentenceCase(s)` — first letter capitalized, rest lowercase.
364. Write `startsWithVowel(word)` using `containsAny`. Test `"apple"`, `"banana"`.
365. Write `longestWordStartingWith(sentence, ch)` — longest word starting with given char.
366. Write `summarize(text, maxWords)` — return first `maxWords` words followed by `"..."` if needed.
367. Write `isProperNoun(word)` — first letter uppercase and rest lowercase.
368. Write `repeatWithSep(s, n, sep)` — repeat string `n` times with separator. Test `("ha", 3, "-")`.
369. Write `toPhoneFormat(digits)` — formats `"1234567890"` as `"(123) 456-7890"`.
370. Write `reverseWords(s)` where each word is reversed but order is preserved. Test.
371. Write `pigLatin(word)` — moves first consonant cluster to end, adds `"ay"`.
372. Write `soundex(word)` — simplified Soundex (first letter + digit codes for consonants).
373. Print the Morse code (as a lookup dict) for letters A-E and encode `"HELLO"`.
374. Write `toBraille(digit)` — return a textual 2x3 pattern for digits 1–5 using text.
375. Write `rotateChars(s, n)` — ROT-n cipher. Test `("Hello", 13)` for ROT13.
376. Write `xorEncrypt(s, key)` — XOR each character code with key (simulate with shift math).
377. Write `countSyllables(word)` — simple heuristic: count vowel groups.
378. Write `formatDuration(seconds)` — output `"Xh Ym Zs"`. Test `3661`.
379. Write `parseDate(s)` — parse `"YYYY-MM-DD"` into dict `{year, month, day}`. Test.
380. Write `formatDate(year, month, day)` — output `"DD/MM/YYYY"`. Test.

---

## Section 4: Lists (Problems 381–560)

381. Declare a list `[10, 20, 30, 40, 50]`. Print the first and last elements.
382. Declare a list of 5 fruits. Print its length.
383. Add `"kiwi"` to a list of fruits and print the updated list.
384. Insert `"mango"` at index `2` in `["apple", "banana", "cherry"]` and print.
385. Remove the second element from `["a", "b", "c", "d"]` and print.
386. Print the index of `"dog"` in `["cat", "dog", "bird"]`.
387. Print whether `7` is in the list `[2, 4, 6, 8, 10]`.
388. Append `[4, 5, 6]` to `[1, 2, 3]` and print the combined list.
389. Print the reverse of `[1, 2, 3, 4, 5]`.
390. Sort and print `[5, 2, 8, 1, 9, 3]`.
391. Print `[10, 20, 30, 40, 50, 60].slice(2, 4)`.
392. Print a random element from `["rock", "paper", "scissors"]` five times.
393. Print `allButFirst` and `allButLast` of `[1, 2, 3, 4, 5]`.
394. Use `map` to triple every element of `[1, 2, 3, 4, 5]`.
395. Use `filter` to keep only values > 5 from `[1, 3, 5, 7, 9, 11]`.
396. Use `reduce(0)` to sum `[10, 20, 30, 40]`.
397. Use `reduce(1)` to compute the product of `[1, 2, 3, 4, 5]`.
398. Sort `["banana", "apple", "date", "cherry"]` alphabetically.
399. Sort `[5, 3, 8, 1]` in descending order using sort lambda.
400. Find the longest name in `["Alice", "Bob", "Christopher", "Di"]` using max lambda.
401. Find the shortest name using min lambda.
402. Use `join` to concatenate `["one", "two", "three"]` with `" | "` separator.
403. Use `lookupInPairs` to find the capital of `"Japan"` in `[["France","Paris"],["Japan","Tokyo"]]`.
404. Convert pairs `[["a",1],["b",2],["c",3]]` to a dict using `pairsToDict`.
405. Copy a list with `copyList` and verify mutations to the copy don't affect the original.
406. Write `flatten(nested)` using `reduce` and `appendList`. Test `[[1,2],[3,4],[5]]`.
407. Write `unique(list)` that removes duplicates. Test `[1,1,2,3,3,4,2]`.
408. Write `zip(a, b)` that pairs elements. Test `(["a","b","c"], [1,2,3])`.
409. Write `chunk(list, n)` that splits a list into groups of n. Test `([1..9], 3)`.
410. Write `sum(list)` using reduce. Test `[1,2,3,4,5]`.
411. Write `product(list)` using reduce.
412. Write `max(list)` manually using a while loop.
413. Write `min(list)` manually.
414. Write `countOccurrences(list, val)` using filter. Test `([1,2,2,3,2], 2)`.
415. Write `rotateLeft(list, n)`. Test `([1,2,3,4,5], 2)`.
416. Write `rotateRight(list, n)`. Test `([1,2,3,4,5], 2)`.
417. Write `intersect(a, b)` using filter. Test `([1,2,3,4],[2,4,6])`.
418. Write `difference(a, b)` — elements in a not in b.
419. Write `union(a, b)` — all unique elements from both.
420. Write `partition(list, thresh)` — split into above and below threshold.
421. Write `transpose(matrix)` for a 3x3 grid using index loops.
422. Write `matMul(a, b)` for 2x2 matrix multiplication.
423. Compute the row sums of `[[1,2,3],[4,5,6],[7,8,9]]`.
424. Compute the column sums of the same matrix.
425. Write `flatten2D(matrix)` that flattens a list of lists. Test.
426. Write `diagonalSum(matrix)` — sum of main diagonal. Test 3x3 grid.
427. Write `reverseEachRow(matrix)` — reverse each inner list. Test.
428. Write `cumSum(list)` that returns the running cumulative sum. Test `[1,2,3,4,5]`.
429. Write `cumProduct(list)` — running product. Test.
430. Write `diff(list)` — consecutive differences. Test `[1,4,9,16,25]`.
431. Write `movingAvg(list, k)` — window average. Test `([1,2,3,4,5,6,7], 3)`.
432. Write `normalize(list)` — scale all values to [0,1]. Test `[0,5,10,15,20]`.
433. Write `standardize(list)` — zero mean, unit variance. Test.
434. Write `dotProduct(a, b)` of two numeric lists. Test `([1,2,3],[4,5,6])`.
435. Write `scalarMultiply(list, k)` — multiply each element by k. Test.
436. Write `addLists(a, b)` — element-wise addition. Test.
437. Write `subtractLists(a, b)` — element-wise subtraction.
438. Write `zipWith(a, b)` — pair elements, using list of pairs. Test.
439. Write `histogram(list)` — returns dict of value frequencies. Test.
440. Write `mostCommon(list)` — the most frequent element.
441. Write `leastCommon(list)`.
442. Write `groupBy(list, keyFn)` — group items by key function output. Simulate by grouping numbers by parity.
443. Write `tally(list)` — returns list of `[value, count]` pairs.
444. Write `topN(list, n)` — top n largest elements. Test `([3,1,4,1,5,9,2,6], 3)`.
445. Write `bottomN(list, n)` — bottom n smallest.
446. Write `range(start, end, step)` — generates a list of numbers. Test `(0, 10, 2)`.
447. Write `repeat(val, n)` — list of `n` copies. Test `("x", 5)`.
448. Write `take(list, n)` — first n elements. Test `([1..10], 4)`.
449. Write `drop(list, n)` — remove first n elements. Test.
450. Write `takeWhile(list, cond)` — take elements while condition holds. Test take while < 5.
451. Write `dropWhile(list, cond)` — drop elements while condition holds.
452. Write `indexOf(list, val)` manually. Confirm with built-in.
453. Write `lastIndexOf(list, val)` — index of last occurrence.
454. Write `removeAll(list, val)` — remove all occurrences of value.
455. Write `replaceAll(list, old, new)` — replace all occurrences.
456. Write `replaceAt(list, idx, val)` — replace element at index.
457. Write `insertBefore(list, val, target)` — insert before first occurrence of target.
458. Write `insertAfter(list, val, target)`.
459. Write `splitAt(list, idx)` — returns `[left, right]`. Test `([1,2,3,4,5], 3)`.
460. Write `mergeAlternating(a, b)` — interleave two lists. Test.
461. Write `runLengthEncode(list)` — encode runs. Test `[1,1,1,2,2,3]`.
462. Write `runLengthDecode(encoded)` — decode it back. Test.
463. Write `allEqual(list)` — true if all elements are the same. Test.
464. Write `anyEquals(list, val)` — true if any element equals val.
465. Write `noneEquals(list, val)`.
466. Write `countIf(list, pred)` — count elements satisfying a condition. Simulate with evens.
467. Write `sumIf(list, pred)` — sum elements satisfying condition.
468. Write `maxIf(list, pred)` — max of elements satisfying condition.
469. Write `filterMap(list)` — filter then map in one pass (filter numbers, then double them). Test.
470. Write `flatMap(list)` — maps each element to a list then flattens. Test.
471. Build a stack using a list: `push(stack, val)`, `pop(stack)`, `peek(stack)`. Simulate 5 ops.
472. Build a queue using a list: `enqueue(q, val)`, `dequeue(q)`. Simulate 5 ops.
473. Write `isSorted(list)` — returns true if ascending. Test.
474. Write `isSortedDescending(list)`.
475. Write `binarySearch(list, val)` — return index or -1. Test.
476. Write `insertSorted(list, val)` — insert into already-sorted list. Test.
477. Write `mergeSorted(a, b)` — merge two sorted lists into one sorted list.
478. Write `quickSort(list)` using filter for less/equal/greater partition + recursive merge.
479. Write `insertionSort(list)` using shift loop. Test.
480. Write `selectionSort(list)`. Test.
481. Write `countSortedPairs(list)` — count pairs `(i,j)` where `i<j` and `list[i] < list[j]`.
482. Write `longestIncreasingSubsequence(list)` — return its length. Test `[3,1,4,1,5,9,2,6]`.
483. Write `maxSubarraySum(list)` using Kadane's algorithm. Test `[-2,1,-3,4,-1,2,1,-5,4]`.
484. Write `twoSum(list, target)` — return indices of two numbers that sum to target. Test.
485. Write `threeSum(list, target)` — return a triple summing to target.
486. Write `productExceptSelf(list)` — each element is product of all others. Test.
487. Write `majorityElement(list)` — element appearing more than n/2 times. Test.
488. Write `missingNumber(list)` from `[0..n]` with one missing. Test.
489. Write `duplicateNumber(list)` from `[1..n]` with one duplicate. Test.
490. Write `longestConsecutiveSequence(list)` — length of longest consecutive run. Test.
491. Write a simple matrix addition for two 2x2 matrices represented as `[[a,b],[c,d]]`.
492. Compute the trace (sum of diagonal) of `[[1,2,3],[4,5,6],[7,8,9]]`.
493. Write `isMagicSquare(matrix)` — row, column, diagonal sums all equal. Test 3x3.
494. Write `spiralOrder(matrix)` — return elements of matrix in spiral order. Test 3x3.
495. Write `rotateMatrix90(matrix)` — rotate 3x3 clockwise. Test.
496. Write `transposeMatrix(matrix)` — flip rows and columns. Test.
497. Write `multiplyListByScalar(list, k)` using map. Test.
498. Write `sumPairs(list)` — sum every consecutive pair. Test `[1,2,3,4,5,6]` → `[3,7,11]`.
499. Write `rollingMax(list)` — running maximum so far. Test.
500. Write `rollingMin(list)` — running minimum so far.
501. Write `deduplicateConsecutive(list)` — remove consecutive duplicates. Test `[1,1,2,2,3,1,1]`.
502. Write `windows(list, size)` — return all sliding windows of given size. Test `([1..5], 3)`.
503. Write `pairwiseSum(list)` — list of sums of adjacent pairs. Test.
504. Write `enumerate(list)` — returns list of `[index, value]` pairs.
505. Write `findAll(list, val)` — all indices of matching value.
506. Write `groupConsecutive(list)` — group equal consecutive elements. Test `[1,1,2,3,3,3]`.
507. Write `interleaveLists(lists)` — interleave elements from multiple lists. Test.
508. Write `cartesianProduct(a, b)` — all pairs. Test `([1,2],["a","b"])`.
509. Write `permutations(list)` — all permutations of a short list. Test `[1,2,3]`.
510. Write `combinations(list, k)` — all k-element subsets. Test `([1,2,3,4], 2)`.
511. Write `powerset(list)` — all subsets. Test `[1,2,3]`.
512. Write `shuffle(list)` — pseudo-random reorder using randInt swaps. Test.
513. Write `sample(list, k)` — pick k random distinct elements.
514. Write `weightedRandom(list, weights)` — pick element based on relative weight.
515. Write `frequencies(list)` — dict of element counts. Test.
516. Write `percentile(sortedList, p)` — value at p-th percentile.
517. Write `quartiles(list)` — return Q1, Q2, Q3. Test `[1..9]`.
518. Write `outliers(list)` — elements more than 2 std devs from mean.
519. Write `removeOutliers(list)` — filter them out.
520. Write `smoothen(list, k)` — replace each element with average of neighbors. Test.
521. Write `encodeBase64(list)` — encode bytes (0-255) in simplified Base64 alphabet.
522. Write `decodeBase64(s)` — reverse.
523. Write `histogram(list, bins)` — count values falling into equal-width bins.
524. Write `zipToDict(keys, values)` — create dict from two lists. Test.
525. Write `invertListToDict(list)` — maps value → index. Test.
526. Write `updateAtIndex(list, idx, fn)` — apply function name to element at index.
527. Write `removeNulls(list)` — remove items that equal `0` or `""`.
528. Write `compact(list)` — remove falsy values (0, false, "").
529. Write `mergeListsOfDicts(listA, listB, key)` — merge dicts from both lists by key.
530. Write `groupByFirstChar(list)` — group strings by first character into dict. Test.
531. Write `longestAscendingRun(list)` — return the longest ascending consecutive run. Test.
532. Write `kMostFrequent(list, k)` — top k elements by frequency. Test.
533. Write `chunksOf(list, n)` and test on a list of 10 items with chunk size 4.
534. Write `sliding(list, n)` — all size-n windows. Test.
535. Write `countBefore(list, val)` — count elements before the first occurrence. Test.
536. Write `countAfter(list, val)` — count elements after the last occurrence. Test.
537. Write `splitOn(list, sep)` — split list on separator element. Test `([1,0,2,0,3], 0)`.
538. Write `mergeConsecutiveDuplicates(list)` — merge runs of equal items into `[val, count]`. Test.
539. Write `applyToEvenIndices(list, fn)` — apply named function to even-indexed elements.
540. Write `applyToOddIndices(list, fn)`.
541. Write `pairwiseProduct(a, b)` — element-wise product of two lists. Test.
542. Write `cumMax(list)` — running maximum list. Test.
543. Write `cumMin(list)`. Test.
544. Write `interquartileRange(list)` — Q3 - Q1. Test.
545. Write `zScore(list, x)` — how many std deviations from mean. Test.
546. Write `cosineSimiliarity(a, b)` — dot product / (|a| * |b|). Test.
547. Write `euclideanDistance(a, b)` as sqrt of sum of squared differences. Test.
548. Write `manhattanDistance(a, b)` as sum of abs differences. Test.
549. Write `chebyshevDistance(a, b)` as max of abs differences. Test.
550. Write `hammingDistance(a, b)` — count positions where elements differ. Test.
551. Write `pearsonCorrelation(x, y)` — covariance / (stddev_x * stddev_y). Test.
552. Write `linearRegression(x, y)` — return slope and intercept. Test simple data.
553. Write `predict(slope, intercept, x)` — linear prediction. Test.
554. Write `residuals(x, y, slope, intercept)` — differences between actual and predicted.
555. Write `meanAbsoluteError(predictions, actuals)`. Test.
556. Write `meanSquaredError(predictions, actuals)`. Test.
557. Write `rootMeanSquaredError(predictions, actuals)` using sqrt. Test.
558. Write `confusionMatrix(predicted, actual)` — count TP, TN, FP, FN. Test binary lists.
559. Write `accuracy(predicted, actual)`. Test.
560. Write `balanceParentheses(list)` — count `(` and `)` items and verify balance.

---

## Section 5: Dictionaries (Problems 561–690)

561. Create a dict `{"name": "Alice", "age": 30}`. Print the name.
562. Add a key `"city"` set to `"Paris"` to an existing dict and print.
563. Delete a key from a dict and confirm with `containsKey`.
564. Print the number of keys in `{"a": 1, "b": 2, "c": 3}`.
565. Check if `"email"` exists in `{"name": "Bob", "phone": "123"}`.
566. Print all keys and all values of a dict separately.
567. Iterate over `{"France":"Paris","Japan":"Tokyo","Brazil":"Brasilia"}` and print each pair.
568. Merge `{"a":1,"b":2}` with `{"c":3,"d":4}` using `mergeInto`.
569. Copy a dict using `copyDict` and show that mutations don't affect the original.
570. Convert a dict to pairs using `toPairs` and iterate.
571. Use `getAtPath` to get `"age"` from `{"user":{"profile":{"age":25}}}`.
572. Use `setAtPath` to set `["user","profile","score"]` to `100` in a nested dict.
573. Use `walkTree` to reach `["root","child"]` in a nested dict.
574. Count how many values in `{"a":5,"b":0,"c":3,"d":0}` are greater than zero.
575. Write `invertDict(d)` — swap keys and values. Test `{"red":"#F00","green":"#0F0"}`.
576. Write a word counter using a dict. Count words in `"the cat sat on the mat the cat"`.
577. Write `filterByValue(d, threshold)` — keep entries where value > threshold. Test.
578. Write `mapValues(d, fn)` — transform all values. Simulate doubling all numeric values.
579. Write `filterKeys(d, list)` — keep only keys in a given list. Test.
580. Write `renameKey(d, old, new)` — rename a key. Test.
581. Write `merge(a, b)` — return new dict with all keys from both. Test.
582. Write `deepMerge(a, b)` — recursively merge nested dicts. Test 2-level nesting.
583. Write `pick(d, keys)` — return new dict with only the picked keys. Test.
584. Write `omit(d, keys)` — return new dict without the listed keys.
585. Write `groupBy(list, key)` — group list of dicts by a shared key. Test.
586. Write `keyWithMaxValue(d)` — return key whose value is largest. Test.
587. Write `keyWithMinValue(d)`.
588. Write `sortedByValue(d)` — return list of `[key, value]` pairs sorted by value. Test.
589. Write `sortedByKey(d)` — sorted by key alphabetically.
590. Write `frequencyDict(list)` — map each element to its count. Test.
591. Write `sumValues(d)` — sum all numeric values. Test.
592. Write `avgValues(d)` — average of all values.
593. Write `maxValue(d)` — largest value.
594. Write `minValue(d)`.
595. Write `containsValue(d, val)` — true if any value equals val. Test.
596. Write `findKeyByValue(d, val)` — return key for given value. Test.
597. Write `allKeysStartWith(d, prefix)` — check all keys start with prefix. Test.
598. Write `subdict(d, from, to)` — return entries within alphabetical key range. Test.
599. Write `transformKeys(d, fn)` — apply function name to all keys. Simulate uppercasing keys.
600. Write `zip(keys, values)` — build dict from two lists. Test.
601. Write `unzip(d)` — return separate lists of keys and values.
602. Write `diff(a, b)` — keys in a not in b and vice versa. Test.
603. Write `intersection(a, b)` — keys present in both dicts. Test.
604. Write `union(a, b)` — all keys from both (a takes precedence). Test.
605. Build a frequency histogram as a dict and print bar chart using text repeat. Test `[1,2,2,3,3,3,4,4,4,4]`.
606. Write `topN(d, n)` — top n entries by value. Test.
607. Write `bottomN(d, n)` — bottom n by value.
608. Build a nested dict representing a tree and write `countLeaves(tree)`.
609. Write `flatten(d, sep)` — flatten nested dict with key paths. Test `{"a":{"b":1}}` → `{"a.b": 1}`.
610. Write `setDefault(d, key, default)` — set key only if not present. Test.
611. Write `increment(d, key)` — increment numeric value by 1. Test.
612. Write `decrement(d, key)`.
613. Write `multiSet(d, keys, val)` — set multiple keys to same value.
614. Write `swap(d, k1, k2)` — swap values of two keys. Test.
615. Write `clearKeys(d, list)` — set multiple keys to empty/zero. Test.
616. Write `jsonLike(d)` — print a readable string representation of a dict.
617. Build a contacts book: add, lookup, delete, list all. Use global dict.
618. Build a grade book: student → list of grades. Add grades, compute averages. Test.
619. Build a shopping cart: item → quantity. Add, remove, update, total. Test.
620. Build a config store: nested dict with `get` and `set`. Test two levels of nesting.
621. Write `countUniqueValues(d)` — count distinct values across all keys.
622. Write `valuesMatchingKey(d, pattern)` — return values for keys containing pattern. Test.
623. Write `longestKey(d)` — key with most characters.
624. Write `shortestKey(d)`.
625. Write `longestValue(d)` — key with longest string value.
626. Write `mergeLists(a, b, key)` — merge list of dicts on common key. Test.
627. Write `indexBy(list, key)` — turn list of dicts into dict indexed by key. Test.
628. Write `pluck(list, key)` — extract a single field from list of dicts. Test.
629. Write `sumField(list, key)` — sum a numeric field across all dicts. Test.
630. Write `avgField(list, key)`.
631. Write `groupCount(list, key)` — count items per group value. Test.
632. Write `pivot(list, rowKey, colKey, valKey)` — simplified pivot. Test.
633. Build an adjacency list graph: dict mapping node → list of neighbors. Write `addEdge`, `neighbors`.
634. Write `hasPath(graph, start, end)` using BFS-like iteration. Test.
635. Write `degree(graph, node)` — count of neighbors. Test.
636. Write `isConnected(graph)` — simple check all nodes reachable from first. Test small graph.
637. Build a simple LRU cache simulation using a list for order and dict for storage.
638. Build a memoize wrapper using a global dict as cache. Test with `fib`.
639. Write `flattenList(d)` — where values are lists, merge them all into one list. Test.
640. Write `invertMultimap(d)` — where values are lists, reverse to value→key mapping. Test.
641. Write `applyAll(d, fn)` — apply a string-named function to all values. Test.
642. Write `mergeWith(a, b, fn)` — merge, applying function on conflicts. Test summing conflicts.
643. Write `updateNested(d, path, fn)` — apply function to value at path. Test.
644. Write `deepEqual(a, b)` — check if two dicts have same key-value structure. Test.
645. Write `selectByType(d, typeName)` — return entries where value matches `? typeName`. Test.
646. Write `histogramDict(list)` and verify by checking all counts sum to list length. Test.
647. Write `bimap(d, keyFn, valFn)` — transform both keys and values. Test.
648. Write `groupAndSum(list, groupKey, sumKey)` — group by one field, sum another. Test.
649. Write `multiKeyLookup(d, keys)` — return list of values for given keys (missing = default). Test.
650. Write `compressDict(d)` — remove keys with falsy values (0, "", false). Test.
651. Write `toSortedList(d)` — sorted list of `[key, value]` pairs by key. Test.
652. Build a frequency-sorted word list from a sentence. Test `"to be or not to be"`.
653. Write `reverseIndex(text)` — maps each word to list of positions where it appears. Test.
654. Write `buildTrie(words)` — nested dict representing a trie. Test `["cat","car","card"]`.
655. Write `trieContains(trie, word)` — check word in trie. Test.
656. Build a settings hierarchy: global defaults, user overrides. Write `resolve(key)` that checks user first.
657. Write `diffDicts(a, b)` — return keys added, removed, changed. Test.
658. Write `patchDict(base, patch)` — apply a patch (additions/changes from patch dict). Test.
659. Write `snapshotAndRestore(d)` — copy, mutate, then restore from snapshot. Test.
660. Write `memoize(cache, key, computeFn)` — check cache first, then compute and store. Test.
661. Write `countByProperty(list, prop)` — count items per value of a given property. Test.
662. Write `maxByField(list, field)` — return dict with highest value in field. Test.
663. Write `minByField(list, field)`.
664. Write `sortByField(list, field)` — sort list of dicts by a numeric field. Test.
665. Write `filterByField(list, field, value)` — keep dicts where field equals value. Test.
666. Build a phone directory: name → phone. Add 5 entries, search by name. Test.
667. Build an inventory system: item → `{qty, price}`. Write `totalValue()` that sums qty*price.
668. Write `deepGet(d, path, default)` — safely navigate nested dicts. Test with missing path.
669. Write `allValues(d)` that recursively collects all leaf values from a nested dict. Test.
670. Write `flatKeys(d, prefix)` — recursively collect all key paths as strings. Test.
671. Write `anyValueAbove(d, threshold)` — true if any value > threshold. Test.
672. Write `allValuesAbove(d, threshold)`.
673. Write `transformByType(d)` — double numbers, uppercase strings. Test mixed dict.
674. Write `jsonPath(d, path)` — navigate using dot-separated path string. Test `"user.address.city"`.
675. Write `setJsonPath(d, path, val)` — set via dot-separated path. Test.
676. Write `mergeDefaults(config, defaults)` — fill in missing keys from defaults. Test.
677. Build an event emitter: dict mapping event name → list of handler names. Write `on`, `emit`.
678. Build a simple state machine: dict mapping `state → {event → nextState}`. Write `transition`.
679. Write `countNested(d)` — total number of key-value pairs at all levels. Test.
680. Write `toDotNotation(d)` — flatten nested dict to dot-notation keys. Test.
681. Write `fromDotNotation(flatDict)` — reconstruct nested dict from dot-notation. Test.
682. Write `validateSchema(d, schema)` — check required keys exist and values match types. Test.
683. Write `defaults(d, keys, val)` — ensure all keys exist, set to val if missing. Test.
684. Write `pickRandom(d)` — return a random key-value pair. Test.
685. Write `toList(d)` — list of `{key, value}` dicts. Test.
686. Write `fromList(list)` — reconstruct dict from list of `{key, value}` dicts. Test.
687. Write `shallowClone(d)` and `deepClone(d)` — the latter recursively copies nested dicts. Test.
688. Write `assignIfEmpty(d, key, val)` — only assigns if key is missing or value is "". Test.
689. Build a simple object store: save/load by ID using global dict. Write `save(id, obj)` and `load(id)`.
690. Write `frequencyRank(list)` — returns dict mapping each item to its rank by frequency. Test.

---

## Section 6: Colors (Problems 691–770)

691. Declare and print color literals `#FF0000`, `#00FF00`, `#0000FF`.
692. Create a coral color using `makeColor([255, 127, 80])` and print it.
693. Split `makeColor([128, 64, 192])` using `splitColor` and print alpha, R, G, B.
694. Write `lighten(color, amount)` that adds amount to each RGB channel. Test with `makeColor([100,100,100])`.
695. Write `darken(color, amount)` that subtracts from each channel. Test.
696. Write `blendColors(a, b, t)` — linear interpolation. Test `(white, black, 0.5)`.
697. Write `toGrayscale(color)` — average the three channels. Test a vibrant color.
698. Write `invertColor(color)` — subtract each channel from 255. Test.
699. Write `colorPalette(base, steps)` — fade from base to black in steps. Test with 5 steps.
700. Write `rgbToHex(r, g, b)` using `decToHex`. Test `(255, 128, 0)`.
701. Write `brightness(color)` using luminance formula `(299*R + 587*G + 114*B)/1000`. Test.
702. Write `isLightColor(color)` — brightness > 128. Test white, black, gray.
703. Write `complementaryColor(color)` — hue rotation by 180 (invert channels). Test.
704. Write `triadic(color)` — return a list of 3 colors spaced 120° apart (simulate via channel rotation). Test.
705. Write `splitColor` to extract and reassemble — set alpha to 128 and rebuild. Test.
706. Write `setRed(color, r)` — return color with modified red channel. Test.
707. Write `setGreen(color, g)`.
708. Write `setBlue(color, b)`.
709. Write `getAlpha(color)` — return alpha component. Test.
710. Write `fullyOpaque(color)` — set alpha to 255. Test.
711. Write `fullyTransparent(color)` — set alpha to 0. Test.
712. Write `withAlpha(color, a)` — set alpha to given value. Test.
713. Write `redShift(color, amount)` — increase red by amount, clamp. Test.
714. Write `blueShift(color, amount)`. Test.
715. Write `colorDistance(a, b)` — Euclidean distance in RGB space. Test.
716. Write `nearestColor(target, palette)` — find closest color in a list by RGB distance. Test.
717. Write `warmUp(color)` — increase R, decrease B by 20. Test.
718. Write `coolDown(color)` — increase B, decrease R by 20. Test.
719. Write `saturate(color, amount)` — move channels away from gray. Test.
720. Write `desaturate(color, amount)` — move channels toward gray. Test.
721. Write `makeGradient(colorA, colorB, steps)` — list of blended colors. Test with 5 steps.
722. Write `applyGamma(color, gamma)` — apply gamma correction. Test `gamma=2.2`.
723. Write `clampColor(color)` — ensure all channels are 0–255. Test with out-of-range values.
724. Write `randomColor()` using `randInt(0, 255)` for each channel. Print 5 random colors.
725. Write `randomBrightColor()` — ensure brightness > 180. Print 3.
726. Write `randomDarkColor()` — brightness < 80. Print 3.
727. Write `contrastRatio(a, b)` — ratio of luminances. Test white on black.
728. Write `pastelify(color)` — blend each channel toward 255 by 50%. Test.
729. Write `sepia(color)` — apply sepia tone formula. Test.
730. Write `redChannel(color)` — extract red as 0–255 integer. Test.
731. Write `greenChannel(color)`.
732. Write `blueChannel(color)`.
733. Write `isGrayscale(color)` — R == G == B. Test.
734. Write `colorToString(color)` — format as `"rgb(R, G, B)"`. Test.
735. Write `averageColors(list)` — component-wise average. Test 3 colors.
736. Write `mostLuminous(palette)` — color with highest brightness. Test.
737. Write `leastLuminous(palette)`.
738. Write `dominantChannel(color)` — which of R, G, B is largest. Test.
739. Write `tint(color, t)` — mix color with white by factor t. Test `t=0.3`.
740. Write `shade(color, t)` — mix color with black by factor t. Test.
741. Write `tone(color, t)` — mix color with gray by factor t. Test.
742. Write `colorMatrix(colors, rows, cols)` — build a 2D grid of colors and print indices.
743. Write `colorFromHSV(h, s, v)` — simplified HSV to RGB conversion. Test `(0, 1, 1)` for red.
744. Write `colorToHSV(color)` — simplified RGB to HSV. Test.
745. Write `isWarm(color)` — R > B. Test.
746. Write `isCool(color)` — B > R. Test.
747. Write `colorHistogram(imageColors)` — count frequency of each color in a list. Test.
748. Write `mostFrequentColor(palette)` — find the mode color. Test.
749. Write `contrastingTextColor(bgColor)` — return black or white based on brightness. Test.
750. Write `colorMix(list)` — mix all colors in a list equally. Test 4 colors.
751. Write `applyMask(color, mask)` — AND each channel with mask value. Test.
752. Write `colorXor(a, b)` — XOR each channel. Test.
753. Write `colorAnd(a, b)` — AND each channel.
754. Write `colorOr(a, b)` — OR each channel.
755. Write `normalizeColor(color)` — scale each channel to 0.0–1.0 list. Test.
756. Write `denormalizeColor(normalized)` — scale back from 0–1 to 0–255. Test.
757. Write `colorLerp(a, b, t)` (same as blendColors but as lerp). Test.
758. Write `colorRotate(color, degrees)` — rotate hue by degrees (simplified). Test `90`.
759. Write `splitComplementary(color)` — one complement and two 30° offsets. Test.
760. Write `colorBrightness(color)` — return text `"bright"`, `"medium"`, or `"dark"`. Test.
761. Write `applyContrast(color, factor)` — multiply deviation from 128 by factor. Test.
762. Write `colorPosterize(color, levels)` — reduce channel precision to n levels. Test.
763. Write `colorSolarize(color, threshold)` — invert channels above threshold. Test.
764. Write `checkerPattern(colorA, colorB, size)` — print color labels in alternating grid.
765. Write `colorToAnsi(color)` — return escape code string for terminal color. Test simple.
766. Write `sortColorsByBrightness(palette)` — ascending brightness order. Test.
767. Write `colorIsRedish(color)` — R > G && R > B. Test.
768. Write `colorIsGreenish(color)` — G > R && G > B.
769. Write `colorIsBlueish(color)` — B > R && B > G.
770. Write `buildColorTable(names, colors)` — dict mapping name → color. Test with 5 entries.

---

## Section 7: Controls (Problems 771–880)

771. Print `"Hot"`, `"Warm"`, or `"Cold"` based on temperature `35`.
772. Use if-else as an expression to assign a label to a score variable.
773. Print numbers `1` to `10` using a while loop.
774. Print numbers `10` down to `1` using a while loop.
775. Use a for range loop to print odd numbers from `1` to `19`.
776. Use a for range loop with step `3` to print `0, 3, 6, ..., 30`.
777. Use for-each to print every item in `["alpha", "beta", "gamma", "delta"]`.
778. Use for-each over a dict to print all `key: value` pairs.
779. Use `break` to exit a loop when a value `7` is found in `[1, 4, 7, 9, 2]`.
780. Use a while loop to compute `2^n` where `n` grows until the result exceeds `1000`.
781. Implement FizzBuzz from `1` to `50`.
782. Implement FizzBuzzWoof — also print `"Woof"` for multiples of `7`.
783. Print a triangle of `*` using nested for loops — row `i` prints `i` stars.
784. Print an inverted triangle — row `i` prints `10 - i` stars.
785. Print a multiplication table for `9` using a for loop.
786. Print the first `n=20` triangular numbers using a loop.
787. Compute the sum of digits of a user-given number using a while loop.
788. Find the first number greater than `100` that is divisible by `7` and `11`.
789. Count how many numbers from `1` to `500` are perfect squares.
790. Print all palindromic numbers from `100` to `999` using a loop.
791. Print all numbers from `1` to `100` whose digit sum equals `10`.
792. Simulate rolling a die 100 times and count how many times each face appears.
793. Find the largest power of `2` less than `1000` using a while loop.
794. Print the Fibonacci sequence until the value exceeds `1000`.
795. Compute `n!` using a for loop for `n=15`.
796. Count the number of vowels in `"The quick brown fox jumps over the lazy dog"` using a for-each.
797. Find the first duplicate in a list `[3, 1, 4, 1, 5, 9, 2, 6, 5]` using a loop.
798. Print numbers from `1` to `1000` that are both odd and prime.
799. Simulate a guessing game — loop until a `randInt` produces the target `42` or after 50 tries.
800. Use a while loop to convert a decimal number to binary step-by-step.
801. Print Pascal's triangle up to 7 rows using nested loops.
802. Find the sum of all prime numbers below `100` using a loop.
803. Count the number of steps in the Collatz sequence for each number from `1` to `20`.
804. Print all Armstrong numbers below `1000`.
805. Print the first 15 perfect numbers (or try to — note they're rare, stop at a limit).
806. Write a loop that collects numbers until their sum exceeds `100`. Print count and final sum.
807. Simulate a simple traffic light: loop through `["red","yellow","green"]` 4 times.
808. Write a loop that finds the longest string in a list without using built-in max.
809. Use a for loop to build a list of squares `[1, 4, 9, 16, ..., 100]`.
810. Use a for loop to check if all elements of a list are positive.
811. Use a while loop to find the GCD of two numbers iteratively.
812. Use a while loop with break to find the square root approximation via Newton's method.
813. Print all two-digit numbers where the product of digits > 20.
814. Simulate a dice game: roll until you get three sixes in a row. Count rolls.
815. Write a loop that computes running average and stops when average > 50 after adding random ints.
816. Use nested if-else inside a for loop to classify each element of a list as small/medium/large.
817. Print a diamond pattern using two for loops (growing then shrinking).
818. Use a for loop to count sign changes in `[1, -2, 3, -4, 5, -6]`.
819. Use a while loop to implement binary search manually. Test.
820. Build a simple state machine loop: start in state `0`, transition based on random inputs.
821. Count how many even numbers appear before the first odd in a list. Test.
822. Find the index of the maximum element in a list using a for loop.
823. Loop over a list and print a running maximum and minimum simultaneously.
824. Simulate a coin toss loop: track longest streak of heads. Test 100 flips.
825. Use nested loops to print all pairs `(i, j)` where `i + j == 10`, `i` and `j` in `[1..9]`.
826. Use a for loop to implement `map` manually: double each element and collect.
827. Use a for loop to implement `filter` manually: keep only even elements.
828. Use a for loop to implement `reduce` manually: compute product.
829. Print all Pythagorean triples with sides up to `20` using three nested loops.
830. Use a while loop to simulate repeated averaging: `x = (x + 1/x) / 2` converging to sqrt(2).
831. Use a for loop to convert a list of Celsius values to Fahrenheit.
832. Print the first `n` terms of the harmonic series `1 + 1/2 + 1/3 + ...` as running sums.
833. Find the smallest `n` such that `n!` has more than 10 digits.
834. Simulate a lottery: pick 6 unique numbers from 1–49 using a while loop.
835. Use a for loop to compute a checksum by summing alternate digits. Test.
836. Write a loop to detect if a list is sorted ascending. Stop at first violation.
837. Use a for loop over a dict to find the key with the maximum value.
838. Write a for-each loop to compute the length of the longest word in a list.
839. Use a while loop to repeatedly halve a number until it drops below `1`. Count steps.
840. Use nested loops to build a multiplication table as a list of lists.
841. Write a for loop to count how many list elements are above the average.
842. Use a while loop to simulate Euclid's extended algorithm.
843. Implement selection sort using for loops. Test on `[5,3,8,6,1,9,2,7,4]`.
844. Write a for-each loop to build a frequency dict from a list of words.
845. Use a for range loop to print a checkerboard pattern of `X` and `.`.
846. Simulate a vending machine: loop through purchases, deducting from stock dict. Test.
847. Use a for loop to walk through a string character by character and print each index and char.
848. Print all composite numbers from `4` to `50` using a for loop and prime check.
849. Write a while loop that performs a bubble sort pass and repeat until sorted. Test.
850. Use a for loop to compute `sin(x)` Maclaurin approximation, adding terms until the last term < 1e-6.
851. Print the number of 3-digit numbers divisible by all of `3`, `5`, and `7`.
852. Use a loop to simulate growth: starting amount doubles each year. Find year it exceeds 1 million.
853. Print a spiral number matrix `1..n^2` using a simulation loop for `n=4`.
854. Use a for-each to validate that all items in a list match a specific type. Print results.
855. Write a loop to collect even numbers until 5 have been found. Print them.
856. Use a while loop to perform repeated squaring for fast exponentiation. Test `(3, 20)`.
857. Print all numbers from `100` to `200` whose digits are in strictly increasing order.
858. Simulate a random walk: 20 steps, each `+1` or `-1`. Print final position.
859. Write a while loop to drain a queue (list). Pop from front until empty.
860. Use nested for loops to compute the outer product of `[1,2,3]` and `[4,5,6]`.
861. Print the sum of all numbers in a list that appear more than once. Test.
862. Simulate a lottery draw loop until all 6 target numbers are drawn. Count rounds.
863. Use a for loop to implement `zip` for three lists. Test.
864. Write a while loop that guesses a number between 1–10 via random bisection. Count guesses.
865. Print only elements at prime indices from a list using a for loop.
866. Write a for loop that rotates a string 1 char at a time and checks if it matches target.
867. Count the number of local maxima (elements greater than both neighbors) in a list.
868. Print the length of each word in a sentence using a for-each loop.
869. Use while loop to simulate a ball bouncing: height halves each bounce, stop below 0.01.
870. Write a for loop to check if two lists have the same elements in same order. Print result.
871. Print `n` rows where each row has the row number printed that many times (e.g., `"3 3 3"`).
872. Use a while loop with accumulator to find the first N primes. Test `N=10`.
873. Use a for loop to find all numbers in `[1..100]` that are both squares and cubes.
874. Simulate a stack with a list and use a while loop to process 5 push and 3 pop operations.
875. Use a nested loop to check if any two elements in a list sum to a given target.
876. Print all anagram pairs from a list of words using nested loops.
877. Implement a basic Caesar cipher encode loop over a string's characters.
878. Count how many characters in `"Hello, World! 123"` are letters, digits, and symbols.
879. Use a while loop to perform Newton's approximation of cube root. Test `n=27`.
880. Simulate a deck of 10 cards: loop and draw random cards without replacement using remove.

---

## Section 8: Procedures (Problems 881–1000)

881. Write a void function `printGreeting(name)` that prints `"Hello, name!"`. Call it 3 times.
882. Write a result function `double(n) = { n * 2 }`. Test with 5 values.
883. Write a result function `square(n)`. Test `1..10`.
884. Write a result function `cube(n)`. Test `1..5`.
885. Write a result function `negate(n)` using the `-` unary operator. Test.
886. Write `isEven(n) = { n % 2 == 0 }`. Test several values.
887. Write `isOdd(n) = { n % 2 != 0 }`. Test.
888. Write `abs2(n)` without `abs()` using if expression. Test.
889. Write `sign(n)` — returns `1`, `-1`, or `0`. Test.
890. Write `max2(a, b) = { if (a > b) a else b }`. Test.
891. Write `min2(a, b)`. Test.
892. Write `clamp(x, lo, hi)`. Test extreme values.
893. Write `lerp(a, b, t) = { a + (b - a) * t }`. Test `t=0`, `0.5`, `1`.
894. Write `within(x, lo, hi) = { x >= lo && x <= hi }`. Test.
895. Write `roundTo(x, places)` using `round` and `^`. Test `(3.14159, 2)`.
896. Write `factorial(n)` recursively. Test `0..10`.
897. Write `fib(n)` recursively. Test `0..12`.
898. Write `gcd(a, b)` recursively. Test `(24, 36)`.
899. Write `lcm(a, b)` using gcd. Test `(4, 6)`.
900. Write `pow(base, exp)` recursively. Test `(2, 10)`.
901. Write `sumList(list)` using reduce. Test.
902. Write `productList(list)` using reduce.
903. Write `maxList(list)` manually. Test.
904. Write `minList(list)` manually.
905. Write `countIf(list, threshold)` — count elements above threshold. Test.
906. Write `filterEvens(list)` using filter lambda. Test.
907. Write `filterOdds(list)`.
908. Write `squareAll(list)` using map. Test.
909. Write `doubleAll(list)` using map.
910. Write `sumSquares(list)` — sum of squared elements. Test.
911. Write `applyN(fn, x, n)` — apply a named function `n` times to `x`. Simulate with dispatch. Test `("double", 3, 4)`.
912. Write `compose(f, g, x)` — applies g then f. Simulate with named functions. Test.
913. Write `memoFib(n)` using global dict cache. Test `n=40`.
914. Write `accumulate(list, fn)` — running results of applying binary op. Simulate sum. Test.
915. Write `scan(list, init, fn)` — like reduce but returns all intermediate values. Test.
916. Write `iterate(f, x, n)` — list of `n` values applying `f` repeatedly. Simulate. Test.
917. Write `unfold(seed, fn, n)` — generate list by repeatedly applying fn to state. Test.
918. Write `memoize(cache, key, fn)` — check-then-store pattern. Test.
919. Write `retry(fn, maxTries)` — simulate calling until returns nonzero or max tries. Test.
920. Write `trampoline(fn, x)` — repeatedly call function while result is a list `[fn, arg]`. Test.
921. Write `bubbleSort(list)` iteratively. Test `[5,3,8,6,1,9,2,7,4]`.
922. Write `selectionSort(list)`. Test.
923. Write `insertionSort(list)`. Test.
924. Write `mergeSort(list)` recursively. Test.
925. Write `quickSort(list)` recursively using filter partitioning. Test.
926. Write `binarySearch(list, val)`. Test.
927. Write `linearSearch(list, val)`. Test.
928. Write `flatten(list)` recursively handling nested lists. Test.
929. Write `deepReverse(list)` — reverse outer and each inner list. Test.
930. Write `deepMap(list, fn)` — map over possibly nested list. Simulate with doubling. Test.
931. Write `treeSum(node)` using nested dicts `{val, left, right}`. Test small tree.
932. Write `treeDepth(node)` — maximum depth. Test.
933. Write `treeBFS(root)` — breadth-first traversal collecting node values. Test.
934. Write `treeDFS(root)` — depth-first (pre-order) collecting values. Test.
935. Write `buildBST(list)` — insert list elements into binary search tree dict. Test.
936. Write `bstContains(tree, val)`. Test.
937. Write `bstInOrder(tree)` — in-order traversal. Test.
938. Write `countNodes(tree)`. Test.
939. Write `countLeaves(tree)`. Test.
940. Write `pathToLeaf(tree, target)` — return path list. Test.
941. Write `isBST(tree)`. Test.
942. Write `zipWith(a, b, fn)` — element-wise apply function. Simulate add. Test.
943. Write `groupBy(list, keyFn)` — group by computed key. Simulate by even/odd. Test.
944. Write `partition(list, pred)` — split into two lists. Test even/odd split.
945. Write `frequency(list)` — dict of element counts. Test.
946. Write `topK(list, k)` — k most frequent elements. Test.
947. Write `runningSum(list)` — cumulative sums list. Test.
948. Write `longestIncreasing(list)` — length of longest increasing subsequence. Test.
949. Write `maxSubarray(list)` — Kadane's algorithm. Test.
950. Write `rotate(list, k)` — rotate left by k. Test.
951. Write `powerset(list)` — all subsets. Test `[1,2,3]`.
952. Write `permutations(list)` — all permutations. Test `[1,2,3]`.
953. Write `combinations(list, k)`. Test.
954. Write `cartesian(a, b)` — all pairs. Test.
955. Write `nQueens(n)` — count solutions to n-queens using backtracking. Test `n=4`.
956. Write `solveMaze(maze, start, end)` — BFS/DFS on grid. Test small 5x5 grid.
957. Write `knapsack(weights, values, cap)` — 0/1 knapsack DP. Test small example.
958. Write `longestCommonSubsequence(a, b)` — length. Test `("ABCBDAB", "BDCABA")`.
959. Write `editDistance(a, b)` — Levenshtein distance. Test `("kitten", "sitting")`.
960. Write `coinChange(coins, amount)` — min coins. Test `([1,5,10,25], 63)`.
961. Write `wordBreak(s, dict)` — can `s` be segmented into dict words. Test.
962. Write `decodeWays(s)` — number of ways to decode digit string as letters. Test `"12"`, `"226"`.
963. Write `trapRainWater(heights)` — water trapped between bars. Test `[0,1,0,2,1,0,1,3,2,1,2,1]`.
964. Write `maxProduct(list)` — max product of contiguous subarray. Test.
965. Write `jumpGame(list)` — can you reach last index with jump limits. Test `[2,3,1,1,4]`.
966. Write `uniquePaths(m, n)` — count paths in grid from top-left to bottom-right. Test `(3,7)`.
967. Write `climbStairs(n)` — number of ways to climb n stairs 1 or 2 at a time. Test `n=10`.
968. Write `houseRobber(list)` — max sum with no two adjacent elements. Test `[2,7,9,3,1]`.
969. Write `longestPalindrome(s)` — longest palindromic substring length. Test `"babad"`.
970. Write `isValidParentheses(s)` — only `()[]{}`. Test several strings.
971. Write `minStack(ops)` — simulate stack supporting getMin in O(1). Test.
972. Write `lruCache(capacity, ops)` — simulate least recently used cache. Test.
973. Write `encode(list)` and `decode(encoded)` for run-length encoding of lists. Test.
974. Write `serialize(tree)` and `deserialize(s)` for a binary tree. Test.
975. Write `evaluateRPN(tokens)` — evaluate Reverse Polish Notation. Test `["2","3","+","4","*"]`.
976. Write `infixToPostfix(tokens)` — convert infix to postfix notation. Test `["3","+","4","*","2"]`.
977. Write `evaluateExpression(expr)` — evaluate a simple `a OP b` expression string. Test.
978. Write `simplify(fraction)` — given `[numerator, denominator]`, return simplified. Test.
979. Write `addFractions(a, b)` — list fractions, return simplified result. Test.
980. Write `multiplyFractions(a, b)`. Test.
981. Write `matrixPower(m, n)` — raise 2x2 matrix to nth power. Test.
982. Write `fibMatrix(n)` — use matrix exponentiation to compute nth Fibonacci. Test.
983. Write `sieveOfEratosthenes(n)` — return all primes up to n. Test `n=100`.
984. Write `millerRabin(n, k)` — probabilistic primality test simulation. Test several primes.
985. Write `extendedGcd(a, b)` — return `[gcd, x, y]` such that `ax + by = gcd`. Test.
986. Write `modInverse(a, m)` using extended GCD. Test.
987. Write `chineseRemainderTheorem(remainders, moduli)`. Test simple case.
988. Write `convexHull2D(points)` — return hull points in order. Test small point set.
989. Write `closestPair(points)` — find closest two points by Euclidean distance. Test.
990. Write `polyhash(s, base, mod)` — rolling polynomial hash of string. Test.
991. Write `rabinKarp(text, pattern)` — pattern search using hash. Test.
992. Write `kmpSearch(text, pattern)` — KMP string search algorithm. Test.
993. Write `trie operations` — insert, search, startsWith — using nested dicts. Test.
994. Write `unionFind(n)` — union-find with path compression. Test `n=10`.
995. Write `dijkstra(graph, start)` — shortest paths using dict-based graph. Test small graph.
996. Write `topologicalSort(graph)` — Kahn's algorithm. Test DAG.
997. Write `tarjan(graph)` — find strongly connected components. Test.
998. Write `floodFill(grid, x, y, newColor)` — fill connected region. Test on 5x5 grid.
999. Write `gameOfLife(grid)` — compute one generation of Conway's Game of Life. Test 5x5.
1000. Write a complete mini-interpreter: `parse(expr)` that handles `+`, `-`, `*`, `/` with integer operands and prints the result. Test 10 expressions.
