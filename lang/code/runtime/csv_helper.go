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

// parseCSVTable parses a multi-row CSV string, correctly handling quoted fields
// that span multiple lines (embedded newlines inside quoted fields are preserved).
func parseCSVTable(s string) [][]string {
	var rows [][]string
	var fields []string
	var field strings.Builder
	inQuote := false

	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case inQuote && c == '"':
			if i+1 < len(s) && s[i+1] == '"' {
				field.WriteByte('"')
				i++ // skip escaped quote
			} else {
				inQuote = false
			}
		case !inQuote && c == '"':
			inQuote = true
		case !inQuote && c == ',':
			fields = append(fields, field.String())
			field.Reset()
		case !inQuote && c == '\n':
			// end of row — skip blank lines
			if field.Len() > 0 || len(fields) > 0 {
				fields = append(fields, field.String())
				rows = append(rows, fields)
				fields = nil
				field.Reset()
			}
		case c == '\r':
			// skip CR in CRLF sequences
		default:
			field.WriteByte(c)
		}
	}
	// flush final row if any content remains
	if field.Len() > 0 || len(fields) > 0 {
		fields = append(fields, field.String())
		rows = append(rows, fields)
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