package runtime

// csv.go provides minimal CSV parsing and formatting without encoding/csv.
// Supports quoted fields, escaped double-quotes (""), and multi-line tables.

import "strings"

// parseCSVRow parses a single CSV line into fields.
// Handles quoted fields and "" escape sequences.
func parseCSVRow(s string) []string {
	var fields []string
	var field strings.Builder
	inQuote := false

	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case inQuote && c == '"':
			// peek ahead for escaped quote
			if i+1 < len(s) && s[i+1] == '"' {
				field.WriteByte('"')
				i++
			} else {
				inQuote = false
			}
		case !inQuote && c == '"':
			inQuote = true
		case !inQuote && c == ',':
			fields = append(fields, field.String())
			field.Reset()
		case c == '\r':
			// skip CR in CRLF sequences
		default:
			field.WriteByte(c)
		}
	}
	fields = append(fields, field.String())
	return fields
}

// parseCSVTable splits s into lines and parses each as a CSV row.
func parseCSVTable(s string) [][]string {
	lines := strings.Split(s, "\n")
	rows := make([][]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimRight(line, "\r")
		if line == "" {
			continue
		}
		rows = append(rows, parseCSVRow(line))
	}
	return rows
}

// formatCSVField quotes a field if it contains a comma, double-quote, or newline.
func formatCSVField(s string) string {
	if strings.ContainsAny(s, ",\"\n\r") {
		return "\"" + strings.ReplaceAll(s, "\"", "\"\"") + "\""
	}
	return s
}

// formatCSVRow formats a slice of fields as a single CSV line (no trailing newline).
func formatCSVRow(fields []string) string {
	parts := make([]string, len(fields))
	for i, f := range fields {
		parts[i] = formatCSVField(f)
	}
	return strings.Join(parts, ",")
}