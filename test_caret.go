package main

import (
	"fmt"
	"strings"
)

func main() {
	line := "  local reversed = s.lowercase().split(\" \").reverse().join(\"\")"
	caret := strings.Repeat(" ", 19) + "^"
	fmt.Println("    " + line)
	fmt.Println("    " + caret)
	fmt.Println("s is at index", 4+19, "in prefixed line")
}
