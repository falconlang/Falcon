package design

import (
	"Falcon/code/compdb"
	"errors"
	"fmt"
	"strconv"
	"strings"
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
	var errs []error
	collectPropertyErrors(screen.Type, screen.Id, screen.Properties, &errs)
	collectChildErrors(screen.Children, &errs)
	return errors.Join(errs...)
}

func collectPropertyErrors(compType, compId string, props map[string]string, errs *[]error) {
	for k := range props {
		if err := compdb.GlobalDB.ValidateProperty(dbCompType(compType), k); err != nil {
			*errs = append(*errs, fmt.Errorf("%s %q: %w", compType, compId, err))
		}
	}
}

func collectChildErrors(children []Component, errs *[]error) {
	for _, child := range children {
		id := child.Id
		if id == "" {
			id = child.Type
		}
		collectPropertyErrors(child.Type, id, child.Properties, errs)
		collectChildErrors(child.Children, errs)
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
			sb.WriteString(annFormatValue(v))
			sb.WriteString(" ")
			sb.WriteString(annYailType(v))
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
			sb.WriteString(annFormatValue(v))
			sb.WriteString(" ")
			sb.WriteString(annYailType(v))
			sb.WriteString(")\n")
		}

		c.genComponents(compId, child.Children, sb, componentNames)

		sb.WriteString(")\n")
	}
}

func annYailType(v string) string {
	if v == "true" || v == "false" {
		return "'boolean"
	}
	if _, err := strconv.ParseFloat(v, 64); err == nil {
		return "'number"
	}
	return "'text"
}

func annFormatValue(v string) string {
	if v == "true" {
		return "#t"
	}
	if v == "false" {
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
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == '\\':
			if i+1 < len(s) && (s[i+1] == 'n' || s[i+1] == 't' || s[i+1] == 'r') {
				sb.WriteByte(c)
			} else {
				sb.WriteString(`\\`)
			}
		case c == '"':
			sb.WriteString(`\"`)
		case c >= 32 && c <= 126:
			sb.WriteByte(c)
		default:
			hex := "000" + strconv.FormatInt(int64(c), 16)
			sb.WriteString(`\u`)
			sb.WriteString(hex[len(hex)-4:])
		}
	}
	sb.WriteByte('"')
	return sb.String()
}
