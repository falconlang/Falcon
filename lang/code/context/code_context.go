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

type Diagnostic struct {
	Message  string
	Severity string
	File     string
	Line     int
	Column   int
	Length   int
}

type DiagnosticError struct {
	Diagnostic Diagnostic
	Title      string
	Raw        string
}

func (e *DiagnosticError) Error() string {
	return e.Raw
}

type DiagnosticListError struct {
	Message     string
	Raw         string
	Diagnostics []Diagnostic
}

func (e *DiagnosticListError) Error() string {
	return e.Raw
}

func (c *CodeContext) ReportError(
	column int,
	row int,
	highlightWordSize int,
	message string,
	args ...string,
) {
	panic(c.BuildDiagnosticError(column, row, highlightWordSize, "CompileError", message, args...))
}

func (c *CodeContext) ReportTypeError(
	column int,
	row int,
	highlightWordSize int,
	message string,
	args ...string,
) {
	panic(c.BuildDiagnosticError(column, row, highlightWordSize, "TypeError", message, args...))
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

func (c *CodeContext) BuildDiagnostic(
	column int,
	row int,
	highlightWordSize int,
	message string,
	args ...string,
) Diagnostic {
	if highlightWordSize <= 0 {
		highlightWordSize = 1
	}
	startColumn := row - highlightWordSize + 1
	if startColumn < 1 {
		startColumn = 1
	}
	return Diagnostic{
		Message:  sugar.Format(message, args...),
		Severity: "error",
		File:     c.FileName,
		Line:     column,
		Column:   startColumn,
		Length:   highlightWordSize,
	}
}

func (c *CodeContext) BuildDiagnosticError(
	column int,
	row int,
	highlightWordSize int,
	title string,
	message string,
	args ...string,
) *DiagnosticError {
	formatted := sugar.Format(message, args...)
	return &DiagnosticError{
		Diagnostic: c.BuildDiagnostic(column, row, highlightWordSize, formatted),
		Title:      title,
		Raw:        c.BuildTracebackError(column, row, highlightWordSize, title, formatted),
	}
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
