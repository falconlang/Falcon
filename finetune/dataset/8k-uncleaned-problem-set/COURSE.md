# Falcon Language — Algorithmic Course

A comprehensive reference of Falcon (`.mist`) syntax and algorithmic patterns across all core domains. Each section moves from fundamentals to advanced patterns.

---

## Table of Contents

1. [Variables](#1-variables)
2. [Math](#2-math)
3. [Text](#3-text)
4. [Lists](#4-lists)
5. [Dictionaries](#5-dictionaries)
6. [Colors](#6-colors)
7. [Controls](#7-controls)
8. [Procedures](#8-procedures)

---

## 1. Variables

### 1.1 Global Variable Declaration

A global variable is declared at the root scope and accessed with `this.`.

```falcon
global score = 0
global playerName = "Alice"
global isActive = true

println(this.score)
println(this.playerName)
```

### 1.2 Local Variable Declaration

A local variable is scoped to its enclosing block.

```falcon
local age = 25
local pi = 3.14159
local greeting = "Hello"
println(age)
println(pi)
println(greeting)
```

### 1.3 Variable Reassignment

Variables can be reassigned at any time. They are dynamically typed.

```falcon
local x = 10
x = x + 5
println(x)

local label = "Start"
label = "Finish"
println(label)
```

### 1.4 Global Variable Mutation

```falcon
global counter = 0

func increment() {
  this.counter = this.counter + 1
}

increment()
increment()
println(this.counter)
```

### 1.5 Swapping Two Variables

```falcon
local a = 100
local b = 200
local temp = a
a = b
b = temp
println(a)
println(b)
```

### 1.6 Multiple Globals Tracking State

```falcon
global totalSales = 0
global numTransactions = 0
global lastItem = "None"

func recordSale(item, amount) {
  this.totalSales = this.totalSales + amount
  this.numTransactions = this.numTransactions + 1
  this.lastItem = item
}

recordSale("Apple", 3)
recordSale("Banana", 5)
println(this.totalSales)
println(this.numTransactions)
println(this.lastItem)
```

### 1.7 Boolean Variables

```falcon
local isLoggedIn = false
local hasPermission = true

if (isLoggedIn && hasPermission) {
  println("Access granted")
} else {
  println("Access denied")
}
```

### 1.8 Dynamic Typing — Number to Text

```falcon
local val = 42
println(val ? number)

val = "hello"
println(val ? text)
```

### 1.9 Accumulator Pattern

```falcon
local sum = 0
local product = 1

for (i: 1 .. 5) {
  sum = sum + i
  product = product * i
}
println(sum)
println(product)
```

### 1.10 Type Checking with `?`

```falcon
local items = ["apple", 3, true, [1, 2], {"k": "v"}]

for (item in items) {
  if (item ? text) {
    println("text: " _ item)
  } else if (item ? number) {
    println("number: " _ item)
  } else if (item ? list) {
    println("list")
  } else if (item ? dict) {
    println("dict")
  } else {
    println("other")
  }
}
```

---

## 2. Math

### 2.1 Basic Arithmetic

```falcon
local a = 20
local b = 6
println(a + b)
println(a - b)
println(a * b)
println(a / b)
println(a % b)
println(a ^ 2)
```

### 2.2 Square Root and Absolute Value

```falcon
println(sqrt(144))
println(abs(-37))
println(abs(37))
```

### 2.3 Rounding Functions

```falcon
local n = 4.7
println(round(n))
println(floor(n))
println(ceil(n))

local m = 4.2
println(round(m))
println(floor(m))
println(ceil(m))
```

### 2.4 Power and Exponent

```falcon
println(2 ^ 10)
println(exp(1))
println(exp(0))
```

### 2.5 Logarithm

```falcon
println(log(1))
println(log(10))
println(log(100))
```

### 2.6 Trigonometry

```falcon
local angle = 30
local rad = radians(angle)
println(sin(rad))
println(cos(rad))
println(tan(rad))
```

### 2.7 Inverse Trigonometry

```falcon
println(degrees(asin(1)))
println(degrees(acos(0)))
println(degrees(atan(1)))
```

### 2.8 Min and Max of Values

```falcon
println(min(3, 7, 1, 9, 2))
println(max(3, 7, 1, 9, 2))
```

### 2.9 Random Integer

```falcon
local roll = randInt(1, 6)
println("Dice roll: " _ roll)
```

### 2.10 Random Float

```falcon
local chance = randFloat()
if (chance > 0.5) {
  println("Heads")
} else {
  println("Tails")
}
```

### 2.11 Seeded Random

```falcon
setRandSeed(42)
println(randInt(1, 100))
println(randInt(1, 100))
```

### 2.12 Modulus, Remainder, Quotient

```falcon
println(mod(17, 5))
println(rem(17, 5))
println(quot(17, 5))
```

### 2.13 Format Decimal

```falcon
local pi = 3.14159265
println(formatDecimal(pi, 2))
println(formatDecimal(pi, 4))
```

### 2.14 Base Conversion — Decimal to Binary

```falcon
println(decToBin(10))
println(decToBin(255))
println(decToBin(0))
```

### 2.15 Base Conversion — Decimal to Hex

```falcon
println(decToHex(255))
println(decToHex(16))
println(decToHex(4096))
```

### 2.16 Binary to Decimal

```falcon
println(binToDec("1010"))
println(binToDec("11111111"))
println(binToDec("1"))
```

### 2.17 Hex to Decimal

```falcon
println(hexToDec("FF"))
println(hexToDec("1A"))
println(hexToDec("100"))
```

### 2.18 Parse Numeric Literals from Strings

```falcon
println(bin("1010"))
println(hexa("1F"))
println(octal("17"))
println(dec("255"))
```

### 2.19 Statistical Functions on a List

```falcon
local scores = [72, 85, 90, 60, 78, 95, 88]
println(avgOf(scores))
println(maxOf(scores))
println(minOf(scores))
println(stdDevOf(scores))
println(stdErrOf(scores))
println(geoMeanOf(scores))
```

### 2.20 Mode of a List

```falcon
local grades = [85, 90, 85, 72, 90, 85, 78]
println(modeOf(grades))
```

### 2.21 atan2 — Angle Between Two Points

```falcon
local dy = 1.0
local dx = 1.0
local angle = degrees(atan2(dy, dx))
println(angle)
```

### 2.22 Integer Square Root Check

```falcon
func isPerfectSquare(n) = {
  local s = floor(sqrt(n))
  s * s == n
}

println(isPerfectSquare(25))
println(isPerfectSquare(26))
println(isPerfectSquare(144))
```

### 2.23 Greatest Common Divisor (GCD)

```falcon
func gcd(a, b) = {
  if (b == 0) {
    a
  } else {
    gcd(b, mod(a, b))
  }
}

println(gcd(48, 18))
println(gcd(100, 75))
```

### 2.24 Least Common Multiple (LCM)

```falcon
func gcd(a, b) = {
  if (b == 0) {
    a
  } else {
    gcd(b, mod(a, b))
  }
}

func lcm(a, b) = {
  (a * b) / gcd(a, b)
}

println(lcm(4, 6))
println(lcm(12, 18))
```

### 2.25 Power Function (Iterative)

```falcon
func power(base, exp) = {
  local result = 1
  local i = 0
  while (i < exp) {
    result = result * base
    i = i + 1
  }
  result
}

println(power(2, 8))
println(power(3, 4))
```

### 2.26 Factorial

```falcon
func factorial(n) = {
  if (n <= 1) {
    1
  } else {
    n * factorial(n - 1)
  }
}

println(factorial(5))
println(factorial(10))
```

### 2.27 Prime Check

```falcon
func isPrime(n) = {
  if (n < 2) {
    false
  } else {
    local i = 2
    local prime = true
    while (i * i <= n) {
      if (mod(n, i) == 0) {
        prime = false
        break
      }
      i = i + 1
    }
    prime
  }
}

println(isPrime(2))
println(isPrime(17))
println(isPrime(18))
```

### 2.28 Sum of Digits

```falcon
func sumDigits(n) = {
  local s = 0
  local num = abs(n)
  while (num > 0) {
    s = s + mod(num, 10)
    num = floor(num / 10)
  }
  s
}

println(sumDigits(1234))
println(sumDigits(9999))
```

### 2.29 Count Digits

```falcon
func countDigits(n) = {
  if (n == 0) {
    1
  } else {
    local count = 0
    local num = abs(n)
    while (num > 0) {
      count = count + 1
      num = floor(num / 10)
    }
    count
  }
}

println(countDigits(12345))
println(countDigits(0))
```

### 2.30 Celsius to Fahrenheit

```falcon
func celsiusToFahrenheit(c) = {
  (c * 9 / 5) + 32
}

println(celsiusToFahrenheit(0))
println(celsiusToFahrenheit(100))
println(celsiusToFahrenheit(37))
```

---

## 3. Text

### 3.1 String Concatenation with `_`

```falcon
local firstName = "John"
local lastName = "Doe"
local fullName = firstName _ " " _ lastName
println(fullName)
```

### 3.2 Text Length

```falcon
local word = "Elephant"
println(word.textLen())
```

### 3.3 Uppercase and Lowercase

```falcon
local msg = "Hello World"
println(msg.uppercase())
println(msg.lowercase())
```

### 3.4 Trim Whitespace

```falcon
local padded = "   Hello   "
println(padded.trim())
```

### 3.5 Contains

```falcon
local sentence = "The quick brown fox"
println(sentence.contains("quick"))
println(sentence.contains("lazy"))
```

### 3.6 Starts With

```falcon
local url = "https://example.com"
println(url.startsWith("https"))
println(url.startsWith("http://"))
```

### 3.7 Split at Delimiter

```falcon
local csv = "apple,banana,cherry"
local fruits = csv.split(",")
for (fruit in fruits) {
  println(fruit)
}
```

### 3.8 Split at Spaces

```falcon
local sentence = "The quick brown fox"
local words = sentence.splitAtSpaces()
println(words.listLen())
for (word in words) {
  println(word)
}
```

### 3.9 Split at First

```falcon
local data = "name:John Doe"
local parts = data.splitAtFirst(":")
println(parts[1])
println(parts[2])
```

### 3.10 Reverse a String

```falcon
local word = "racecar"
println(word.reverse())

local word2 = "hello"
println(word2.reverse())
```

### 3.11 Replace Substring

```falcon
local msg = "I like cats and cats are cute"
println(msg.replace("cats", "dogs"))
```

### 3.12 Replace from Map

```falcon
local template = "Hello, NAME! You are AGE years old."
local replacements = {"NAME": "Alice", "AGE": "30"}
println(template.replaceFrom(replacements))
```

### 3.13 Replace from Map — Longest First

```falcon
local template = "FIRSTNAME FIRSTNAMESURNAME"
local map = {"FIRSTNAME": "John", "FIRSTNAMESURNAME": "John Doe"}
println(template.replaceFromLongestFirst(map))
```

### 3.14 Segment (Substring)

```falcon
local str = "Hello, World!"
println(str.segment(1, 5))
println(str.segment(8, 5))
```

### 3.15 Text Comparison Operators

```falcon
local a = "apple"
local b = "banana"

println(a === "apple")
println(a !== "apple")
println(a << b)
println(b >> a)
```

### 3.16 ContainsAny

```falcon
local sentence = "The sky is blue today"
local keywords = ["blue", "red", "green"]
println(sentence.containsAny(keywords))
```

### 3.17 ContainsAll

```falcon
local sentence = "The quick brown fox"
local words = ["quick", "fox"]
println(sentence.containsAll(words))
```

### 3.18 SplitAtAny

```falcon
local data = "one;two|three,four"
local parts = data.splitAtAny([";", "|", ","])
for (p in parts) {
  println(p)
}
```

### 3.19 CSV Row to List

```falcon
local row = "Alice,30,Engineer"
local fields = row.csvRowToList()
println(fields[1])
println(fields[2])
println(fields[3])
```

### 3.20 List to CSV Row

```falcon
local record = ["Bob", "25", "Designer"]
println(record.toCsvRow())
```

### 3.21 Count Vowels in a String

```falcon
func countVowels(s) = {
  local lower = s.lowercase()
  local vowels = ["a", "e", "i", "o", "u"]
  local chars = lower.split("")
  chars.filter { c -> vowels.containsItem(c) }.listLen()
}

println(countVowels("Hello World"))
println(countVowels("Elephant"))
```

### 3.22 Palindrome Check

```falcon
func isPalindrome(s) = {
  local lower = s.lowercase()
  lower === lower.reverse()
}

println(isPalindrome("racecar"))
println(isPalindrome("hello"))
println(isPalindrome("Madam"))
```

### 3.23 Word Count

```falcon
func wordCount(sentence) = {
  sentence.trim().splitAtSpaces().listLen()
}

println(wordCount("The quick brown fox"))
println(wordCount("  Hello   World  "))
```

### 3.24 Capitalize First Letter

```falcon
func capitalize(s) = {
  if (s.textLen() == 0) {
    s
  } else {
    s.segment(1, 1).uppercase() _ s.segment(2, s.textLen() - 1)
  }
}

println(capitalize("hello"))
println(capitalize("world"))
```

### 3.25 Repeat String N Times

```falcon
func repeatStr(s, n) = {
  local result = ""
  local i = 0
  while (i < n) {
    result = result _ s
    i = i + 1
  }
  result
}

println(repeatStr("ab", 4))
println(repeatStr("-", 10))
```

### 3.26 Check if String is Numeric

```falcon
func isNumericString(s) = {
  s ? number
}

println(isNumericString("3.14"))
println(isNumericString("hello"))
println(isNumericString("42"))
```

### 3.27 Join Words with Separator

```falcon
local words = ["one", "two", "three", "four"]
println(words.join(", "))
println(words.join(" | "))
println(words.join(""))
```

### 3.28 String Contains Binary

```falcon
local binStr = "1010"
println(binStr ? bin)

local notBin = "1020"
println(notBin ? bin)
```

### 3.29 Truncate String

```falcon
func truncate(s, maxLen, ellipsis) = {
  if (s.textLen() <= maxLen) {
    s
  } else {
    s.segment(1, maxLen) _ ellipsis
  }
}

println(truncate("Hello, World!", 5, "..."))
println(truncate("Hi", 10, "..."))
```

### 3.30 Title Case

```falcon
func titleCase(sentence) = {
  local words = sentence.splitAtSpaces()
  local result = words.map { w ->
    w.segment(1, 1).uppercase() _ w.segment(2, w.textLen() - 1).lowercase()
  }
  result.join(" ")
}

println(titleCase("the quick brown fox"))
println(titleCase("hello world"))
```

---

## 4. Lists

### 4.1 List Creation and Access

```falcon
local fruits = ["apple", "banana", "cherry"]
println(fruits[1])
println(fruits[2])
println(fruits[3])
```

### 4.2 List Length

```falcon
local nums = [10, 20, 30, 40, 50]
println(nums.listLen())
```

### 4.3 Add Elements

```falcon
local colors = ["red", "green"]
colors.add("blue")
colors.add("yellow", "purple")
println(colors)
```

### 4.4 Insert at Index

```falcon
local list = ["a", "c", "d"]
list.insert(2, "b")
println(list)
```

### 4.5 Remove at Index

```falcon
local list = ["a", "b", "c", "d"]
list.remove(2)
println(list)
```

### 4.6 Index Of

```falcon
local list = ["cat", "dog", "bird"]
println(list.indexOf("dog"))
println(list.indexOf("fish"))
```

### 4.7 Contains Item

```falcon
local primes = [2, 3, 5, 7, 11]
println(primes.containsItem(5))
println(primes.containsItem(6))
```

### 4.8 Append Two Lists

```falcon
local a = [1, 2, 3]
local b = [4, 5, 6]
a.appendList(b)
println(a)
```

### 4.9 Reverse a List

```falcon
local list = [1, 2, 3, 4, 5]
println(list.reverseList())
```

### 4.10 Sort a List

```falcon
local nums = [5, 2, 8, 1, 9, 3]
println(nums.sort())

local words = ["banana", "apple", "cherry"]
println(words.sort())
```

### 4.11 Slice a List

```falcon
local list = [10, 20, 30, 40, 50, 60]
println(list.slice(2, 4))
```

### 4.12 Random Element

```falcon
local options = ["rock", "paper", "scissors"]
println(options.random())
```

### 4.13 All But First / All But Last

```falcon
local list = [1, 2, 3, 4, 5]
println(list.allButFirst())
println(list.allButLast())
```

### 4.14 Map Lambda — Double Values

```falcon
local nums = [1, 2, 3, 4, 5]
local doubled = nums.map { n -> n * 2 }
println(doubled)
```

### 4.15 Map Lambda — Square Values

```falcon
local nums = [1, 2, 3, 4, 5]
local squares = nums.map { n -> n ^ 2 }
println(squares)
```

### 4.16 Filter Lambda — Even Numbers

```falcon
local nums = [1, 2, 3, 4, 5, 6, 7, 8]
local evens = nums.filter { n -> n % 2 == 0 }
println(evens)
```

### 4.17 Filter Lambda — Strings Only

```falcon
local mixed = [1, "apple", 2, "banana", 3, "cherry"]
local strings = mixed.filter { x -> x ? text }
println(strings)
```

### 4.18 Reduce Lambda — Sum

```falcon
local nums = [1, 2, 3, 4, 5]
local total = nums.reduce(0) { x, acc -> x + acc }
println(total)
```

### 4.19 Reduce Lambda — Product

```falcon
local nums = [1, 2, 3, 4, 5]
local product = nums.reduce(1) { x, acc -> x * acc }
println(product)
```

### 4.20 Sort Lambda — Descending Numbers

```falcon
local nums = [3, 1, 4, 1, 5, 9, 2, 6]
local desc = nums.sort { a, b -> a > b }
println(desc)
```

### 4.21 Sort Lambda — By String Length

```falcon
local words = ["banana", "fig", "apple", "kiwi"]
local byLen = words.sort { a, b -> a.textLen() < b.textLen() }
println(byLen)
```

### 4.22 Max Lambda — Longest String

```falcon
local words = ["cat", "elephant", "ox", "hippopotamus"]
local longest = words.max { a, b -> b.textLen() > a.textLen() }
println(longest)
```

### 4.23 Min Lambda — Shortest String

```falcon
local words = ["cat", "elephant", "ox", "hippopotamus"]
local shortest = words.min { a, b -> a.textLen() < b.textLen() }
println(shortest)
```

### 4.24 Pairs to Dictionary

```falcon
local pairs = [["name", "Alice"], ["age", 30], ["city", "Paris"]]
local dict = pairs.pairsToDict()
println(dict.get("name", ""))
println(dict.get("city", ""))
```

### 4.25 Lookup in Pairs

```falcon
local pairs = [["red", "#FF0000"], ["green", "#00FF00"], ["blue", "#0000FF"]]
println(pairs.lookupInPairs("green", "not found"))
println(pairs.lookupInPairs("purple", "not found"))
```

### 4.26 Copy List (Independent)

```falcon
local original = [1, 2, 3]
local copy = copyList(original)
copy.add(4)
println(original)
println(copy)
```

### 4.27 Flatten Two Levels

```falcon
func flatten(nested) = {
  nested.reduce([]) { sublist, acc ->
    acc.appendList(copyList(sublist))
    acc
  }
}

local nested = [[1, 2], [3, 4], [5, 6]]
println(flatten(nested))
```

### 4.28 Unique Elements (Deduplicate)

```falcon
func unique(list) = {
  local seen = []
  for (item in list) {
    if (!seen.containsItem(item)) {
      seen.add(item)
    }
  }
  seen
}

println(unique([1, 2, 2, 3, 3, 3, 4]))
println(unique(["a", "b", "a", "c", "b"]))
```

### 4.29 Zip Two Lists

```falcon
func zip(listA, listB) = {
  local result = []
  local len = min(listA.listLen(), listB.listLen())
  local i = 1
  while (i <= len) {
    result.add([listA[i], listB[i]])
    i = i + 1
  }
  result
}

local keys = ["a", "b", "c"]
local vals = [1, 2, 3]
println(zip(keys, vals))
```

### 4.30 Chunk List into Groups

```falcon
func chunk(list, size) = {
  local result = []
  local i = 1
  while (i <= list.listLen()) {
    local end = min(i + size - 1, list.listLen())
    result.add(list.slice(i, end))
    i = i + size
  }
  result
}

println(chunk([1, 2, 3, 4, 5, 6, 7], 3))
```

### 4.31 Sum of List

```falcon
local nums = [10, 20, 30, 40, 50]
local total = nums.reduce(0) { x, acc -> x + acc }
println(total)
```

### 4.32 Average of List

```falcon
local scores = [80, 90, 70, 85, 95]
local avg = avgOf(scores)
println(avg)
```

### 4.33 Find Maximum Without Built-in

```falcon
func findMax(list) = {
  local maxVal = list[1]
  local i = 2
  while (i <= list.listLen()) {
    if (list[i] > maxVal) {
      maxVal = list[i]
    }
    i = i + 1
  }
  maxVal
}

println(findMax([3, 7, 2, 9, 4, 6]))
```

### 4.34 Count Occurrences of a Value

```falcon
func countOccurrences(list, val) = {
  list.filter { x -> x == val }.listLen()
}

println(countOccurrences([1, 2, 2, 3, 2, 4], 2))
println(countOccurrences(["a", "b", "a", "c"], "a"))
```

### 4.35 Rotate List Left by N

```falcon
func rotateLeft(list, n) = {
  local len = list.listLen()
  local shift = mod(n, len)
  local tail = list.slice(shift + 1, len)
  local head = list.slice(1, shift)
  tail.appendList(head)
  tail
}

println(rotateLeft([1, 2, 3, 4, 5], 2))
```

### 4.36 Matrix (List of Lists) — Row Sum

```falcon
local matrix = [[1, 2, 3], [4, 5, 6], [7, 8, 9]]

for (row in matrix) {
  local rowSum = row.reduce(0) { x, acc -> x + acc }
  println(rowSum)
}
```

### 4.37 Frequency Map from List

```falcon
func frequencyMap(list) = {
  local freq = {}
  for (item in list) {
    local key = item _ ""
    if (freq.containsKey(key)) {
      freq.set(key, freq.get(key, 0) + 1)
    } else {
      freq.set(key, 1)
    }
  }
  freq
}

local votes = ["Alice", "Bob", "Alice", "Charlie", "Bob", "Alice"]
println(frequencyMap(votes))
```

### 4.38 Intersection of Two Lists

```falcon
func intersect(listA, listB) = {
  listA.filter { x -> listB.containsItem(x) }
}

println(intersect([1, 2, 3, 4], [2, 4, 6]))
```

### 4.39 Difference of Two Lists

```falcon
func difference(listA, listB) = {
  listA.filter { x -> !listB.containsItem(x) }
}

println(difference([1, 2, 3, 4, 5], [2, 4]))
```

### 4.40 Partition List by Condition

```falcon
func partition(list, predResult) = {
  local yes = list.filter { x -> x % 2 == 0 }
  local no = list.filter { x -> x % 2 != 0 }
  [yes, no]
}

local nums = [1, 2, 3, 4, 5, 6, 7, 8]
local result = partition(nums, true)
println(result[1])
println(result[2])
```

---

## 5. Dictionaries

### 5.1 Create and Access

```falcon
local person = {"name": "Alice", "age": 30, "city": "Paris"}
println(person.get("name", "unknown"))
println(person.get("age", 0))
```

### 5.2 Set a Key

```falcon
local config = {"theme": "dark", "lang": "en"}
config.set("fontSize", 14)
println(config.get("fontSize", 12))
```

### 5.3 Delete a Key

```falcon
local data = {"a": 1, "b": 2, "c": 3}
data.delete("b")
println(data.containsKey("b"))
println(data.dictLen())
```

### 5.4 Dictionary Length

```falcon
local d = {"x": 1, "y": 2, "z": 3}
println(d.dictLen())
```

### 5.5 Contains Key

```falcon
local user = {"name": "Bob", "email": "bob@example.com"}
println(user.containsKey("name"))
println(user.containsKey("phone"))
```

### 5.6 Get Keys and Values

```falcon
local scores = {"Alice": 95, "Bob": 87, "Carol": 92}
local names = scores.keys()
local points = scores.values()
println(names)
println(points)
```

### 5.7 Iterate Over a Dictionary

```falcon
local inventory = {"apples": 10, "bananas": 5, "cherries": 25}

for (item, qty in inventory) {
  println(item _ ": " _ qty)
}
```

### 5.8 Merge Two Dictionaries

```falcon
local base = {"a": 1, "b": 2}
local extra = {"c": 3, "d": 4}
base.mergeInto(extra)
println(base)
```

### 5.9 Copy Dictionary

```falcon
local original = {"x": 10, "y": 20}
local copy = copyDict(original)
copy.set("z", 30)
println(original.dictLen())
println(copy.dictLen())
```

### 5.10 To Pairs

```falcon
local d = {"a": 1, "b": 2, "c": 3}
local pairs = d.toPairs()
for (pair in pairs) {
  println(pair[1] _ " -> " _ pair[2])
}
```

### 5.11 Get at Path (Nested Dictionary)

```falcon
local data = {
  "user": {
    "profile": {
      "name": "Alice",
      "age": 30
    }
  }
}
println(data.getAtPath(["user", "profile", "name"], "not found"))
println(data.getAtPath(["user", "profile", "age"], 0))
```

### 5.12 Set at Path (Nested Dictionary)

```falcon
local data = {"settings": {"theme": "light"}}
data.setAtPath(["settings", "fontSize"], 16)
println(data.getAtPath(["settings", "fontSize"], 12))
```

### 5.13 Walk Tree

```falcon
local tree = {
  "root": {
    "left": {"value": 1},
    "right": {"value": 2}
  }
}
local subtree = tree.walkTree(["root", "left"])
println(subtree)
```

### 5.14 Count Entries by Condition

```falcon
local stock = {"apple": 5, "banana": 0, "cherry": 12, "date": 0, "elderberry": 3}

local inStock = 0
for (item, qty in stock) {
  if (qty > 0) {
    inStock = inStock + 1
  }
}
println(inStock)
```

### 5.15 Invert a Dictionary

```falcon
func invertDict(d) = {
  local inverted = {}
  for (key, value in d) {
    inverted.set(value _ "", key)
  }
  inverted
}

local colorCodes = {"red": "#FF0000", "green": "#00FF00", "blue": "#0000FF"}
println(invertDict(colorCodes))
```

### 5.16 Group List into Dictionary by Category

```falcon
local items = ["apple", "avocado", "banana", "blueberry", "cherry"]

local groups = {}
for (item in items) {
  local key = item.segment(1, 1)
  if (groups.containsKey(key)) {
    groups.get(key, []).add(item)
  } else {
    groups.set(key, [item])
  }
}

for (letter, words in groups) {
  println(letter _ ": " _ words.join(", "))
}
```

### 5.17 Most Frequent Item

```falcon
func mostFrequent(list) = {
  local freq = {}
  for (item in list) {
    local key = item _ ""
    if (freq.containsKey(key)) {
      freq.set(key, freq.get(key, 0) + 1)
    } else {
      freq.set(key, 1)
    }
  }

  local bestKey = ""
  local bestCount = 0
  for (key, count in freq) {
    if (count > bestCount) {
      bestCount = count
      bestKey = key
    }
  }
  bestKey
}

println(mostFrequent(["a", "b", "a", "c", "a", "b"]))
```

### 5.18 Histogram from List

```falcon
local ratings = [3, 5, 5, 4, 3, 2, 5, 4, 5, 3]
local histogram = {}

for (r in ratings) {
  local key = r _ ""
  if (histogram.containsKey(key)) {
    histogram.set(key, histogram.get(key, 0) + 1)
  } else {
    histogram.set(key, 1)
  }
}

for (star, count in histogram) {
  println(star _ " stars: " _ count)
}
```

### 5.19 Dictionary as a Counter

```falcon
global wordCount = {}

func countWord(word) {
  if (this.wordCount.containsKey(word)) {
    this.wordCount.set(word, this.wordCount.get(word, 0) + 1)
  } else {
    this.wordCount.set(word, 1)
  }
}

local words = ["the", "cat", "sat", "on", "the", "mat", "the", "cat"]
for (w in words) {
  countWord(w)
}
println(this.wordCount)
```

### 5.20 Filter Dictionary by Value

```falcon
local prices = {"apple": 1.5, "truffle": 200.0, "banana": 0.5, "saffron": 50.0}
local affordable = {}

for (item, price in prices) {
  if (price < 10) {
    affordable.set(item, price)
  }
}
println(affordable)
```

---

## 6. Colors

### 6.1 Color Literal

```falcon
local red = #FF0000
local green = #00FF00
local blue = #0000FF
println(red)
println(green)
println(blue)
```

### 6.2 Make Color from RGB List

```falcon
local coral = makeColor([255, 127, 80])
println(coral)
```

### 6.3 Split Color into Components

```falcon
local color = makeColor([128, 64, 192])
local components = splitColor(color)
println("Alpha: " _ components[1])
println("Red: " _ components[2])
println("Green: " _ components[3])
println("Blue: " _ components[4])
```

### 6.4 Lighten a Color

```falcon
func lighten(color, amount) = {
  local parts = splitColor(color)
  local r = min(parts[2] + amount, 255)
  local g = min(parts[3] + amount, 255)
  local b = min(parts[4] + amount, 255)
  makeColor([r, g, b])
}

local base = makeColor([100, 100, 100])
println(lighten(base, 50))
```

### 6.5 Darken a Color

```falcon
func darken(color, amount) = {
  local parts = splitColor(color)
  local r = max(parts[2] - amount, 0)
  local g = max(parts[3] - amount, 0)
  local b = max(parts[4] - amount, 0)
  makeColor([r, g, b])
}

local base = makeColor([200, 200, 200])
println(darken(base, 80))
```

### 6.6 Blend Two Colors

```falcon
func blendColors(colorA, colorB, t) = {
  local a = splitColor(colorA)
  local b = splitColor(colorB)
  local r = round(a[2] * (1 - t) + b[2] * t)
  local g = round(a[3] * (1 - t) + b[3] * t)
  local bl = round(a[4] * (1 - t) + b[4] * t)
  makeColor([r, g, bl])
}

local white = makeColor([255, 255, 255])
local black = makeColor([0, 0, 0])
println(blendColors(white, black, 0.5))
```

### 6.7 Grayscale a Color

```falcon
func toGrayscale(color) = {
  local parts = splitColor(color)
  local gray = round((parts[2] + parts[3] + parts[4]) / 3)
  makeColor([gray, gray, gray])
}

local vibrant = makeColor([200, 50, 100])
println(toGrayscale(vibrant))
```

### 6.8 Invert a Color

```falcon
func invertColor(color) = {
  local parts = splitColor(color)
  makeColor([255 - parts[2], 255 - parts[3], 255 - parts[4]])
}

local original = makeColor([100, 150, 200])
println(invertColor(original))
```

### 6.9 Color Palette Generation

```falcon
func colorPalette(baseColor, steps) = {
  local palette = []
  local i = 0
  while (i < steps) {
    local factor = i / (steps - 1)
    local parts = splitColor(baseColor)
    local r = round(parts[2] * (1 - factor))
    local g = round(parts[3] * (1 - factor))
    local b = round(parts[4] * (1 - factor))
    palette.add(makeColor([r, g, b]))
    i = i + 1
  }
  palette
}

local base = makeColor([255, 100, 50])
println(colorPalette(base, 5))
```

### 6.10 RGB to Hex String

```falcon
func rgbToHex(r, g, b) = {
  decToHex(r) _ decToHex(g) _ decToHex(b)
}

println(rgbToHex(255, 128, 0))
println(rgbToHex(0, 255, 0))
```

### 6.11 Brightness of a Color (Luminance)

```falcon
func brightness(color) = {
  local parts = splitColor(color)
  (parts[2] * 299 + parts[3] * 587 + parts[4] * 114) / 1000
}

local light = makeColor([240, 240, 240])
local dark = makeColor([30, 30, 30])
println(brightness(light))
println(brightness(dark))
```

### 6.12 Is Color Light or Dark

```falcon
func isLightColor(color) = {
  local parts = splitColor(color)
  local lum = (parts[2] * 299 + parts[3] * 587 + parts[4] * 114) / 1000
  lum > 128
}

println(isLightColor(makeColor([255, 255, 255])))
println(isLightColor(makeColor([0, 0, 0])))
println(isLightColor(makeColor([200, 100, 50])))
```

---

## 7. Controls

### 7.1 If-Else Statement

```falcon
local temp = 35

if (temp > 30) {
  println("Hot day")
} else if (temp > 20) {
  println("Warm day")
} else {
  println("Cool day")
}
```

### 7.2 If-Else as Expression

```falcon
local score = 72
local grade = if (score >= 90) "A" else if (score >= 80) "B" else if (score >= 70) "C" else "F"
println(grade)
```

### 7.3 While Loop with Counter

```falcon
local i = 1
while (i <= 5) {
  println(i)
  i = i + 1
}
```

### 7.4 While Loop with Break

```falcon
local n = 100
local i = 2
while (i * i <= n) {
  if (mod(n, i) == 0) {
    println("Not prime: divisible by " _ i)
    break
  }
  i = i + 1
}
```

### 7.5 For Range Loop

```falcon
for (i: 1 .. 10) {
  println(i)
}
```

### 7.6 For Range with Step

```falcon
for (i: 0 .. 20 step 5) {
  println(i)
}
```

### 7.7 For Range Backwards

```falcon
for (i: 10 .. 1 step -1) {
  println(i)
}
```

### 7.8 For Each Over List

```falcon
local animals = ["Lion", "Tiger", "Bear"]
for (animal in animals) {
  println("Animal: " _ animal)
}
```

### 7.9 For Each Over Dictionary

```falcon
local capitals = {"France": "Paris", "Japan": "Tokyo", "Brazil": "Brasília"}
for (country, capital in capitals) {
  println(country _ " -> " _ capital)
}
```

### 7.10 Nested Loops — Multiplication Table

```falcon
for (i: 1 .. 5) {
  local row = ""
  for (j: 1 .. 5) {
    row = row _ (i * j) _ "\t"
  }
  println(row)
}
```

### 7.11 Loop with Early Break — Search

```falcon
local items = [3, 7, 12, 5, 9, 2]
local target = 5
local found = false
local foundAt = -1

local i = 1
while (i <= items.listLen()) {
  if (items[i] == target) {
    found = true
    foundAt = i
    break
  }
  i = i + 1
}

if (found) {
  println("Found at index " _ foundAt)
} else {
  println("Not found")
}
```

### 7.12 Counting Loop — Collect Evens

```falcon
local evens = []
for (i: 1 .. 20) {
  if (i % 2 == 0) {
    evens.add(i)
  }
}
println(evens)
```

### 7.13 While — Digit Sum Loop

```falcon
local n = 9875
local sum = 0
while (n > 0) {
  sum = sum + mod(n, 10)
  n = floor(n / 10)
}
println(sum)
```

### 7.14 FizzBuzz

```falcon
for (i: 1 .. 30) {
  if (i % 15 == 0) {
    println("FizzBuzz")
  } else if (i % 3 == 0) {
    println("Fizz")
  } else if (i % 5 == 0) {
    println("Buzz")
  } else {
    println(i)
  }
}
```

### 7.15 Collatz Sequence

```falcon
local n = 27
local steps = 0
while (n != 1) {
  if (n % 2 == 0) {
    n = n / 2
  } else {
    n = 3 * n + 1
  }
  steps = steps + 1
}
println("Steps: " _ steps)
```

### 7.16 Binary Search

```falcon
func binarySearch(list, target) = {
  local lo = 1
  local hi = list.listLen()
  local result = -1
  while (lo <= hi) {
    local mid = floor((lo + hi) / 2)
    if (list[mid] == target) {
      result = mid
      break
    } else if (list[mid] < target) {
      lo = mid + 1
    } else {
      hi = mid - 1
    }
  }
  result
}

local sorted = [2, 5, 8, 12, 16, 23, 38, 56, 72, 91]
println(binarySearch(sorted, 23))
println(binarySearch(sorted, 50))
```

### 7.17 Nested Conditionals — Grade Letter

```falcon
func letterGrade(score) = {
  if (score >= 90) {
    "A"
  } else if (score >= 80) {
    "B"
  } else if (score >= 70) {
    "C"
  } else if (score >= 60) {
    "D"
  } else {
    "F"
  }
}

local testScores = [95, 83, 71, 65, 58]
for (s in testScores) {
  println(s _ " -> " _ letterGrade(s))
}
```

### 7.18 Loop — Build Fibonacci Sequence

```falcon
local fib = [0, 1]
local i = 3
while (i <= 15) {
  local next = fib[fib.listLen() - 1] + fib[fib.listLen()]
  fib.add(next)
  i = i + 1
}
println(fib)
```

### 7.19 Loop — Find All Primes up to N (Sieve-like)

```falcon
func primesUpTo(n) = {
  local primes = []
  for (candidate: 2 .. n) {
    local isPrime = true
    local j = 0
    while (j < primes.listLen()) {
      j = j + 1
      if (primes[j] * primes[j] > candidate) {
        break
      }
      if (mod(candidate, primes[j]) == 0) {
        isPrime = false
        break
      }
    }
    if (isPrime) {
      primes.add(candidate)
    }
  }
  primes
}

println(primesUpTo(50))
```

### 7.20 Countdown Timer Simulation

```falcon
local count = 10
while (count >= 0) {
  if (count == 0) {
    println("Liftoff!")
  } else {
    println(count _ "...")
  }
  count = count - 1
}
```

---

## 8. Procedures

### 8.1 Void Function — No Return

```falcon
func greet(name) {
  println("Hello, " _ name _ "!")
}

greet("Alice")
greet("Bob")
```

### 8.2 Result Function — Single Expression

```falcon
func square(n) = { n * n }

println(square(5))
println(square(12))
```

### 8.3 Result Function — Multi-Expression Body

```falcon
func clamp(value, low, high) = {
  if (value < low) {
    low
  } else if (value > high) {
    high
  } else {
    value
  }
}

println(clamp(5, 1, 10))
println(clamp(-3, 1, 10))
println(clamp(15, 1, 10))
```

### 8.4 Recursive — Fibonacci

```falcon
func fib(n) = {
  if (n < 2) {
    n
  } else {
    fib(n - 1) + fib(n - 2)
  }
}

println(fib(0))
println(fib(1))
println(fib(10))
```

### 8.5 Recursive — Factorial

```falcon
func factorial(n) = {
  if (n <= 1) {
    1
  } else {
    n * factorial(n - 1)
  }
}

println(factorial(6))
println(factorial(10))
```

### 8.6 Recursive — Sum of List

```falcon
func sumList(list, idx) = {
  if (idx > list.listLen()) {
    0
  } else {
    list[idx] + sumList(list, idx + 1)
  }
}

println(sumList([1, 2, 3, 4, 5], 1))
```

### 8.7 Recursive — Reverse a String

```falcon
func reverseStr(s) = {
  if (s.textLen() <= 1) {
    s
  } else {
    reverseStr(s.segment(2, s.textLen() - 1)) _ s.segment(1, 1)
  }
}

println(reverseStr("hello"))
println(reverseStr("racecar"))
```

### 8.8 Recursive — Power

```falcon
func pow(base, exp) = {
  if (exp == 0) {
    1
  } else {
    base * pow(base, exp - 1)
  }
}

println(pow(2, 10))
println(pow(3, 5))
```

### 8.9 Recursive — GCD

```falcon
func gcd(a, b) = {
  if (b == 0) {
    a
  } else {
    gcd(b, mod(a, b))
  }
}

println(gcd(56, 98))
println(gcd(100, 75))
```

### 8.10 Recursive — Flatten Nested List

```falcon
func flattenList(list) = {
  local result = []
  for (item in list) {
    if (item ? list) {
      result.appendList(flattenList(item))
    } else {
      result.add(item)
    }
  }
  result
}

println(flattenList([1, [2, [3, 4]], [5, 6], 7]))
```

### 8.11 Higher-Order Pattern — Apply Twice

```falcon
func double(x) = { x * 2 }
func addOne(x) = { x + 1 }

func applyTwice(f, x) = { f(f(x)) }

// Simulate: pass function name as string and dispatch
func applyByName(name, x) = {
  if (name === "double") {
    double(double(x))
  } else if (name === "addOne") {
    addOne(addOne(x))
  } else {
    x
  }
}

println(applyByName("double", 3))
println(applyByName("addOne", 10))
```

### 8.12 Memoization Pattern (Manual Cache)

```falcon
global fibCache = {}

func fibMemo(n) = {
  local key = n _ ""
  if (this.fibCache.containsKey(key)) {
    this.fibCache.get(key, 0)
  } else {
    local result = if (n < 2) n else fibMemo(n - 1) + fibMemo(n - 2)
    this.fibCache.set(key, result)
    result
  }
}

println(fibMemo(30))
println(fibMemo(35))
```

### 8.13 Pipeline Pattern

```falcon
func pipeline(value, steps) = {
  local result = value
  for (step in steps) {
    if (step === "double") {
      result = result * 2
    } else if (step === "increment") {
      result = result + 1
    } else if (step === "square") {
      result = result ^ 2
    } else if (step === "negate") {
      result = neg(result)
    }
  }
  result
}

println(pipeline(3, ["double", "increment", "square"]))
println(pipeline(5, ["square", "negate"]))
```

### 8.14 Recursive — Digit Count

```falcon
func numDigits(n) = {
  if (n < 10) {
    1
  } else {
    1 + numDigits(floor(n / 10))
  }
}

println(numDigits(1))
println(numDigits(12345))
println(numDigits(1000000))
```

### 8.15 Accumulator via Global State

```falcon
global log = []

func logMessage(msg) {
  this.log.add(msg)
}

func getLog() = {
  this.log
}

logMessage("System started")
logMessage("User logged in")
logMessage("Data loaded")
println(getLog())
```

### 8.16 Currying-Style with Dispatch

```falcon
func adder(n, x) = { n + x }
func multiplier(n, x) = { n * x }

func applyOp(op, n, x) = {
  if (op === "add") {
    adder(n, x)
  } else if (op === "multiply") {
    multiplier(n, x)
  } else {
    x
  }
}

println(applyOp("add", 5, 3))
println(applyOp("multiply", 4, 7))
```

### 8.17 Recursive — Count Down and Collect

```falcon
func countDownList(n) = {
  if (n <= 0) {
    []
  } else {
    local rest = countDownList(n - 1)
    rest.insert(1, n)
    rest
  }
}

println(countDownList(5))
```

### 8.18 Bubble Sort

```falcon
func bubbleSort(list) = {
  local arr = copyList(list)
  local n = arr.listLen()
  local i = 1
  while (i <= n) {
    local j = 1
    while (j <= n - i) {
      if (arr[j] > arr[j + 1]) {
        local tmp = arr[j]
        arr[j] = arr[j + 1]
        arr[j + 1] = tmp
      }
      j = j + 1
    }
    i = i + 1
  }
  arr
}

println(bubbleSort([64, 34, 25, 12, 22, 11, 90]))
```

### 8.19 Selection Sort

```falcon
func selectionSort(list) = {
  local arr = copyList(list)
  local n = arr.listLen()
  local i = 1
  while (i <= n - 1) {
    local minIdx = i
    local j = i + 1
    while (j <= n) {
      if (arr[j] < arr[minIdx]) {
        minIdx = j
      }
      j = j + 1
    }
    local tmp = arr[i]
    arr[i] = arr[minIdx]
    arr[minIdx] = tmp
    i = i + 1
  }
  arr
}

println(selectionSort([29, 10, 14, 37, 13]))
```

### 8.20 Recursive — Merge Sort

```falcon
func merge(left, right) = {
  local result = []
  local i = 1
  local j = 1
  while (i <= left.listLen() && j <= right.listLen()) {
    if (left[i] <= right[j]) {
      result.add(left[i])
      i = i + 1
    } else {
      result.add(right[j])
      j = j + 1
    }
  }
  while (i <= left.listLen()) {
    result.add(left[i])
    i = i + 1
  }
  while (j <= right.listLen()) {
    result.add(right[j])
    j = j + 1
  }
  result
}

func mergeSort(list) = {
  local n = list.listLen()
  if (n <= 1) {
    list
  } else {
    local mid = floor(n / 2)
    local left = mergeSort(list.slice(1, mid))
    local right = mergeSort(list.slice(mid + 1, n))
    merge(left, right)
  }
}

println(mergeSort([38, 27, 43, 3, 9, 82, 10]))
```

### 8.21 Caesar Cipher Encode

```falcon
func caesarEncode(text, shift) = {
  local chars = text.split("")
  local encoded = chars.map { c ->
    local code = c.segment(1, 1)
    if (code >= "a" && code <= "z") {
      local base = 97
      local shifted = mod(c.segment(1, 1).lowercase() _ "" + 0 - base + shift, 26) + base
      shifted _ ""
    } else {
      c
    }
  }
  encoded.join("")
}
```

### 8.22 Function Returning Multiple Values via List

```falcon
func minMax(list) = {
  local mn = list[1]
  local mx = list[1]
  for (x in list) {
    if (x < mn) {
      mn = x
    }
    if (x > mx) {
      mx = x
    }
  }
  [mn, mx]
}

local result = minMax([4, 2, 9, 1, 7, 3])
println("Min: " _ result[1])
println("Max: " _ result[2])
```

### 8.23 Function Returning Dictionary as Named Result

```falcon
func describe(list) = {
  local result = {}
  result.set("length", list.listLen())
  result.set("sum", list.reduce(0) { x, acc -> x + acc })
  result.set("average", avgOf(list))
  result.set("min", minOf(list))
  result.set("max", maxOf(list))
  result
}

local stats = describe([10, 20, 30, 40, 50])
println(stats.get("sum", 0))
println(stats.get("average", 0))
println(stats.get("min", 0))
println(stats.get("max", 0))
```

### 8.24 Recursive — Binary to Decimal

```falcon
func binToDec2(s, idx) = {
  if (idx > s.textLen()) {
    0
  } else {
    local bit = if (s.segment(idx, 1) === "1") 1 else 0
    bit * round(2 ^ (s.textLen() - idx)) + binToDec2(s, idx + 1)
  }
}

println(binToDec2("1010", 1))
println(binToDec2("11111111", 1))
```

### 8.25 Tail-Recursive Sum (Simulated)

```falcon
func sumTail(list, idx, acc) = {
  if (idx > list.listLen()) {
    acc
  } else {
    sumTail(list, idx + 1, acc + list[idx])
  }
}

println(sumTail([1, 2, 3, 4, 5, 6, 7, 8, 9, 10], 1, 0))
```

### 8.26 String Parsing — Extract Numbers from Text

```falcon
func extractNumbers(sentence) = {
  local words = sentence.splitAtSpaces()
  words.filter { w -> w ? number }.map { w -> w + 0 }
}

println(extractNumbers("I have 3 cats and 2 dogs and 10 fish"))
```

### 8.27 Running Average

```falcon
global runningSum = 0
global runningCount = 0

func addDataPoint(value) {
  this.runningSum = this.runningSum + value
  this.runningCount = this.runningCount + 1
}

func currentAverage() = {
  if (this.runningCount == 0) {
    0
  } else {
    this.runningSum / this.runningCount
  }
}

addDataPoint(10)
addDataPoint(20)
addDataPoint(30)
println(currentAverage())
addDataPoint(40)
println(currentAverage())
```

### 8.28 Map of Functions (Strategy Pattern)

```falcon
func applyStrategy(strategyName, list) = {
  if (strategyName === "sum") {
    list.reduce(0) { x, acc -> x + acc }
  } else if (strategyName === "product") {
    list.reduce(1) { x, acc -> x * acc }
  } else if (strategyName === "max") {
    maxOf(list)
  } else if (strategyName === "min") {
    minOf(list)
  } else {
    0
  }
}

local nums = [1, 2, 3, 4, 5]
println(applyStrategy("sum", nums))
println(applyStrategy("product", nums))
println(applyStrategy("max", nums))
```

### 8.29 Chained List Transformation Pipeline

```falcon
global salesData = [9, "N/A", 12, 15, "N/A", 8, 20, 5, "N/A", 11]

func totalRevenue(pricePerUnit) = {
  this.salesData
    .filter { n -> n ? number }
    .map { n -> n * pricePerUnit }
    .reduce(0) { x, acc -> x + acc }
}

println("Revenue at $3/unit: " _ totalRevenue(3))
println("Revenue at $5/unit: " _ totalRevenue(5))
```

### 8.30 Recursive Tree Traversal (Nested Dicts)

```falcon
func treeSum(node) = {
  if (!(node ? dict)) {
    node
  } else {
    local left = if (node.containsKey("left")) treeSum(node.get("left", 0)) else 0
    local right = if (node.containsKey("right")) treeSum(node.get("right", 0)) else 0
    local val = if (node.containsKey("val")) node.get("val", 0) else 0
    val + left + right
  }
}

local tree = {
  "val": 1,
  "left": {
    "val": 2,
    "left": {"val": 4},
    "right": {"val": 5}
  },
  "right": {
    "val": 3,
    "right": {"val": 6}
  }
}

println(treeSum(tree))
```

---

## Quick Reference

### Operator Summary

| Operator | Meaning |
|----------|---------|
| `+` `-` `*` `/` `%` `^` | Arithmetic |
| `&&` `\|\|` `!` | Logical |
| `&` `\|` `~` | Bitwise AND / OR / XOR |
| `==` `!=` | Equality (numeric/boolean) |
| `===` `!==` | Text equality |
| `<` `<=` `>` `>=` | Relational |
| `<<` `>>` | Text less / greater than |
| `_` | Text join |
| `?` | Type check |
| `:` | Pair (key: value) |

### Variable Scope

| Declaration | Scope | Access |
|-------------|-------|--------|
| `global x = ...` | Root / entire program | `this.x` |
| `local x = ...` | Enclosing block | `x` |

### Function Types

| Form | Returns? |
|------|----------|
| `func foo(a, b) { ... }` | No (void) |
| `func foo(a) = { expr }` | Yes (last expression) |

### List Lambdas

| Lambda | Purpose |
|--------|---------|
| `.map { x -> expr }` | Transform each element |
| `.filter { x -> bool }` | Keep matching elements |
| `.reduce(init) { x, acc -> expr }` | Collapse to single value |
| `.sort { a, b -> bool_a_before_b }` | Custom sort |
| `.min { a, b -> bool }` | Minimum element |
| `.max { a, b -> bool }` | Maximum element |

### Type Check (`?`)

| Expression | Checks |
|------------|--------|
| `x ? text` | Is string |
| `x ? number` | Is number |
| `x ? list` | Is list |
| `x ? dict` | Is dictionary |
| `x ? bin` | Is binary string |
| `x ? emptyList` | Is empty list |
| `x ? emptyText` | Is empty string |
