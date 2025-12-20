package ast

import (
	"encoding/xml"
	"strconv"
	"strings"
)

type XmlRoot struct {
	XMLName xml.Name `xml:"xml"`
	XMLNS   string   `xml:"xmlns,attr"`
	Blocks  []Block  `xml:"block"`
}

type Block struct {
	XMLName    xml.Name    `xml:"block"`
	Type       string      `xml:"type,attr"`
	Mutation   *Mutation   `xml:"mutation,omitempty"`
	Fields     []Field     `xml:"field"`
	Values     []Value     `xml:"value"`
	Statements []Statement `xml:"statement"`
	Next       *Next       `xml:"next"`
}

type Field struct {
	Name  string `xml:"name,attr"`
	Value string `xml:",chardata"`
}

type Value struct {
	XMLName xml.Name `xml:"value"`
	Name    string   `xml:"name,attr"`
	Block   Block    `xml:"block"`
}

type Mutation struct {
	XMLName       xml.Name     `xml:"mutation"`
	ItemCount     int          `xml:"items,attr"`
	ElseIfCount   int          `xml:"elseif,attr"`
	ElseCount     int          `xml:"else,attr"`
	LocalNames    []LocalName  `xml:"localname"`
	Args          []Arg        `xml:"arg"`
	EventParams   []EventParam `xml:"eventparam"`
	Key           string       `xml:"key,attr,omitempty"`
	SetOrGet      string       `xml:"set_or_get,attr,omitempty"`
	PropertyName  string       `xml:"property_name,attr,omitempty"`
	IsGeneric     bool         `xml:"is_generic,attr,omitempty"`
	ComponentType string       `xml:"component_type,attr,omitempty"`
	InstanceName  string       `xml:"instance_name,attr,omitempty"`
	EventName     string       `xml:"event_name,attr,omitempty"`
	MethodName    string       `xml:"method_name,attr,omitempty"`
	ParamCount    int          `xml:"param_count,attr,omitempty"`
	Mode          string       `xml:"mode,attr,omitempty"`
	Cofounder     string       `xml:"confounder,attr,omitempty"`
	Inline        bool         `xml:"inline,attr,omitempty"`
	Name          string       `xml:"name,attr,omitempty"`
}

type EventParam struct {
	XMLName xml.Name `xml:"eventparam"`
	Name    string   `xml:"name,attr"`
}

type LocalName struct {
	XMLName xml.Name `xml:"localname"`
	Name    string   `xml:"name,attr"`
}

type Statement struct {
	XMLName xml.Name `xml:"statement"`
	Name    string   `xml:"name,attr"`
	Block   *Block   `xml:"block"`
}

type Next struct {
	XMLName xml.Name `xml:"next"`
	Block   *Block   `xml:"block"`
}

type Arg struct {
	Name string `xml:"name,attr"`
}

func FieldsFromMap(m map[string]string) []Field {
	fields := make([]Field, 0, len(m))
	for k, v := range m {
		fields = append(fields, Field{k, v})
	}
	return fields
}

func ToFields(prefix string, values []string) []Field {
	fields := make([]Field, len(values))
	for i, value := range values {
		fields[i] = Field{prefix + strconv.Itoa(i), value}
	}
	return fields
}

func ToArgs(names []string) []Arg {
	args := make([]Arg, len(names))
	for i, name := range names {
		args[i] = Arg{Name: name}
	}
	return args
}

func ValuesByPrefix(namePrefix string, operands []Expr) []Value {
	values := make([]Value, len(operands))
	for i, operand := range operands {
		values[i] = Value{Name: namePrefix + strconv.Itoa(i), Block: operand.Blockly(false)}
	}
	return values
}

func ValueArgsByPrefix(on Expr, onName string, namePrefix string, operands []Expr) []Value {
	values := make([]Value, len(operands)+1)
	values[0] = Value{Name: onName, Block: on.Blockly()}
	for i, operand := range operands {
		values[i+1] = Value{Name: namePrefix + strconv.Itoa(i), Block: operand.Blockly(false)}
	}
	return values
}

func MakeValues(operands []Expr, names ...string) []Value {
	if len(operands) != len(names) {
		panic("len(operands) != len(names)")
	}
	values := make([]Value, len(operands))
	for i, operand := range operands {
		values[i] = Value{Name: names[i], Block: operand.Blockly(false)}
	}
	return values
}

func MakeValueArgs(on Expr, onName string, operands []Expr, names ...string) []Value {
	if len(operands) != len(names) {
		panic("len(operands) != len(names)")
	}
	values := make([]Value, len(operands)+1)
	values[0] = Value{Name: onName, Block: on.Blockly()}
	for i, operand := range operands {
		values[i+1] = Value{Name: names[i], Block: operand.Blockly(false)}
	}
	return values
}

func OptionalStatement(name string, body []Expr) []Statement {
	if len(body) > 0 {
		return []Statement{CreateStatement(name, body)}
	}
	return nil
}

func CreateStatement(name string, body []Expr) Statement {
	// Pass true to indicate a statement
	headBlock := ensureStatement(body[0])
	currBlock := &headBlock
	bodyLen := len(body)
	currI := 1

	for currI < bodyLen {
		aBlock := ensureStatement(body[currI])
		currBlock.Next = &Next{Block: &aBlock}
		currBlock = &aBlock
		currI++
	}
	return Statement{Name: name, Block: &headBlock}
}

func ensureStatement(expr Expr) Block {
	// First evaluate Blockly(). True indicates we expect a statement.
	// This gives time for if expressions to mutate to if statement.
	aBlock := expr.Blockly(true)
	if expr.Consumable(true) {
		// It's still consumable, wrap around evaluate but ignore result
		return Block{Type: "controls_eval_but_ignore", Values: []Value{{Block: aBlock}}}
	}
	return aBlock
}

func ToStatements(namePrefix string, bodies [][]Expr) []Statement {
	var statements []Statement
	for i, aBody := range bodies {
		if len(aBody) > 0 {
			statements = append(statements, CreateStatement(namePrefix+strconv.Itoa(i), aBody))
		}
	}
	return statements
}

func MakeLocalNames(names ...string) []LocalName {
	localNames := make([]LocalName, len(names))
	for i, name := range names {
		localNames[i] = LocalName{Name: name}
	}
	return localNames
}

func JoinExprs(separator string, expressions []Expr) string {
	exprStrings := make([]string, len(expressions))
	for i, expr := range expressions {
		exprStrings[i] = expr.String()
	}
	return strings.Join(exprStrings, separator)
}
