package ast

import (
	"Falcon/code/xmlutil"
	"strconv"
	"strings"
)

// XmlRoot is the root <xml> element for a Blockly XML document.
type XmlRoot struct {
	XMLNS  string
	Blocks []Block
}

type Block struct {
	Type       string
	Disabled   bool
	Mutation   *Mutation
	Fields     []Field
	Values     []Value
	Statements []Statement
	Next       *Next
}

type Field struct {
	Name  string
	Value string
}

type Value struct {
	Name   string
	Block  Block
	Shadow *Shadow
}

type Shadow struct {
	Type       string
	Disabled   bool
	Mutation   *Mutation
	Fields     []Field
	Values     []Value
	Statements []Statement
	Next       *Next
}

func (s *Shadow) BlockValue() Block {
	if s == nil {
		return Block{}
	}
	return Block{
		Type:       s.Type,
		Disabled:   s.Disabled,
		Mutation:   s.Mutation,
		Fields:     s.Fields,
		Values:     s.Values,
		Statements: s.Statements,
		Next:       s.Next,
	}
}

type Mutation struct {
	ItemCount     int
	Rows          int
	Cols          int
	Matrix        string
	ElseIfCount   int
	ElseCount     int
	LocalNames    []LocalName
	Args          []Arg
	EventParams   []EventParam
	Key           string
	SetOrGet      string
	PropertyName  string
	IsGeneric     bool
	ComponentType string
	InstanceName  string
	EventName     string
	MethodName    string
	Shape         string
	ParamCount    int
	Mode          string
	Cofounder     string
	Inline        bool
	Name          string
}

type EventParam struct {
	Name string
}

type LocalName struct {
	Name string
}

type Statement struct {
	Name  string
	Block *Block
}

type Next struct {
	Block *Block
}

type Arg struct {
	Name string
	Type string
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
	values[0] = Value{Name: onName, Block: on.Blockly(false)}
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
	values[0] = Value{Name: onName, Block: on.Blockly(false)}
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
	aBlock := expr.Blockly(true)
	if expr.Consumable() {
		panic("result of `" + expr.String() + "` is never used")
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

// ── XML serialization ─────────────────────────────────────────────────────────

// MarshalIndent serializes an XmlRoot to indented XML, matching the output of
// encoding/xml.MarshalIndent.
func (r XmlRoot) MarshalIndent(prefix, indent string) []byte {
	var sb strings.Builder
	sb.WriteString(prefix)
	sb.WriteString(`<xml xmlns="`)
	sb.WriteString(xmlutil.EscapeText(r.XMLNS))
	sb.WriteString(`">`)
	for _, b := range r.Blocks {
		writeBlock(&sb, b, prefix+indent, indent)
	}
	sb.WriteByte('\n')
	sb.WriteString(prefix)
	sb.WriteString("</xml>")
	return []byte(sb.String())
}

func writeBlock(sb *strings.Builder, b Block, ind, step string) {
	sb.WriteByte('\n')
	sb.WriteString(ind)
	sb.WriteString(`<block type="`)
	sb.WriteString(xmlutil.EscapeText(b.Type))
	sb.WriteByte('"')
	if b.Disabled {
		sb.WriteString(` disabled="true"`)
	}
	sb.WriteByte('>')
	writeMutation(sb, b.Mutation, ind+step, step)
	for _, f := range b.Fields {
		sb.WriteByte('\n')
		sb.WriteString(ind + step)
		sb.WriteString(`<field name="`)
		sb.WriteString(xmlutil.EscapeText(f.Name))
		sb.WriteString(`">`)
		sb.WriteString(xmlutil.EscapeText(f.Value))
		sb.WriteString("</field>")
	}
	for _, v := range b.Values {
		sb.WriteByte('\n')
		sb.WriteString(ind + step)
		sb.WriteString(`<value name="`)
		sb.WriteString(xmlutil.EscapeText(v.Name))
		sb.WriteString(`">`)
		writeValueContent(sb, v, ind+step, step)
		sb.WriteByte('\n')
		sb.WriteString(ind + step)
		sb.WriteString("</value>")
	}
	for _, stmt := range b.Statements {
		sb.WriteByte('\n')
		sb.WriteString(ind + step)
		sb.WriteString(`<statement name="`)
		sb.WriteString(xmlutil.EscapeText(stmt.Name))
		sb.WriteString(`">`)
		if stmt.Block != nil {
			writeBlock(sb, *stmt.Block, ind+step+step, step)
			sb.WriteByte('\n')
			sb.WriteString(ind + step)
		}
		sb.WriteString("</statement>")
	}
	if b.Next != nil && b.Next.Block != nil {
		sb.WriteByte('\n')
		sb.WriteString(ind + step)
		sb.WriteString("<next>")
		writeBlock(sb, *b.Next.Block, ind+step+step, step)
		sb.WriteByte('\n')
		sb.WriteString(ind + step)
		sb.WriteString("</next>")
	}
	sb.WriteByte('\n')
	sb.WriteString(ind)
	sb.WriteString("</block>")
}

func writeValueContent(sb *strings.Builder, v Value, ind, step string) {
	if v.Block.Type != "" {
		writeBlock(sb, v.Block, ind+step, step)
		return
	}
	if v.Shadow != nil {
		writeShadow(sb, v.Shadow, ind+step, step)
	}
}

func writeShadow(sb *strings.Builder, s *Shadow, ind, step string) {
	if s == nil {
		return
	}
	sb.WriteByte('\n')
	sb.WriteString(ind)
	sb.WriteString(`<shadow type="`)
	sb.WriteString(xmlutil.EscapeText(s.Type))
	sb.WriteByte('"')
	if s.Disabled {
		sb.WriteString(` disabled="true"`)
	}
	sb.WriteByte('>')
	writeMutation(sb, s.Mutation, ind+step, step)
	for _, f := range s.Fields {
		sb.WriteByte('\n')
		sb.WriteString(ind + step)
		sb.WriteString(`<field name="`)
		sb.WriteString(xmlutil.EscapeText(f.Name))
		sb.WriteString(`">`)
		sb.WriteString(xmlutil.EscapeText(f.Value))
		sb.WriteString("</field>")
	}
	sb.WriteByte('\n')
	sb.WriteString(ind)
	sb.WriteString("</shadow>")
}

func writeMutation(sb *strings.Builder, m *Mutation, ind, step string) {
	if m == nil {
		return
	}
	sb.WriteByte('\n')
	sb.WriteString(ind)
	sb.WriteString("<mutation")
	writeIntAttr(sb, "items", m.ItemCount)
	writeIntAttrOmitZero(sb, "rows", m.Rows)
	writeIntAttrOmitZero(sb, "cols", m.Cols)
	writeStrAttr(sb, "matrix", m.Matrix)
	writeIntAttrOmitZero(sb, "elseif", m.ElseIfCount)
	writeIntAttrOmitZero(sb, "else", m.ElseCount)
	writeStrAttr(sb, "key", m.Key)
	writeStrAttr(sb, "set_or_get", m.SetOrGet)
	writeStrAttr(sb, "property_name", m.PropertyName)
	writeBoolAttr(sb, "is_generic", m.IsGeneric)
	writeStrAttr(sb, "component_type", m.ComponentType)
	writeStrAttr(sb, "instance_name", m.InstanceName)
	writeStrAttr(sb, "event_name", m.EventName)
	writeStrAttr(sb, "method_name", m.MethodName)
	writeStrAttr(sb, "shape", m.Shape)
	writeIntAttrOmitZero(sb, "param_count", m.ParamCount)
	writeStrAttr(sb, "mode", m.Mode)
	writeStrAttr(sb, "confounder", m.Cofounder)
	writeBoolAttr(sb, "inline", m.Inline)
	writeStrAttr(sb, "name", m.Name)

	hasChildren := len(m.LocalNames) > 0 || len(m.Args) > 0 || len(m.EventParams) > 0
	if !hasChildren {
		sb.WriteString("/>")
		return
	}
	sb.WriteByte('>')
	for _, ln := range m.LocalNames {
		sb.WriteByte('\n')
		sb.WriteString(ind + step)
		sb.WriteString(`<localname name="`)
		sb.WriteString(xmlutil.EscapeText(ln.Name))
		sb.WriteString(`"/>`)
	}
	for _, arg := range m.Args {
		sb.WriteByte('\n')
		sb.WriteString(ind + step)
		sb.WriteString(`<arg name="`)
		sb.WriteString(xmlutil.EscapeText(arg.Name))
		sb.WriteByte('"')
		if arg.Type != "" {
			sb.WriteString(` type="`)
			sb.WriteString(xmlutil.EscapeText(arg.Type))
			sb.WriteByte('"')
		}
		sb.WriteString("/>")
	}
	for _, ep := range m.EventParams {
		sb.WriteByte('\n')
		sb.WriteString(ind + step)
		sb.WriteString(`<eventparam name="`)
		sb.WriteString(xmlutil.EscapeText(ep.Name))
		sb.WriteString(`"/>`)
	}
	sb.WriteByte('\n')
	sb.WriteString(ind)
	sb.WriteString("</mutation>")
}

func writeIntAttr(sb *strings.Builder, name string, v int) {
	sb.WriteByte(' ')
	sb.WriteString(name)
	sb.WriteString(`="`)
	sb.WriteString(strconv.Itoa(v))
	sb.WriteByte('"')
}

func writeIntAttrOmitZero(sb *strings.Builder, name string, v int) {
	if v != 0 {
		writeIntAttr(sb, name, v)
	}
}

func writeStrAttr(sb *strings.Builder, name, v string) {
	if v != "" {
		sb.WriteByte(' ')
		sb.WriteString(name)
		sb.WriteString(`="`)
		sb.WriteString(xmlutil.EscapeText(v))
		sb.WriteByte('"')
	}
}

func writeBoolAttr(sb *strings.Builder, name string, v bool) {
	if v {
		sb.WriteByte(' ')
		sb.WriteString(name)
		sb.WriteString(`="true"`)
	}
}

// ── XML deserialization ───────────────────────────────────────────────────────

// ParseBlocklyXML parses a Blockly XML string into an XmlRoot.
func ParseBlocklyXML(src string) (XmlRoot, error) {
	doc, err := xmlutil.ParseDocument(src)
	if err != nil {
		return XmlRoot{}, err
	}
	root := XmlRoot{
		XMLNS: doc.AttrVal("xmlns"),
	}
	for _, child := range doc.Children {
		if child.Name == "block" {
			root.Blocks = append(root.Blocks, elementToBlock(child))
		}
	}
	return root, nil
}

func elementToBlock(e *xmlutil.Element) Block {
	b := Block{
		Type:     e.AttrVal("type"),
		Disabled: e.AttrBool("disabled"),
	}
	for _, child := range e.Children {
		switch child.Name {
		case "mutation":
			m := elementToMutation(child)
			b.Mutation = &m
		case "field":
			b.Fields = append(b.Fields, Field{
				Name:  child.AttrVal("name"),
				Value: child.Text,
			})
		case "value":
			b.Values = append(b.Values, elementToValue(child))
		case "statement":
			b.Statements = append(b.Statements, elementToStatement(child))
		case "next":
			if blk := child.FirstChild("block"); blk != nil {
				nb := elementToBlock(blk)
				b.Next = &Next{Block: &nb}
			}
		}
	}
	return b
}

func elementToValue(e *xmlutil.Element) Value {
	v := Value{Name: e.AttrVal("name")}
	for _, child := range e.Children {
		switch child.Name {
		case "block":
			v.Block = elementToBlock(child)
		case "shadow":
			s := elementToShadow(child)
			v.Shadow = &s
		}
	}
	return v
}

func elementToStatement(e *xmlutil.Element) Statement {
	s := Statement{Name: e.AttrVal("name")}
	if blk := e.FirstChild("block"); blk != nil {
		b := elementToBlock(blk)
		s.Block = &b
	}
	return s
}

func elementToShadow(e *xmlutil.Element) Shadow {
	s := Shadow{
		Type:     e.AttrVal("type"),
		Disabled: e.AttrBool("disabled"),
	}
	for _, child := range e.Children {
		switch child.Name {
		case "mutation":
			m := elementToMutation(child)
			s.Mutation = &m
		case "field":
			s.Fields = append(s.Fields, Field{
				Name:  child.AttrVal("name"),
				Value: child.Text,
			})
		case "value":
			s.Values = append(s.Values, elementToValue(child))
		case "statement":
			s.Statements = append(s.Statements, elementToStatement(child))
		case "next":
			if blk := child.FirstChild("block"); blk != nil {
				nb := elementToBlock(blk)
				s.Next = &Next{Block: &nb}
			}
		}
	}
	return s
}

func elementToMutation(e *xmlutil.Element) Mutation {
	m := Mutation{
		ItemCount:     e.AttrInt("items"),
		Rows:          e.AttrInt("rows"),
		Cols:          e.AttrInt("cols"),
		Matrix:        e.AttrVal("matrix"),
		ElseIfCount:   e.AttrInt("elseif"),
		ElseCount:     e.AttrInt("else"),
		Key:           e.AttrVal("key"),
		SetOrGet:      e.AttrVal("set_or_get"),
		PropertyName:  e.AttrVal("property_name"),
		IsGeneric:     e.AttrBool("is_generic"),
		ComponentType: e.AttrVal("component_type"),
		InstanceName:  e.AttrVal("instance_name"),
		EventName:     e.AttrVal("event_name"),
		MethodName:    e.AttrVal("method_name"),
		Shape:         e.AttrVal("shape"),
		ParamCount:    e.AttrInt("param_count"),
		Mode:          e.AttrVal("mode"),
		Cofounder:     e.AttrVal("confounder"),
		Inline:        e.AttrBool("inline"),
		Name:          e.AttrVal("name"),
	}
	for _, child := range e.Children {
		switch child.Name {
		case "localname":
			m.LocalNames = append(m.LocalNames, LocalName{Name: child.AttrVal("name")})
		case "arg":
			m.Args = append(m.Args, Arg{Name: child.AttrVal("name"), Type: child.AttrVal("type")})
		case "eventparam":
			m.EventParams = append(m.EventParams, EventParam{Name: child.AttrVal("name")})
		}
	}
	return m
}
