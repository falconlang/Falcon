package design

import (
	"Falcon/code/compdb"
	"sort"
	"strconv"
	"strings"
	"unicode/utf16"
)

const yailComponentPkg = "com.google.appinventor.components.runtime."
const appInventorYaVersion = "233"

var suppressedDesignerProperties = map[string]bool{
	"Uuid":          true,
	"TutorialURL":   true,
	"BlocksToolkit": true,
}

var alwaysSendDesignerDefaults = map[string][]string{
	"Form": {"ShowListsAsJson", "Sizing"},
}

var canvasChildTypes = map[string]bool{
	"Ball":        true,
	"ImageSprite": true,
}

var mapChildTypes = map[string]bool{
	"Circle":            true,
	"FeatureCollection": true,
	"LineString":        true,
	"Marker":            true,
	"Polygon":           true,
	"Rectangle":         true,
}

var featureCollectionChildTypes = map[string]bool{
	"Circle":     true,
	"LineString": true,
	"Marker":     true,
	"Polygon":    true,
	"Rectangle":  true,
}

var chartChildTypes = map[string]bool{
	"ChartData2D": true,
	"Trendline":   true,
}

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

func appInventorComponentVersion(annType string) string {
	return compdb.GlobalDB.ComponentVersion(dbCompType(annType))
}

func shouldSendDesignerProperty(propName string) bool {
	return !strings.HasPrefix(propName, "$") && !suppressedDesignerProperties[propName]
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
	collectAnnErrors(screen, &diagnostics)
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

func collectAnnErrors(screen Component, diagnostics *[]AnnDiagnostic) {
	if screen.Type != "Screen" && screen.Type != "Form" {
		*diagnostics = append(*diagnostics, AnnDiagnostic{
			Message:  "the root designer component must be Screen",
			Position: screen.typePosition,
			Length:   len([]rune(screen.Type)),
		})
	}
	seen := make(map[string]AnnDiagnostic)
	autoIds := make(map[string]int)
	componentIds := make(map[string]bool)
	collectComponentIds(screen, nil, componentIds, make(map[string]int))
	collectComponentErrors(screen, nil, diagnostics, seen, autoIds, componentIds)
}

func collectComponentIds(component Component, parent *Component, ids map[string]bool, autoIds map[string]int) {
	compId := component.Id
	if compId == "" {
		if parent == nil {
			compId = "Screen1"
		} else {
			count := autoIds[component.Type] + 1
			autoIds[component.Type] = count
			compId = component.Type + strconv.Itoa(count)
		}
	}
	ids[compId] = true
	for _, child := range component.Children {
		collectComponentIds(child, &component, ids, autoIds)
	}
}

func collectComponentErrors(component Component, parent *Component, diagnostics *[]AnnDiagnostic, seen map[string]AnnDiagnostic, autoIds map[string]int, componentIds map[string]bool) {
	compId := component.Id
	if compId == "" {
		if parent == nil {
			compId = "Screen1"
		} else {
			count := autoIds[component.Type] + 1
			autoIds[component.Type] = count
			compId = component.Type + strconv.Itoa(count)
		}
	}

	dbType := dbCompType(component.Type)
	if !compdb.GlobalDB.IsKnownComponent(dbType) {
		*diagnostics = append(*diagnostics, AnnDiagnostic{
			Message:  "unknown component type " + strconv.Quote(component.Type),
			Position: component.typePosition,
			Length:   len([]rune(component.Type)),
		})
	}

	if parent != nil && !canContainAnnComponent(parent.Type, component.Type) {
		parentId := parent.Id
		if parentId == "" {
			parentId = parent.Type
		}
		*diagnostics = append(*diagnostics, AnnDiagnostic{
			Message:  component.Type + "." + compId + " cannot be placed inside " + parent.Type + "." + parentId,
			Position: component.typePosition,
			Length:   len([]rune(component.Type)),
		})
	}

	if previous, ok := seen[compId]; ok {
		position := component.idPosition
		length := len([]rune(compId))
		if position == 0 {
			position = component.typePosition
			length = len([]rune(component.Type))
		}
		*diagnostics = append(*diagnostics, AnnDiagnostic{
			Message:  "duplicate component name " + strconv.Quote(compId) + "; first used at position " + strconv.Itoa(previous.Position),
			Position: position,
			Length:   length,
		})
	} else {
		position := component.idPosition
		length := len([]rune(compId))
		if position == 0 {
			position = component.typePosition
			length = len([]rune(component.Type))
		}
		seen[compId] = AnnDiagnostic{Position: position, Length: length}
	}

	for propName, propValue := range component.Properties {
		if !shouldSendDesignerProperty(propName) {
			continue
		}
		if err := compdb.GlobalDB.ValidateProperty(dbCompType(component.Type), propName); err != nil {
			message := component.Type + " " + strconv.Quote(compId) + ": " + err.Error()
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
			continue
		}
		if compdb.GlobalDB.GetPropType(dbCompType(component.Type), propName) == "component" {
			target := strings.TrimSpace(propValue)
			if target != "" && target != "null" && !componentIds[target] {
				position := component.typePosition
				if propPosition, ok := component.propertyPositions[propName]; ok {
					position = propPosition
				}
				length := len([]rune(propName))
				if propLength, ok := component.propertyLengths[propName]; ok && propLength > 0 {
					length = propLength
				}
				*diagnostics = append(*diagnostics, AnnDiagnostic{
					Message:  component.Type + " " + strconv.Quote(compId) + ": component property " + strconv.Quote(propName) + " references unknown component " + strconv.Quote(target),
					Position: position,
					Length:   length,
				})
			}
		}
	}

	for _, child := range component.Children {
		collectComponentErrors(child, &component, diagnostics, seen, autoIds, componentIds)
	}
}

func canContainAnnComponent(parentType, childType string) bool {
	parent := dbCompType(parentType)
	child := dbCompType(childType)
	if canvasChildTypes[child] {
		return parent == "Canvas"
	}
	if mapChildTypes[child] {
		return parent == "Map" || (parent == "FeatureCollection" && featureCollectionChildTypes[child])
	}
	if chartChildTypes[child] {
		return parent == "Chart"
	}
	if compdb.GlobalDB.IsNonVisible(child) {
		return parent == "Form"
	}
	if compdb.GlobalDB.IsNonVisible(parent) {
		return false
	}
	if parent == "Canvas" || parent == "Map" || parent == "FeatureCollection" || parent == "Chart" {
		return false
	}
	return parent == "Form" ||
		parent == "ScrollHorizontal" ||
		parent == "ScrollVertical" ||
		compdb.GlobalDB.CategoryString(parent) == "LAYOUT"
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

	screenProperties := sortedSendableProperties(screen)
	if len(screenProperties) > 0 {
		sb.WriteString("(do-after-form-creation")
		for _, entry := range screenProperties {
			sb.WriteString("\n  (set-and-coerce-property! '")
			sb.WriteString(formName)
			sb.WriteString(" '")
			sb.WriteString(entry.name)
			sb.WriteString(" ")
			sb.WriteString(annFormatValue(screen.Type, entry.name, entry.value))
			sb.WriteString(" ")
			sb.WriteString(annYailType(screen.Type, entry.name))
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

type annPropertyEntry struct {
	name  string
	value string
}

func sortedSendableProperties(component Component) []annPropertyEntry {
	values := make(map[string]string, len(component.Properties))
	for k, v := range component.Properties {
		if shouldSendDesignerProperty(k) {
			values[k] = v
		}
	}

	dbType := dbCompType(component.Type)
	for _, propName := range alwaysSendDesignerDefaults[dbType] {
		if _, ok := values[propName]; ok {
			continue
		}
		if defaultValue, ok := compdb.GlobalDB.DesignerPropertyDefault(dbType, propName); ok {
			values[propName] = defaultValue
		}
	}

	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	entries := make([]annPropertyEntry, 0, len(keys))
	for _, k := range keys {
		entries = append(entries, annPropertyEntry{name: k, value: values[k]})
	}
	return entries
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

		for _, entry := range sortedSendableProperties(child) {
			sb.WriteString("  (set-and-coerce-property! '")
			sb.WriteString(compId)
			sb.WriteString(" '")
			sb.WriteString(entry.name)
			sb.WriteString(" ")
			sb.WriteString(annFormatValue(child.Type, entry.name, entry.value))
			sb.WriteString(" ")
			sb.WriteString(annYailType(child.Type, entry.name))
			sb.WriteString(")\n")
		}

		sb.WriteString(")\n")
		c.genComponents(compId, child.Children, sb, componentNames)
	}
}

func annYailType(componentType, propName string) string {
	propType := compdb.GlobalDB.GetPropType(dbCompType(componentType), propName)
	return "'" + propType
}

func annFormatValue(componentType, propName, v string) string {
	propType := compdb.GlobalDB.GetPropType(dbCompType(componentType), propName)
	switch propType {
	case "number":
		if isAnnNumberLiteral(v) {
			return v
		}
		if strings.HasPrefix(strings.ToLower(v), "&h") && len(v) > 2 {
			return "#x" + v[2:]
		}
	case "boolean":
		if strings.EqualFold(v, "true") || v == "1" {
			return "#t"
		}
		if strings.EqualFold(v, "false") || v == "0" {
			return "#f"
		}
	case "component":
		if v == "" {
			return `""`
		}
		return "(get-component " + v + ")"
	}
	if v == "" || v == "null" {
		return `""`
	}
	return annQuoteStr(v)
}

func isAnnNumberLiteral(v string) bool {
	if v == "" {
		return false
	}
	_, err := strconv.ParseFloat(v, 64)
	return err == nil
}

func annQuoteStr(s string) string {
	var sb strings.Builder
	sb.WriteByte('"')
	for _, c := range s {
		switch {
		case c == '\\':
			sb.WriteString(`\\`)
		case c == '"':
			sb.WriteString(`\"`)
		case c >= 32 && c <= 126:
			sb.WriteRune(c)
		default:
			for _, codeUnit := range utf16.Encode([]rune{c}) {
				sb.WriteString(`\u`)
				sb.WriteString(lowerHex4(codeUnit))
			}
		}
	}
	sb.WriteByte('"')
	return sb.String()
}
