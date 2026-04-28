package mistparser

import (
	"sort"
	"strings"
)

// SourcePatch describes a single in-place text replacement within a source line.
// Start and End are 0-indexed positions within the line (End is exclusive).
// Line is the 1-indexed line number (token.Column).
type SourcePatch struct {
	Line  int
	Start int
	End   int
	Text  string
}

// ApplyPatches applies all patches to the original source string and returns
// the modified source with the same line structure. Patches on the same line
// are applied right-to-left so earlier positions are not invalidated.
func ApplyPatches(source string, patches []SourcePatch) string {
	if len(patches) == 0 {
		return source
	}

	lines := strings.Split(source, "\n")

	byLine := make(map[int][]SourcePatch)
	for _, p := range patches {
		byLine[p.Line] = append(byLine[p.Line], p)
	}

	for lineNum, linePatches := range byLine {
		if lineNum < 1 || lineNum > len(lines) {
			continue
		}
		line := lines[lineNum-1]

		sort.Slice(linePatches, func(i, j int) bool {
			return linePatches[i].Start > linePatches[j].Start
		})

		for _, p := range linePatches {
			if p.Start < 0 || p.End > len(line) || p.Start > p.End {
				continue
			}
			line = line[:p.Start] + p.Text + line[p.End:]
		}

		lines[lineNum-1] = line
	}

	return strings.Join(lines, "\n")
}
