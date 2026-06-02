![image|690x338](https://community.appinventor.mit.edu/uploads/default/original/3X/e/d/ed3b9d22ddaefffb4fd5ab71964b8816c56c63a1.png)

Falcon is a language designed for App Inventor to enable syntax-based programming and for incorporating agentic coding abilities.

## Quirks
1. Falcon follows 1-based indexing.
2. Falcon variables are dynamically typed. Do not declare variables.
3. Lists and dictionaries are passed as references.
4. Falcon follows Kotlin's style of functional expressions.
5. Falcon does not have a traditional return statement. In result functions, the last expression is returned. Use `yield` inside a result function for an early return.
6. Falcon does NOT have a try-catch or a throw statement.
7. Only single-line comments using double slash `//` are supported.
8. Do not use `_` in place of unused variables
9. Variables can never be uninitialized.
10. Always keep the last expression for returning functions.
11. Always generate codes between ``` and code tag falcon.
12. If the contents of two strings are numeric, then they can be numerically operated on, e.g. `"2" + "3.14"` is a valid code.

## Data types
1. String `"Hello, world!"`
2. Boolean `true` and `false`
3. Number `123` and `3.14`
4. List `[1, 2, 3, 4]`
5. Dictionary `{"Animal": "Tiger", "Scientific Name": "Panthera tigris"}`
6. Colour `#FFFFFF`
7. Matrix `matrix[[1, 2], [3, 4]]`

## Operators

1. Arithmetic: `+`, `-`, `*`, `/`, `%` (remainder), `^` (power)
2. Logical: `&&`, and `||`
3. Bitwise: `&`, `|`, `~` (xor)
4. Equality: `==`, and `!=`
5. Relational: `<`, `<=`, `>`, and `>=`
6. Text lexicographic: `===` (text equals), `!==` (text not equals), `<<` (text less than), `>>` (text greater than)
7. Unary: `!` (not), and `-` (negate)
8. Join: `"Hello " _ "World!"`
9. Pair: `"Fruit": "Mango"`
10. Question (`?`):
    - To check a value for a specific type (`text`, `number`, `list`, `dict`, `matrix`)
        - E.g.,"Hello" ? text`
    - Check for a number type (`number`, `base10`, `hexa`, `bin`)
        - E.g. `"1010" ? bin` is a true expression.
    - Check for empty text (`emptyText`) or an empty list (`emptyList`)
        - E.g. `[] ? emptyList` or `"Cat" ? emptyText`
11. Matrix arithmetic: `[+]`, `[-]`, `[*]`, `[^]`
    - `[+]` and `[-]` perform element-wise matrix addition and subtraction.
    - `[*]` performs matrix multiplication, or scalar multiplication when one operand is a number.
    - `[^]` raises a square matrix to a non-negative integer power.

### Operator precedence

Precedence controls which operators group first. Higher-precedence operators are parsed before lower-precedence operators. For example, `3.14 * r ^ 2` is parsed as `3.14 * (r ^ 2)`, because power has higher precedence than multiplication.

The table below is ordered from lowest precedence to highest precedence.

| Level | Operators | Meaning | Associativity |
| --- | --- | --- | --- |
| 1 | `=` | Assignment | Special |
| 2 | `:` | Pair creation | Left |
| 3 | `_` | Text join | Left |
| 4 | <code>||</code> | Logical OR | Left |
| 5 | `&&` | Logical AND | Left |
| 6 | <code>|</code> | Bitwise OR | Left |
| 7 | `&` | Bitwise AND | Left |
| 8 | `~` | Bitwise XOR | Left |
| 9 | `==`, `!=`, `===`, `!==` | Equality and text equality | Left |
| 10 | `<`, `<=`, `>`, `>=`, `<<`, `>>` | Numeric and text comparison | Left |
| 11 | `+`, `-`, `[+]`, `[-]` | Addition, subtraction, matrix addition, matrix subtraction | Left |
| 12 | `*`, `/`, `%`, `[*]` | Multiplication, division, remainder, matrix multiplication | Left |
| 13 | unary `-`, `!` | Numeric negation and logical NOT | Prefix |
| 14 | `^`, `[^]` | Numeric power and matrix power | Right |

Power follows the usual mathematical convention:

```
2 ^ 3 ^ 2 == 2 ^ (3 ^ 2) == 512
(2 ^ 3) ^ 2 == 64
```

Unary negation has lower precedence than power, so:

```
-2 ^ 2 == -(2 ^ 2) == -4
(-2) ^ 2 == 4
```

Matrix power follows the same precedence and associativity rule as numeric power:

```
A [*] B [^] 2 == A [*] (B [^] 2)
M [^] 2 ^ 3 == M [^] (2 ^ 3)
```


## Variables

### Global variable

A global variable is always declared at the root:
```
global name = "Kumaraswamy B G"

// access the global variable
println(this.name)
```

### Local variable

```
local age = 17

// access the local variable
println(age)
```

## If else

If-else can be a statement or an expression depending on the context.

```
local x = 8
local y = 12

if (x > y) {
  println("X is greater")
} else if (y > x) {
  println("Y is greater")
} else {
  println("They both are equal!")
}
```

Used an expression:

```
println(  if (x > y) "X is greater" else if  (y > x) "Y is greater" else "They both are equal!"  )
```


## While loop

```
local x = 8

while (true) {
  x = x + 1
  if (x == 5) {
    break  
  }
}
```

## For n loop

```
for (i: 1 .. 10 step 2) {
  println(i)
}
```

The `step` clause is optional and defaults to 1.


## Each loop

To iterate over a list:

```
local names = ["India", "Japan", "Russia", "Germany"]

for (country in names) {
  println(country)
}
```

Or over a dictionary:

```
local animalInfo = { "Animal": "Tiger", "Scientific Name": "Panthera tigris" }

for (key, value in animalInfo) {
  println(key _ " : " _ value) // e.g prints  "Animal: Tiger" to the console
}
```


## Functions

Functions are declared using the `func` keyword.

### Void function


```
func fooBar(x, y) {
  println(x + y)
}
```

### Result function

Use the `=` symbol followed by an expression between curly braces.

```
func double(n) = { n * 2 }
```

For a single expression, the braces are optional:

```
func double(n) = n * 2
```

Or multiple expressions:

```
func Fib(n) = {
  if (n < 2) {
    n  
  } else {
    Fib(n - 1) + Fib(n - 2)
  }
}
```
Note that there is no `return` statement in Falcon. The last expression in a body is taken as the output of an expression.

### Anonymous functions

Anonymous functions are procedure values. They can be assigned to variables, passed around, called directly with `()`, or called dynamically with `.call(inputList)`.

Void anonymous function:

```
local greet = func(x) {
  println("Hello " _ x _ "!")
}

greet("Melon")
greet.call(["Melon"])
```

Result anonymous function:

```
local circleArea = func(r) = 3.14 * r ^ 2

println(circleArea(5))
println(circleArea.call([5]))
```

You can also use a braced result body when the function has multiple expressions or needs `yield`:

```
local clampPositive = func(n) = {
  if (n < 0) {
    yield 0
  }
  n
}
```

Anonymous functions capture lexical variables:

```
local makeAdder = func(x) = {
  local add = func(y) = x + y
  add
}

local addFive = makeAdder(5)
println(addFive(7))  // Output: 12
```

Helper calls:

```
println(greet.numArgs())

func sayHello(x) {
  println("Hello " _ x _ "!")
}

local byName = getFunc("sayHello")
local byDropdown = func.sayHello

byName("Ada")
byDropdown.call(["Grace"])
```

### yield

`yield <expr>` exits a result function early and returns the given value to the caller. It only works inside result functions (`= { ... }`). If no `yield` is reached, the last expression in the body is returned as usual.

```
func first_divisible(n, list) = {
  for (i: 1..list.listLen()) {
    if (list[i] % n == 0) {
      yield list[i]
    }
  }
  -1
}

println(first_divisible(7, [3, 10, 21, 44]))  // Output: 21
println(first_divisible(7, [1, 2, 3]))         // Output: -1
```

A common pattern is using `yield` as a guard clause:

```
func safe_div(a, b) = {
  if (b == 0) {
    yield 0
  }
  a / b
}
```

## Functions


### Math

-  `dec(string)`, `bin(string)`, `octal(string)`, `hexa(string)`
  <br>Parse a static constant string from the respective base.
  e.g. `bin("1010")`

- `sqrt(number)`
- `abs(number)`
- `neg(number)`
- `log(number)`
- `exp(number)`
- `round(number)`
- `ceil(number)`
- `floor(number)`
- `sin(number)`
- `cos(number)`
- `tan(number)`
- `asin(number)`
- `acos(number)`
- `atan(number)`
- `degrees(number)`
- `radians(number)`
- `decToHex(number)`
- `decToBin(number)`
- `hexToDec(number)`
- `binToDec(number)`

- `randInt(from, to)`
- `randFloat()`
- `setRandSeed(number)` sets the random generator seed
- `min(...)` and `max(...)`
- `avgOf(list)`, `maxOf(list)`, `minOf(list)`, `geoMeanOf()`, `stdDevOf()`, `stdErrOf()`
- `modeOf(list)`
- `mod(x, y)`, `rem(x, y)`, `quot(x, y)` for modulus, remainder and quotient
- `atan2(a, b)`
- `formatDecimal(number, places)`

### Control

- `println(any)`
- `openScreen(name)` opens an App Inventor screen
- `openScreenWithValue()` opens App Inventor screen with a value
- `closeScreenWithValue()` closes the screen with a val
- `getStartValue()` returns start value of the App
- `closeSceen()` closes current App Inventor screen
- `closeApp()` closes the Android App
- `getPlainStartText()` returns plain start text of the App

### Values

- `copyList(list)`
- `copyDict(dict)`
- `makeNdArray(dimensions list, initial number)`
- `makeColor(rgb list)`
- `splitColor(number)`

## Methods

e.g. `"Hello  ".trim()`

### Text

- `textLen()`
- `trim()`
- `uppercase()`
- `lowercase()`
- `startsAt(piece)` → Int (1-based position of first occurrence, or 0 if not found)
- `contains(piece)`
- `containsAny(word list)`
- `containsAll(word list)`
- `split(at)`
- `splitAtFirst(at)`
- `splitAtAny(word list)`
- `splitAtFirstOfAny(word list)`
- `splitAtSpaces()`
- `reverse()`
- `csvRowToList()`
- `csvTableToList()`
- `segment(from number, length number)`
- `replace(target, replacement)`
- `replaceFrom(map dictionary)`
- `replaceFromLongestFirst(map dictionary)`

### List

- `listLen()`
- `add(any...)`
- `containsItem(any)`
- `indexOf(any)`
- `insert(at_index, any)`
- `remove(at_index)`
- `appendList(another list)`
- `lookupInPairs(key, notfound)`
- `join(text separator)`
- `slice(index1, index2)`
- `random()`
- `reverseList()`
- `toCsvRow()`
- `toCsvTable()`
- `sort()`
- `allButFirst()`
- `allButLast()`
- `pairsToDict()`

### Dictionary

- `dictLen()`
- `get(key, notFound)`
- `set(key, value)`
- `delete(key)`
- `getAtPath(path_list, notfound)`
- `setAtPath(path_list, value)`
- `containsKey(key)`
- `mergeInto(another_dict)`
- `walkTree(path)`
- `keys()`
- `values()`
- `toPairs()`

### Matrix

- `row(row number)`
- `col(column number)`
- `dimension()`
- `inverse()`
- `transpose()`
- `rotateLeft()`
- `rotateRight()`

## List access

```
local numbers = [1, 2, 4]
// access second element (1 based indexing)
println(numbers[2])
// change the first element
numbers[1] = 8
```

## Dictionary access

```
local animalInfo = { "Animal": "Tiger", "Scientific Name": "Panthera tigris" }
// Get a value by key
println(animalInfo.get("Scientific Name", "Not found"))
```

## Matrix access

Matrices compile to App Inventor Matrices blocks. They use rectangular numeric lists, and Falcon follows 1-based indexing for matrix cells, rows, and columns.

```falcon
local a = matrix[[1, 2], [3, 4]]
local b = matrix[[5, 6], [7, 8]]

// Get and set cells with double square brackets.
println(a[[2, 1]])  // Output: 3
a[[1, 2]] = 9

// Create an N-dimensional matrix filled with an initial number.
local zeros = makeNdArray([2, 3], 0)

// Matrix arithmetic.
println(a [+] b)  // Element-wise addition
println(b [-] a)  // Element-wise subtraction
println(a [*] b)  // Matrix multiplication
println(a [*] 2)  // Scalar multiplication
println(matrix[[1, 1], [1, 0]] [^] 5)

// Matrix methods.
println(a.row(1))
println(a.col(2))
println(a.dimension())
println(a.transpose())
println(a.rotateLeft())
println(a.rotateRight())
println(matrix[[4, 7], [2, 6]].inverse())

// Type check.
println(a ? matrix)
```

## List lambdas

Inspired by Kotlin, list lambdas allow for list manipulation.

### Map lambda

Maps each element of a list to a new value.

```
local numbers = [1, 2, 3]
// Double all the numbers
local doubled = numbers.map { n -> n * 2 }
println(doubled)  // Output: [2, 4, 6]
```

### Filter lambda

Filters out unwanted elements.

```
local numbers = [1, 2, 3, 4]
// Filter for even numbers
local evens = numbers.filter { n -> n % 2 == 0 }
println(evens)  // Output: [2, 4]
```

### Sort lambda

Helps to define a custom sort method.
Usage `.sort { m, n -> bool_m_preceeds_n } `

```
local names = ["Bob", "Alice", "John"]
// Sort names in descending order
local namesSorted = names
  .sort { m, n -> m.textLen() > m.textLen() }
println(namesSorted) // Output:  ["John", "Alice", "Bob"]
```

### Min and Max lambdas

Sorts the elements in a list and returns the maximum or minimum value.
Usage `.min { m, n -> bool_m_preceeds_n }` and `.max { m, n -> bool_m_preceeds_n }`

```
local names = ["Bob", "Alice", "John"]
// Find the longest name
local longestName = names
  .max { m, n -> m.textLen() < n.textLen() }  // use min { } for the shortest name
println(longestName)
```

### Reduce lambda

Reduce lambda reduces many elements to a single element.
Usage `.reduce(initValue) { x, valueSoFar -> newValue }`

```
local numbers = [1, 2, 3, 4, 5, 6, 7]
// Sum up all the numbers
local numbersSum  = numbers.reduce(0) { x, valueSoFar -> x + valueSoFar }
println(numbersSum) // Output: 28
```

### Example

For example, let’s say Bob has a list of lemons sold per day for the last week and he’d like to calculate his revenue for lemon priced at $2 each.

The days he missed are marked as "N/A"

```
global LemonadeSold = [9, 12, "N/A", 15, 18, "N/A", 8]
``` 

Then we create a function that calculates the total revenue using list lambdas:

```
func GetTotalRevenue() = {
  this.LemonadeSold
    .filter { n -> n ? number }    // Filters for numeric entries, "N/A" is dropped
    .map { n -> n * 2 }	    // Multiply lemons sold in a day by the price of a lemon
    .reduce(0) { x, soFar -> x + soFar }  // Sum up all the entries
}
```

Now, when we call `GetTotalRevenue()`:

```
println("Last week’s revenue was " _ GetTotalRevenue())
```

## Components

### Defining components

```
@ComponentType { InstanceName1, InstanceName2 } 
``` 

e.g.
```
@Button { Button1, Button2 }
```

### Events

```
@Web { Web1 }

when Web1.GotText(url, responseCode, responseType, responseContent) {
  println(responseType)
}
```

### Generic Events

```
@Web { Web1 }

when any Web.GotText(url, responseCode, responseType, responseContent) {
  println(responseType)
}
```

### Property Set

```
@Web { Web1 }

Web1.Url = "https://google.com"
```

### Property Get

```
@Web { Web1 }

println(Web1.Url)
```

### Generic Property Set

```
@Web { Web1 }

set("Web", Web1, "Url", "https://google.com")
```

### Generic Property Get

```
@Web { Web1 }

println(get("Web", Web1, "Url"))
```

### Method Call (limited support)

```
@Web { Web1 }

Web1.Get()
```

### Generic Method Call

Not yet supported
