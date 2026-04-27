package main

import (
	"fmt"
	"os/exec"
)

func main() {
	cmd := exec.Command("./lang/falcon", "run", "testing/run.mist")
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	lines := []string{}
	start := 0
	for i, c := range out {
		if c == '\n' {
			lines = append(lines, string(out[start:i]))
			start = i + 1
		}
	}
	if start < len(out) {
		lines = append(lines, string(out[start:]))
	}

	for i, line := range lines {
		fmt.Printf("Line %d (len=%d): %q\n", i, len(line), line)
		for j, c := range line {
			if c == '^' {
				fmt.Printf("  ^ at rune index %d\n", j)
			}
		}
	}
}
