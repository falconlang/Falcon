package context

import (
	"Falcon/code/sugar"
	"strconv"
	"strings"
)

type CodeContext struct {
	SourceCode *string
	FileName   string
}

func (c *CodeContext) ReportError(
	column int,
	row int,
	highlightWordSize int,
	message string,
	args ...string,
) {
	panic(c.BuildTracebackError(column, row, highlightWordSize, "CompileError", message, args...))
}

func (c *CodeContext) ReportTypeError(
	column int,
	row int,
	highlightWordSize int,
	message string,
	args ...string,
) {
	panic(c.BuildTracebackError(column, row, highlightWordSize, "TypeError", message, args...))
}

func (c *CodeContext) GetLine(lineNum int) string {
	code := *c.SourceCode
	beginOfLine := sugar.IndexAfterNthOccurrence(code, lineNum-1, '\n') + 1
	endOfLine := strings.Index(code[beginOfLine:], "\n")
	if endOfLine == -1 {
		endOfLine = len(code) - beginOfLine
	}
	return code[beginOfLine : beginOfLine+endOfLine]
}

func (c *CodeContext) BuildCaret(endColumn, highlightSize int) string {
	if highlightSize <= 0 {
		highlightSize = 1
	}
	start := endColumn - highlightSize
	if start < 0 {
		start = 0
	}
	return strings.Repeat(" ", start) + strings.Repeat("^", highlightSize)
}

// FormatTracebackFrame formats a single traceback frame in the style of
// Python tracebacks.  It is the common ground shared by compile-time
// reporting and runtime.FormatRuntimeError.
func (c *CodeContext) FormatTracebackFrame(
	fileName string,
	line int,
	column int,
	highlightSize int,
	funcName string,
) string {
	var sb strings.Builder
	sb.WriteString("  File \"" + fileName + "\", line " + strconv.Itoa(line) + ", in " + funcName + "\n")
	sourceLine := c.GetLine(line)
	sb.WriteString("    " + sourceLine + "\n")
	sb.WriteString("    " + c.BuildCaret(column, highlightSize) + "\n")
	return sb.String()
}

// BuildTracebackError assembles a full traceback-style error message.
func (c *CodeContext) BuildTracebackError(
	column int,
	row int,
	highlightWordSize int,
	title string,
	message string,
	args ...string,
) string {
	var sb strings.Builder
	sb.WriteString("Traceback (most recent call last):\n")
	sb.WriteString(c.FormatTracebackFrame(c.FileName, column, row, highlightWordSize, "<module>"))
	sb.WriteString(title + ": " + sugar.Format(message, args...) + "\n")
	return sb.String()
}

func (c *CodeContext) BuildError(
	decorate bool,
	column int,
	row int,
	highlightWordSize int,
	message string,
	args ...string,
) string {
	err := sugar.Format(message, args...) + "\n[line " + strconv.Itoa(column) + "]"
	code := *c.SourceCode
	beginOfLine := sugar.IndexAfterNthOccurrence(code, column-1, '\n') + 1
	endOfLine := strings.Index(code[beginOfLine:], "\n")

	if endOfLine == -1 {
		endOfLine = len(code) - beginOfLine
	}

	line := code[beginOfLine : beginOfLine+endOfLine]

	var builder strings.Builder
	boxTop := strings.Repeat(".", max(len(line), len(err)))

	builder.WriteByte('\n')
	if decorate {
		builder.WriteString(boxTop)
	}
	builder.WriteByte('\n')
	builder.WriteString(line)
	builder.WriteByte('\n')
	builder.WriteString(strings.Repeat(" ", row-highlightWordSize))
	builder.WriteString(strings.Repeat("^", highlightWordSize))
	builder.WriteByte('\n')
	builder.WriteString(err)
	builder.WriteByte('\n')
	if decorate {
		builder.WriteString(boxTop)
	}
	return builder.String()
}
