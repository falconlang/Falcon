package mistparser

import (
	"sort"
	"strconv"
	"strings"

	l "Falcon/code/lex"
)

type pendingCallError struct {
	token      *l.Token
	name       string
	suggestion string
}

// renderCallErrorGroups groups bad call names by source line and renders
// each group as a single annotated block. All carets appear on one row
// and all hints appear on the row below that, both aligned to each bad name.
// For method chains, dots are expanded with surrounding spaces for readability.
//
// Example (autocorrect off, three bad names on one line):
//
//	println(s . upperCase() . replaceAll(" ", "") . size())
//	            ^^^^^^^^^     ^^^^^^^^^^            ^^^^
//	            uppercase     replace               textLen  ← correct names
//	[line 2]
func renderCallErrorGroups(byLine map[int][]pendingCallError) []string {
	lineNums := make([]int, 0, len(byLine))
	for ln := range byLine {
		lineNums = append(lineNums, ln)
	}
	sort.Ints(lineNums)

	var blocks []string
	for _, lineNum := range lineNums {
		items := byLine[lineNum]
		sort.Slice(items, func(i, j int) bool {
			return items[i].token.Row < items[j].token.Row
		})

		ctx := items[0].token.Context
		if ctx == nil {
			for _, item := range items {
				msg := "\nNo method named ." + item.name + "()"
				if item.suggestion != "" {
					msg += " → " + item.suggestion
				}
				msg += "\n[line " + strconv.Itoa(lineNum) + "]"
				blocks = append(blocks, msg)
			}
			continue
		}

		sourceLine := (*ctx).GetLine(lineNum)
		formatted, offsets := expandDots(sourceLine)

		// Compute the rightmost extent so buffers are large enough.
		maxEnd := len(formatted) + 20
		for _, item := range items {
			origStart := item.token.Row - len(item.name)
			newStart := offsetAt(offsets, origStart)
			end := newStart + len(item.name)
			if end > maxEnd {
				maxEnd = end + 20
			}
		}

		caretBuf := makeLine(maxEnd)
		hintBuf := makeLine(maxEnd)

		hasSuggestion := false
		for _, item := range items {
			origStart := item.token.Row - len(item.name)
			newStart := offsetAt(offsets, origStart)

			writeTo(caretBuf, newStart, strings.Repeat("^", len(item.name)))

			hint := item.suggestion
			if hint == "" {
				hint = "?"
			} else {
				hasSuggestion = true
			}
			writeTo(hintBuf, newStart, hint)
		}

		hintLine := strings.TrimRight(string(hintBuf), " ")
		if hasSuggestion {
			hintLine += "  ← correct names"
		}

		var sb strings.Builder
		sb.WriteByte('\n')
		sb.WriteString(formatted)
		sb.WriteByte('\n')
		sb.WriteString(strings.TrimRight(string(caretBuf), " "))
		sb.WriteByte('\n')
		sb.WriteString(hintLine)
		sb.WriteByte('\n')
		sb.WriteString("[line " + strconv.Itoa(lineNum) + "]")
		blocks = append(blocks, sb.String())
	}
	return blocks
}

// expandDots returns a copy of line with a space inserted before and after
// every dot that is immediately followed by a letter (a method-call dot).
// It also returns an offsets slice where offsets[i] is the position in the
// formatted string that corresponds to position i in the original line.
func expandDots(line string) (formatted string, offsets []int) {
	offsets = make([]int, len(line)+1)
	var sb strings.Builder
	pos := 0
	for i := 0; i < len(line); i++ {
		offsets[i] = pos
		c := line[i]
		if c == '.' && i+1 < len(line) && isAlphaChar(line[i+1]) {
			sb.WriteString(" . ")
			pos += 3
		} else {
			sb.WriteByte(c)
			pos++
		}
	}
	offsets[len(line)] = pos
	formatted = sb.String()
	return
}

func offsetAt(offsets []int, orig int) int {
	if orig < 0 {
		return 0
	}
	if orig >= len(offsets) {
		return offsets[len(offsets)-1]
	}
	return offsets[orig]
}

func makeLine(size int) []byte {
	b := make([]byte, size)
	for i := range b {
		b[i] = ' '
	}
	return b
}

func writeTo(buf []byte, pos int, text string) {
	for i := 0; i < len(text) && pos+i < len(buf); i++ {
		buf[pos+i] = text[i]
	}
}

func isAlphaChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}
