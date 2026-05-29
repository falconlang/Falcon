// Package jsonutil provides a minimal JSON parser and builder that does not
// import fmt or reflect, keeping WASM binary size small.
package jsonutil

import (
	"errors"
	"sort"
	"strconv"
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

// ParseAny parses JSON bytes and returns a Go value using the same type mapping
// as encoding/json: objects→map[string]interface{}, arrays→[]interface{},
// strings→string, numbers→float64, booleans→bool, null→nil.
func ParseAny(data []byte) (interface{}, error) {
	p := jsonParser{data: data}
	v, err := p.parseValue()
	if err != nil {
		return nil, err
	}
	p.skipWS()
	if p.pos < len(p.data) {
		return nil, errors.New("unexpected trailing content at position " + strconv.Itoa(p.pos))
	}
	return v, nil
}

// Marshal converts v to compact JSON.
// v must be one of: nil, bool, float64, int, string, []interface{}, map[string]interface{}.
func Marshal(v interface{}) ([]byte, error) {
	var sb strings.Builder
	if err := marshalValue(&sb, v); err != nil {
		return nil, err
	}
	return []byte(sb.String()), nil
}

// MarshalIndent converts v to indented JSON.
func MarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	var sb strings.Builder
	if err := marshalIndent(&sb, v, prefix, indent, 0); err != nil {
		return nil, err
	}
	return []byte(sb.String()), nil
}

// ── parser ────────────────────────────────────────────────────────────────────

type jsonParser struct {
	data []byte
	pos  int
}

func (p *jsonParser) skipWS() {
	for p.pos < len(p.data) {
		switch p.data[p.pos] {
		case ' ', '\t', '\n', '\r':
			p.pos++
		default:
			return
		}
	}
}

func (p *jsonParser) parseValue() (interface{}, error) {
	p.skipWS()
	if p.pos >= len(p.data) {
		return nil, errors.New("unexpected end of JSON input")
	}
	switch p.data[p.pos] {
	case '{':
		return p.parseObject()
	case '[':
		return p.parseArray()
	case '"':
		return p.parseString()
	case 't':
		return p.parseLit("true", true)
	case 'f':
		return p.parseLit("false", false)
	case 'n':
		return p.parseLit("null", nil)
	default:
		return p.parseNumber()
	}
}

func (p *jsonParser) parseObject() (map[string]interface{}, error) {
	p.pos++ // consume '{'
	obj := make(map[string]interface{})
	p.skipWS()
	if p.pos < len(p.data) && p.data[p.pos] == '}' {
		p.pos++
		return obj, nil
	}
	for {
		p.skipWS()
		key, err := p.parseString()
		if err != nil {
			return nil, err
		}
		p.skipWS()
		if p.pos >= len(p.data) || p.data[p.pos] != ':' {
			return nil, errors.New("expected ':' in JSON object")
		}
		p.pos++ // consume ':'
		val, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		obj[key] = val
		p.skipWS()
		if p.pos >= len(p.data) {
			return nil, errors.New("unexpected end of JSON object")
		}
		switch p.data[p.pos] {
		case '}':
			p.pos++
			return obj, nil
		case ',':
			p.pos++ // consume ','
		default:
			return nil, errors.New("expected ',' or '}' in JSON object at position " + strconv.Itoa(p.pos))
		}
	}
}

func (p *jsonParser) parseArray() ([]interface{}, error) {
	p.pos++ // consume '['
	var arr []interface{}
	p.skipWS()
	if p.pos < len(p.data) && p.data[p.pos] == ']' {
		p.pos++
		return arr, nil
	}
	for {
		val, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		arr = append(arr, val)
		p.skipWS()
		if p.pos >= len(p.data) {
			return nil, errors.New("unexpected end of JSON array")
		}
		switch p.data[p.pos] {
		case ']':
			p.pos++
			return arr, nil
		case ',':
			p.pos++ // consume ','
		default:
			return nil, errors.New("expected ',' or ']' in JSON array at position " + strconv.Itoa(p.pos))
		}
	}
}

func (p *jsonParser) parseString() (string, error) {
	if p.pos >= len(p.data) || p.data[p.pos] != '"' {
		return "", errors.New("expected '\"' at position " + strconv.Itoa(p.pos))
	}
	p.pos++ // consume '"'
	var sb strings.Builder
	for p.pos < len(p.data) {
		c := p.data[p.pos]
		if c == '"' {
			p.pos++
			return sb.String(), nil
		}
		if c == '\\' {
			p.pos++
			if p.pos >= len(p.data) {
				return "", errors.New("unexpected end of string escape")
			}
			esc := p.data[p.pos]
			p.pos++
			switch esc {
			case '"', '\\', '/':
				sb.WriteByte(esc)
			case 'b':
				sb.WriteByte('\b')
			case 'f':
				sb.WriteByte('\f')
			case 'n':
				sb.WriteByte('\n')
			case 'r':
				sb.WriteByte('\r')
			case 't':
				sb.WriteByte('\t')
			case 'u':
				r1, err := p.parseHex4()
				if err != nil {
					return "", err
				}
				// handle surrogate pairs
				if r1 >= 0xD800 && r1 <= 0xDBFF && p.pos+1 < len(p.data) && p.data[p.pos] == '\\' && p.data[p.pos+1] == 'u' {
					p.pos += 2
					r2, err := p.parseHex4()
					if err != nil {
						return "", err
					}
					sb.WriteRune(utf16.DecodeRune(rune(r1), rune(r2)))
				} else {
					sb.WriteRune(rune(r1))
				}
			default:
				sb.WriteByte(esc)
			}
		} else {
			r, size := utf8.DecodeRune(p.data[p.pos:])
			sb.WriteRune(r)
			p.pos += size
		}
	}
	return "", errors.New("unterminated JSON string")
}

func (p *jsonParser) parseHex4() (uint16, error) {
	if p.pos+4 > len(p.data) {
		return 0, errors.New("invalid \\u escape: not enough input")
	}
	var result uint16
	for i := 0; i < 4; i++ {
		c := p.data[p.pos]
		p.pos++
		var v uint16
		switch {
		case c >= '0' && c <= '9':
			v = uint16(c - '0')
		case c >= 'a' && c <= 'f':
			v = uint16(c-'a') + 10
		case c >= 'A' && c <= 'F':
			v = uint16(c-'A') + 10
		default:
			return 0, errors.New("invalid hex digit in \\u escape")
		}
		result = result*16 + v
	}
	return result, nil
}

func (p *jsonParser) parseLit(lit string, val interface{}) (interface{}, error) {
	end := p.pos + len(lit)
	if end > len(p.data) || string(p.data[p.pos:end]) != lit {
		return nil, errors.New("unexpected token at position " + strconv.Itoa(p.pos))
	}
	p.pos = end
	return val, nil
}

func (p *jsonParser) parseNumber() (float64, error) {
	start := p.pos
	if p.pos < len(p.data) && p.data[p.pos] == '-' {
		p.pos++
	}
	for p.pos < len(p.data) && p.data[p.pos] >= '0' && p.data[p.pos] <= '9' {
		p.pos++
	}
	if p.pos < len(p.data) && p.data[p.pos] == '.' {
		p.pos++
		for p.pos < len(p.data) && p.data[p.pos] >= '0' && p.data[p.pos] <= '9' {
			p.pos++
		}
	}
	if p.pos < len(p.data) && (p.data[p.pos] == 'e' || p.data[p.pos] == 'E') {
		p.pos++
		if p.pos < len(p.data) && (p.data[p.pos] == '+' || p.data[p.pos] == '-') {
			p.pos++
		}
		for p.pos < len(p.data) && p.data[p.pos] >= '0' && p.data[p.pos] <= '9' {
			p.pos++
		}
	}
	if p.pos == start {
		return 0, errors.New("expected number at position " + strconv.Itoa(p.pos))
	}
	return strconv.ParseFloat(string(p.data[start:p.pos]), 64)
}

// ── builder ───────────────────────────────────────────────────────────────────

func marshalValue(sb *strings.Builder, v interface{}) error {
	switch val := v.(type) {
	case nil:
		sb.WriteString("null")
	case bool:
		if val {
			sb.WriteString("true")
		} else {
			sb.WriteString("false")
		}
	case float64:
		sb.WriteString(strconv.FormatFloat(val, 'f', -1, 64))
	case int:
		sb.WriteString(strconv.Itoa(val))
	case string:
		writeJSONString(sb, val)
	case []interface{}:
		sb.WriteByte('[')
		for i, elem := range val {
			if i > 0 {
				sb.WriteByte(',')
			}
			if err := marshalValue(sb, elem); err != nil {
				return err
			}
		}
		sb.WriteByte(']')
	case map[string]interface{}:
		sb.WriteByte('{')
		keys := sortedKeys(val)
		for i, k := range keys {
			if i > 0 {
				sb.WriteByte(',')
			}
			writeJSONString(sb, k)
			sb.WriteByte(':')
			if err := marshalValue(sb, val[k]); err != nil {
				return err
			}
		}
		sb.WriteByte('}')
	default:
		return errors.New("jsonutil: unsupported type in Marshal")
	}
	return nil
}

func marshalIndent(sb *strings.Builder, v interface{}, prefix, indent string, depth int) error {
	cur := prefix + strings.Repeat(indent, depth)
	next := cur + indent
	switch val := v.(type) {
	case nil:
		sb.WriteString("null")
	case bool:
		if val {
			sb.WriteString("true")
		} else {
			sb.WriteString("false")
		}
	case float64:
		sb.WriteString(strconv.FormatFloat(val, 'f', -1, 64))
	case int:
		sb.WriteString(strconv.Itoa(val))
	case string:
		writeJSONString(sb, val)
	case []interface{}:
		if len(val) == 0 {
			sb.WriteString("[]")
			return nil
		}
		sb.WriteString("[\n")
		for i, elem := range val {
			sb.WriteString(next)
			if err := marshalIndent(sb, elem, prefix, indent, depth+1); err != nil {
				return err
			}
			if i < len(val)-1 {
				sb.WriteByte(',')
			}
			sb.WriteByte('\n')
		}
		sb.WriteString(cur + "]")
	case map[string]interface{}:
		if len(val) == 0 {
			sb.WriteString("{}")
			return nil
		}
		sb.WriteString("{\n")
		keys := sortedKeys(val)
		for i, k := range keys {
			sb.WriteString(next)
			writeJSONString(sb, k)
			sb.WriteString(": ")
			if err := marshalIndent(sb, val[k], prefix, indent, depth+1); err != nil {
				return err
			}
			if i < len(keys)-1 {
				sb.WriteByte(',')
			}
			sb.WriteByte('\n')
		}
		sb.WriteString(cur + "}")
	default:
		return errors.New("jsonutil: unsupported type in MarshalIndent")
	}
	return nil
}

func writeJSONString(sb *strings.Builder, s string) {
	sb.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			sb.WriteString(`\"`)
		case '\\':
			sb.WriteString(`\\`)
		case '\n':
			sb.WriteString(`\n`)
		case '\r':
			sb.WriteString(`\r`)
		case '\t':
			sb.WriteString(`\t`)
		case '\b':
			sb.WriteString(`\b`)
		case '\f':
			sb.WriteString(`\f`)
		default:
			if r < 0x20 {
				sb.WriteString(`\u00`)
				sb.WriteByte("0123456789abcdef"[r>>4])
				sb.WriteByte("0123456789abcdef"[r&0xf])
			} else {
				sb.WriteRune(r)
			}
		}
	}
	sb.WriteByte('"')
}

func sortedKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// ── helpers for callers ───────────────────────────────────────────────────────

// GetString extracts a string field from a parsed JSON object.
func GetString(obj map[string]interface{}, key string) string {
	if v, ok := obj[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// GetBool extracts a boolean field from a parsed JSON object.
func GetBool(obj map[string]interface{}, key string) bool {
	if v, ok := obj[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

// GetArray extracts a []interface{} field from a parsed JSON object.
func GetArray(obj map[string]interface{}, key string) []interface{} {
	if v, ok := obj[key]; ok {
		if a, ok := v.([]interface{}); ok {
			return a
		}
	}
	return nil
}

// GetObject extracts a map[string]interface{} field from a parsed JSON object.
func GetObject(obj map[string]interface{}, key string) map[string]interface{} {
	if v, ok := obj[key]; ok {
		if m, ok := v.(map[string]interface{}); ok {
			return m
		}
	}
	return nil
}
