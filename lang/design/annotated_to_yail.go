package design

import (
	"Falcon/code/compdb"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode/utf16"
)

const yailComponentPkg = "com.google.appinventor.components.runtime."

type AnnYailConverter struct {
	autoIdCount map[string]int
}

func NewAnnYailConverter() *AnnYailConverter {
	return &AnnYailConverter{autoIdCount: make(map[string]int)}
}

// ParseAnn parses an annotated .ann source and returns the root Component tree.
func ParseAnn(source string) (Component, error) {
	return NewAimlParser(source).parseDocument()
}

// ExtractComponents walks a Component tree and builds:
//   - typeMap:    component type → list of instance names  (e.g. "Button" → ["AddButton", "SubtractButton"])
//   - reverseMap: instance name → component type          (e.g. "AddButton" → "Button")
//
// Auto-IDs are generated with the same logic as AnnYailConverter.
func ExtractComponents(root Component) (typeMap map[string][]string, reverseMap map[string]string) {
	typeMap = make(map[string][]string)
	reverseMap = make(map[string]string)
	autoIds := make(map[string]int)
	walkExtractComponents(root.Children, typeMap, reverseMap, autoIds)
	return
}

func walkExtractComponents(children []Component, typeMap map[string][]string, reverseMap map[string]string, autoIds map[string]int) {
	for _, child := range children {
		id := child.Id
		if id == "" {
			count := autoIds[child.Type] + 1
			autoIds[child.Type] = count
			id = child.Type + strconv.Itoa(count)
		}
		typeMap[child.Type] = append(typeMap[child.Type], id)
		reverseMap[id] = child.Type
		walkExtractComponents(child.Children, typeMap, reverseMap, autoIds)
	}
}

// ConvertAnnToYail parses an annotated .ann screen file and generates standalone
// design YAIL for testing/inspection.
func (c *AnnYailConverter) ConvertAnnToYail(source string) (string, error) {
	screen, err := ParseAnn(source)
	if err != nil {
		return "", err
	}
	return c.generateYail(screen, "")
}

// ConvertAnnToReplYail generates combined YAIL for the App Inventor Companion REPL.
// codeYail (procedures, globals, event handlers from a .mist file) is inserted
// after (clear-current-form) and before the component declarations, matching
// AppInventor's own code ordering so that event handlers are registered before
// call-Initialize-of-components fires.
//
// The caller should wrap the result in:
//
//	(begin (require <com.google.youngandroid.runtime>) (process-repl-input -1 <result>))
func (c *AnnYailConverter) ConvertAnnToReplYail(annSource, codeYail string) (string, error) {
	screen, err := ParseAnn(annSource)
	if err != nil {
		return "", err
	}
	return c.generateYail(screen, codeYail)
}

// dbCompType maps the .ann component type to the DB component type.
// The root screen is @Screen in .ann but registered as "Form" in the component DB.
func dbCompType(annType string) string {
	if annType == "Screen" {
		return "Form"
	}
	return annType
}

// ValidateAnnSource parses source and validates every property name against the
// component DB. All errors are returned together so the caller can report them
// all at once rather than one at a time.
func ValidateAnnSource(source string) error {
	screen, err := ParseAnn(source)
	if err != nil {
		return err
	}
	return validateAnn(screen)
}

// validateAnn walks the entire component tree and collects every property
// validation error. All errors are returned together so the user can fix them
// all in one pass rather than one at a time.
func validateAnn(screen Component) error {
	var diagnostics []AnnDiagnostic
	state := &annValidationState{
		seenIds: make(map[string]bool),
		autoIds: make(map[string]int),
	}
	collectComponentPropertyErrors(screen, Component{}, &diagnostics, state, true)
	if len(diagnostics) == 0 {
		return nil
	}

	sort.SliceStable(diagnostics, func(i, j int) bool {
		return diagnostics[i].Position < diagnostics[j].Position
	})

	messages := make([]string, len(diagnostics))
	for i, diagnostic := range diagnostics {
		messages[i] = diagnostic.Message
	}
	return &AnnDiagnosticListError{
		Message:     "design validation failed",
		Raw:         strings.Join(messages, "\n"),
		Diagnostics: diagnostics,
	}
}

type annValidationState struct {
	seenIds map[string]bool
	autoIds map[string]int
}

func componentValidationId(component Component, state *annValidationState, isRoot bool) string {
	if component.Id != "" {
		return component.Id
	}
	if isRoot {
		return "Screen1"
	}
	count := state.autoIds[component.Type] + 1
	state.autoIds[component.Type] = count
	return component.Type + strconv.Itoa(count)
}

func componentNamePosition(component Component) int {
	if component.Id != "" && component.idPosition > 0 {
		return component.idPosition
	}
	return component.typePosition
}

func collectComponentPropertyErrors(component Component, parent Component, diagnostics *[]AnnDiagnostic, state *annValidationState, isRoot bool) {
	compId := componentValidationId(component, state, isRoot)
	dbType := dbCompType(component.Type)
	parentType := dbCompType(parent.Type)
	parentId := parent.Id
	if parentId == "" {
		parentId = parent.Type
	}

	if !compdb.GlobalDB.HasComponent(dbType) {
		length := len([]rune(component.Type))
		if length < 1 {
			length = 1
		}
		*diagnostics = append(*diagnostics, AnnDiagnostic{
			Message:  fmt.Sprintf("unknown component type %q", component.Type),
			Position: component.typePosition,
			Length:   length,
		})
	}
	if state.seenIds[compId] {
		length := len([]rune(compId))
		if length < 1 {
			length = 1
		}
		*diagnostics = append(*diagnostics, AnnDiagnostic{
			Message:  fmt.Sprintf("duplicate component name %q", compId),
			Position: componentNamePosition(component),
			Length:   length,
		})
	} else {
		state.seenIds[compId] = true
	}
	if parent.Type != "" && !compdb.GlobalDB.CanContainComponent(parentType, dbType) {
		length := len([]rune(component.Type))
		if length < 1 {
			length = 1
		}
		*diagnostics = append(*diagnostics, AnnDiagnostic{
			Message:  fmt.Sprintf("%s %q cannot be placed inside %s %q", component.Type, compId, parent.Type, parentId),
			Position: component.typePosition,
			Length:   length,
		})
	}

	for propName := range component.Properties {
		if err := compdb.GlobalDB.ValidateProperty(dbType, propName); err != nil {
			message := fmt.Sprintf("%s %q: %v", component.Type, compId, err)
			position := component.typePosition
			if propPosition, ok := component.propertyPositions[propName]; ok {
				position = propPosition
			}
			length := len([]rune(propName))
			if propLength, ok := component.propertyLengths[propName]; ok && propLength > 0 {
				length = propLength
			}
			*diagnostics = append(*diagnostics, AnnDiagnostic{
				Message:  message,
				Position: position,
				Length:   length,
			})
		}
	}

	for _, child := range component.Children {
		collectComponentPropertyErrors(child, component, diagnostics, state, false)
	}
}

func (c *AnnYailConverter) generateYail(screen Component, codeYail string) (string, error) {
	if err := validateAnn(screen); err != nil {
		return "", err
	}

	// Reset auto-IDs so each conversion starts from the same state.
	c.autoIdCount = make(map[string]int)

	formName := screen.Id
	if formName == "" {
		formName = "Screen1"
	}

	var sb strings.Builder
	var componentNames []string

	sb.WriteString("(begin\n")
	sb.WriteString("(clear-current-form)\n")

	if formName != "Screen1" {
		sb.WriteString("(rename-component \"Screen1\" \"")
		sb.WriteString(formName)
		sb.WriteString("\")\n")
	}

	if codeYail != "" {
		sb.WriteString(codeYail)
		sb.WriteString("\n")
	}

	if len(screen.Properties) > 0 {
		sb.WriteString("(do-after-form-creation")
		for k, v := range screen.Properties {
			sb.WriteString("\n  (set-and-coerce-property! '")
			sb.WriteString(formName)
			sb.WriteString(" '")
			sb.WriteString(k)
			sb.WriteString(" ")
			sb.WriteString(annFormatValue(screen.Type, k, v))
			sb.WriteString(" ")
			sb.WriteString(annYailType(screen.Type, k, v))
			sb.WriteString(")")
		}
		sb.WriteString(")\n")
	}

	c.genComponents(formName, screen.Children, &sb, &componentNames)

	sb.WriteString("(init-runtime)\n")
	sb.WriteString("(call-Initialize-of-components '")
	sb.WriteString(formName)
	for _, name := range componentNames {
		sb.WriteString(" '")
		sb.WriteString(name)
	}
	sb.WriteString(")\n")
	sb.WriteString(")")

	return sb.String(), nil
}

func (c *AnnYailConverter) genComponents(parentName string, children []Component, sb *strings.Builder, componentNames *[]string) {
	for _, child := range children {
		compId := child.Id
		if compId == "" {
			count := c.autoIdCount[child.Type] + 1
			c.autoIdCount[child.Type] = count
			compId = child.Type + strconv.Itoa(count)
		}
		*componentNames = append(*componentNames, compId)

		sb.WriteString("(add-component ")
		sb.WriteString(parentName)
		sb.WriteString(" ")
		sb.WriteString(yailComponentPkg)
		sb.WriteString(child.Type)
		sb.WriteString(" ")
		sb.WriteString(compId)
		sb.WriteString("\n")

		for k, v := range child.Properties {
			sb.WriteString("  (set-and-coerce-property! '")
			sb.WriteString(compId)
			sb.WriteString(" '")
			sb.WriteString(k)
			sb.WriteString(" ")
			sb.WriteString(annFormatValue(child.Type, k, v))
			sb.WriteString(" ")
			sb.WriteString(annYailType(child.Type, k, v))
			sb.WriteString(")\n")
		}

		c.genComponents(compId, child.Children, sb, componentNames)

		sb.WriteString(")\n")
	}
}

func annYailType(compType, propName, v string) string {
	switch compdb.GlobalDB.GetPropType(dbCompType(compType), propName) {
	case "boolean":
		return "'boolean"
	case "number":
		return "'number"
	case "text":
		return "'text"
	case "list":
		return "'list"
	case "component":
		return "'component"
	}

	if isAnnColorProperty(propName) {
		if _, ok, err := annColorLiteralToIntString(v); ok && err == nil {
			return "'number"
		}
	}
	if strings.EqualFold(v, "true") || strings.EqualFold(v, "false") {
		return "'boolean"
	}
	if _, err := strconv.ParseFloat(v, 64); err == nil {
		return "'number"
	}
	return "'text"
}

func annFormatValue(compType, propName, v string) string {
	switch compdb.GlobalDB.GetPropType(dbCompType(compType), propName) {
	case "boolean":
		if strings.EqualFold(v, "true") || v == "1" {
			return "#t"
		}
		if strings.EqualFold(v, "false") || v == "0" {
			return "#f"
		}
		return annQuoteStr(v)
	case "number":
		if isAnnColorProperty(propName) {
			if colorInt, ok, err := annColorLiteralToIntString(v); ok && err == nil {
				return colorInt
			}
		}
		if _, err := strconv.ParseFloat(v, 64); err == nil {
			return v
		}
		return annQuoteStr(v)
	case "text", "list", "component":
		return annQuoteStr(v)
	}

	if isAnnColorProperty(propName) {
		if colorInt, ok, err := annColorLiteralToIntString(v); ok && err == nil {
			return colorInt
		}
	}
	if strings.EqualFold(v, "true") {
		return "#t"
	}
	if strings.EqualFold(v, "false") {
		return "#f"
	}
	if _, err := strconv.ParseFloat(v, 64); err == nil {
		return v
	}
	return annQuoteStr(v)
}

func annQuoteStr(s string) string {
	var sb strings.Builder
	sb.WriteByte('"')
	for _, c := range s {
		switch c {
		case '\\':
			sb.WriteString(`\\`)
		case '"':
			sb.WriteString(`\"`)
		case '\n':
			sb.WriteString(`\n`)
		case '\t':
			sb.WriteString(`\t`)
		case '\r':
			sb.WriteString(`\r`)
		default:
			if c >= 32 && c <= 126 {
				sb.WriteRune(c)
				continue
			}
			if c <= 0xffff {
				sb.WriteString(`\u`)
				sb.WriteString(fmt.Sprintf("%04x", c))
				continue
			}
			pair := utf16.Encode([]rune{c})
			for _, part := range pair {
				sb.WriteString(`\u`)
				sb.WriteString(fmt.Sprintf("%04x", part))
			}
		}
	}
	sb.WriteByte('"')
	return sb.String()
}
