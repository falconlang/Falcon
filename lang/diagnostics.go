package main

import (
	codecontext "Falcon/code/context"
	"Falcon/design"
)

type wasmDiagnostic struct {
	Message  string
	Severity string
	Phase    string
	File     string
	Line     int
	Column   int
	Length   int
	Raw      string
}

func wasmDiagnosticsFromContext(diagnostics []codecontext.Diagnostic, phase, raw, fallbackFile string) []wasmDiagnostic {
	values := make([]wasmDiagnostic, 0, len(diagnostics))
	for _, diagnostic := range diagnostics {
		file := diagnostic.File
		if file == "" {
			file = fallbackFile
		}
		values = append(values, wasmDiagnostic{
			Message:  diagnostic.Message,
			Severity: diagnostic.Severity,
			Phase:    phase,
			File:     file,
			Line:     max(diagnostic.Line, 1),
			Column:   max(diagnostic.Column, 1),
			Length:   max(diagnostic.Length, 1),
			Raw:      raw,
		})
	}
	return values
}

func fallbackDiagnostic(message, phase, fileName string) []wasmDiagnostic {
	return []wasmDiagnostic{{
		Message:  message,
		Severity: "error",
		Phase:    phase,
		File:     fileName,
		Line:     1,
		Column:   1,
		Length:   1,
		Raw:      message,
	}}
}

func wasmDiagnosticsFromAnn(source string, diagnostics []design.AnnDiagnostic, phase, raw, fileName string) []wasmDiagnostic {
	values := make([]wasmDiagnostic, 0, len(diagnostics))
	for _, diagnostic := range diagnostics {
		line, column := lineColumnForRuneIndex(source, diagnostic.Position)
		values = append(values, wasmDiagnostic{
			Message:  diagnostic.Message,
			Severity: "error",
			Phase:    phase,
			File:     fileName,
			Line:     line,
			Column:   column,
			Length:   max(diagnostic.Length, 1),
			Raw:      raw,
		})
	}
	return values
}

func lineColumnForRuneIndex(source string, runeIndex int) (int, int) {
	line, column := 1, 1
	for i, r := range []rune(source) {
		if i >= runeIndex {
			return line, column
		}
		if r == '\n' {
			line++
			column = 1
		} else {
			column++
		}
	}
	return line, column
}
