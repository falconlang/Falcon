![image|690x338](https://community.appinventor.mit.edu/uploads/default/original/3X/e/d/ed3b9d22ddaefffb4fd5ab71964b8816c56c63a1.png)

Falcon is a language designed for App Inventor to enable syntax-based programming and for incorporating agenting coding abilities.

## Quirks
1. Falcon follows 1-based indexing.
2. Falcon variables are dynamically typed. Do not declare variables.
3. Lists and dictionaries are passed as references.
4. Falcon follows Kotlin's style of functional expressions.
5. Falcon does not have a return statement; the last expression in a body is returned.
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
    - To check a value for a specific type (`text`, `number`, `list`, `dict`)
        - E.g.,"Hello" ? text`
    - Check for a number type (`number`, `base10`, `hexa`, `bin`)
        - E.g. `"1010" ? bin` is a true expression.
    - Check for empty text (`emptyText`) or an empty list (`emptyList`)
        - E.g. `[] ? emptyList` or `"Cat" ? emptyText`

Operator precedence
The precedence of an operator dictates its parse order. E.g. `*` and `/` is parsed before `+` and `-`.

It is similar to that of Java. Below is a ranking from the lowest to the highest precedence.

1. Assignment `=`
2. Pair `:`
3. TextJoin `_`
4. LogicOr `||`
5. LogicAnd `&&`
6. BitwiseOr `|`
7. BitwiseAnd `&`
8. BitwiseXor `~`
9. Equality `==`, `!=`, `===`, and `!==`
10. Relational `<`, `<=`, `>`, `>=`, `<<`, and `>>`
11. Binary `+`, and `-`
12. BinaryL1 `*`, `/`, and `%`
13. BinaryL2 `^`


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

Or multiple expressions:

```
func FibSum(n) = {
  if (n < 2) {
    n  
  } else {
    FibSum(n - 1) + FibSum(n - 2)
  }
}
```
Note that there is no `return` statement in Falcon. The last statement in a body is taken as the output of an expression.

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
- `makeColor(rgb list)`
- `splitColor(number)`

## Methods

e.g. `"Hello  ".trim()`

### Text

- `textLen()`
- `trim()`
- `uppercase()`
- `lowercase()`
- `startsWith(piece)`
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
- `get(key)`
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
  .max { m, n -> n.textLen() > m.textLen() }  // use min { } for the shortest name
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