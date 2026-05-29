package compdb

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
)

//go:embed simple_components.json
var simpleComponentsJSON []byte

// ── JSON types ────────────────────────────────────────────────────────────────

type scComponent struct {
	Name            string       `json:"name"`
	Type            string       `json:"type"`
	CategoryString  string       `json:"categoryString"`
	NonVisible      string       `json:"nonVisible"`
	HelpString      string       `json:"helpString"`
	ShowOnPalette   string       `json:"showOnPalette"`
	Properties      []scProperty `json:"properties"`
	BlockProperties []scProperty `json:"blockProperties"`
	Methods         []scMethod   `json:"methods"`
	Events          []scEvent    `json:"events"`
}

type scProperty struct {
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	EditorType  string    `json:"editorType"`
	RW          string    `json:"rw"`
	Deprecated  string    `json:"deprecated"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	Helper      *scHelper `json:"helper,omitempty"`
}

type scMethod struct {
	Name         string    `json:"name"`
	Deprecated   string    `json:"deprecated"`
	Description  string    `json:"description"`
	Continuation bool      `json:"continuation,omitempty"`
	ReturnType   string    `json:"returnType,omitempty"`
	Params       []scParam `json:"params"`
	Helper       *scHelper `json:"helper,omitempty"`
}

type scEvent struct {
	Name        string    `json:"name"`
	Deprecated  string    `json:"deprecated"`
	Description string    `json:"description"`
	Params      []scParam `json:"params"`
}

type scParam struct {
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Description string    `json:"description,omitempty"`
	Helper      *scHelper `json:"helper,omitempty"`
}

type scHelper struct {
	Type string       `json:"type"`
	Data scHelperData `json:"data"`
}

type scHelperData struct {
	ClassName      string     `json:"className"`
	Key            string     `json:"key"`
	UnderlyingType string     `json:"underlyingType"`
	Options        []scOption `json:"options"`
}

type scOption struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// ── OptionList ────────────────────────────────────────────────────────────────

// OptionList holds the resolved data for a helpers_dropdown option list.
type OptionList struct {
	ClassName      string
	UnderlyingType string
	Options        map[string]string // option name → concrete value
}

// ── EventDef ──────────────────────────────────────────────────────────────────

// EventDef holds the canonical parameter names for a component event.
type EventDef struct {
	Params []string
}

// MethodParam holds the canonical name and YAIL type for a component method
// parameter.
type MethodParam struct {
	Name string
	Type string
}

// ── CompDB ────────────────────────────────────────────────────────────────────

type CompDB struct {
	fqcn               map[string]string            // shortName → FQCN
	propType           map[string]map[string]string // shortName → propName → YAIL type
	methodContinuation map[string]bool              // "CompName.MethodName" → true
	optionLists        map[string]*OptionList       // option list key → OptionList
	events             map[string]EventDef          // "CompName.EventName" → EventDef
	components         map[string]scComponent       // shortName → full component data
}

var (
	canvasChildTypes = map[string]bool{
		"Ball":        true,
		"ImageSprite": true,
	}
	mapChildTypes = map[string]bool{
		"Circle":            true,
		"FeatureCollection": true,
		"LineString":        true,
		"Marker":            true,
		"Polygon":           true,
		"Rectangle":         true,
	}
	featureCollectionChildTypes = map[string]bool{
		"Circle":     true,
		"LineString": true,
		"Marker":     true,
		"Polygon":    true,
		"Rectangle":  true,
	}
	chartChildTypes = map[string]bool{
		"ChartData2D": true,
		"Trendline":   true,
	}
)

// GlobalDB is the singleton component database, initialized once at package load.
var GlobalDB = initCompDB()

func initCompDB() *CompDB {
	var components []scComponent
	if err := json.Unmarshal(simpleComponentsJSON, &components); err != nil {
		panic("compdb: failed to parse simple_components.json: " + err.Error())
	}

	db := &CompDB{
		fqcn:               make(map[string]string),
		propType:           make(map[string]map[string]string),
		methodContinuation: make(map[string]bool),
		optionLists:        make(map[string]*OptionList),
		events:             make(map[string]EventDef),
		components:         make(map[string]scComponent, len(components)),
	}

	for _, comp := range components {
		db.fqcn[comp.Name] = comp.Type

		pm := make(map[string]string, len(comp.BlockProperties))
		for _, prop := range comp.BlockProperties {
			pm[prop.Name] = prop.Type
			if prop.Helper != nil {
				db.indexHelper(prop.Helper)
			}
		}
		for _, prop := range comp.Properties {
			if pm[prop.Name] == "" {
				pm[prop.Name] = inferDesignerPropertyType(prop)
			}
		}
		db.propType[comp.Name] = pm

		for _, method := range comp.Methods {
			if method.Continuation {
				db.methodContinuation[comp.Name+"."+method.Name] = true
			}
			if method.Helper != nil {
				db.indexHelper(method.Helper)
			}
			for _, param := range method.Params {
				if param.Helper != nil {
					db.indexHelper(param.Helper)
				}
			}
		}

		for _, event := range comp.Events {
			names := make([]string, len(event.Params))
			for i, p := range event.Params {
				names[i] = p.Name
			}
			db.events[comp.Name+"."+event.Name] = EventDef{Params: names}
		}

		db.components[comp.Name] = comp
	}

	return db
}

func inferDesignerPropertyType(prop scProperty) string {
	switch prop.Type {
	case "boolean", "number", "text", "list", "component", "any":
		return prop.Type
	}
	switch prop.EditorType {
	case "boolean":
		return "boolean"
	case "color", "float", "integer", "layout_size", "non_negative_float", "non_negative_integer":
		return "number"
	case "ListViewAddData":
		return "list"
	}
	return "text"
}

func (db *CompDB) indexHelper(h *scHelper) {
	if h.Type != "OPTION_LIST" {
		return
	}
	key := h.Data.Key
	if key == "" || db.optionLists[key] != nil {
		return
	}
	opts := make(map[string]string, len(h.Data.Options))
	for _, opt := range h.Data.Options {
		opts[opt.Name] = opt.Value
	}
	db.optionLists[key] = &OptionList{
		ClassName:      h.Data.ClassName,
		UnderlyingType: h.Data.UnderlyingType,
		Options:        opts,
	}
}

// ── Lookup methods ────────────────────────────────────────────────────────────

// GetFQCN returns the fully-qualified class name for a component short name,
// falling back to the short name itself if not found.
func (db *CompDB) GetFQCN(shortName string) string {
	if fqcn := db.fqcn[shortName]; fqcn != "" {
		return fqcn
	}
	return shortName
}

// GetPropType returns the YAIL type for a component property (e.g. "number",
// "text", "boolean"), falling back to "any".
func (db *CompDB) GetPropType(compType, propName string) string {
	if pm := db.propType[compType]; pm != nil {
		if t := pm[propName]; t != "" {
			return t
		}
	}
	return "any"
}

// IsContinuation reports whether a component method uses blocking-continuation.
func (db *CompDB) IsContinuation(compType, methodName string) bool {
	return db.methodContinuation[compType+"."+methodName]
}

// GetMethodParams returns the parameters for a component method, or nil if the
// component or method is unknown.
func (db *CompDB) GetMethodParams(compType, methodName string) []MethodParam {
	comp, ok := db.components[compType]
	if !ok {
		return nil
	}
	for _, method := range comp.Methods {
		if method.Name != methodName {
			continue
		}
		params := make([]MethodParam, len(method.Params))
		for i, param := range method.Params {
			paramType := param.Type
			if paramType == "" {
				paramType = "any"
			}
			params[i] = MethodParam{Name: param.Name, Type: paramType}
		}
		return params
	}
	return nil
}

// GetOptionList returns the OptionList for the given key, or nil if not found.
func (db *CompDB) GetOptionList(key string) *OptionList {
	return db.optionLists[key]
}

// HasComponent reports whether compType is a known built-in App Inventor
// component type.
func (db *CompDB) HasComponent(compType string) bool {
	_, ok := db.components[compType]
	return ok
}

// IsNonVisible reports whether compType is a known non-visible component.
func (db *CompDB) IsNonVisible(compType string) bool {
	if compType == "Form" {
		return false
	}
	comp, ok := db.components[compType]
	return ok && comp.NonVisible == "true"
}

// CanContainComponent applies the designer's structural placement rules for
// built-in components.
func (db *CompDB) CanContainComponent(parentType, childType string) bool {
	if parentType == "" || childType == "" {
		return false
	}
	if canvasChildTypes[childType] {
		return parentType == "Canvas"
	}
	if mapChildTypes[childType] {
		if parentType == "Map" {
			return true
		}
		return parentType == "FeatureCollection" && featureCollectionChildTypes[childType]
	}
	if chartChildTypes[childType] {
		return parentType == "Chart"
	}
	if db.IsNonVisible(childType) {
		return parentType == "Form"
	}
	if db.IsNonVisible(parentType) {
		return false
	}
	switch parentType {
	case "Form":
		return true
	case "Canvas", "Map", "FeatureCollection", "Chart":
		return false
	}
	return db.components[parentType].CategoryString == "LAYOUT"
}

// ValidateProperty checks that propName exists on compType. Returns nil when the
// component type is unknown (can't validate extensions). When the property is not
// found, the error includes a "did you mean" hint if a case-insensitive match exists.
func (db *CompDB) ValidateProperty(compType, propName string) error {
	pm := db.propType[compType]
	if pm == nil {
		return nil // unknown component — extension or unmapped type
	}
	if pm[propName] != "" {
		return nil
	}
	lower := strings.ToLower(propName)
	for k := range pm {
		if strings.ToLower(k) == lower {
			return fmt.Errorf("%s: unknown property %q, did you mean %q?", compType, propName, k)
		}
	}
	return fmt.Errorf("%s: unknown property %q", compType, propName)
}

// DescribeComponent returns the JSON-encoded description of a component
// (properties, methods, events with full metadata), or ("", false) if unknown.
func (db *CompDB) DescribeComponent(name string) (string, bool) {
	comp, ok := db.components[name]
	if !ok {
		return "", false
	}
	data, err := json.Marshal(comp)
	if err != nil {
		return "", false
	}
	return string(data), true
}

// ListComponentNames returns the names of all known component types.
func (db *CompDB) ListComponentNames() []string {
	names := make([]string, 0, len(db.components))
	for name := range db.components {
		names = append(names, name)
	}
	return names
}

// ValidateEvent checks that the given event exists on the component type and
// that the number of declared parameters matches. Returns a descriptive error
// including the canonical parameter names if either check fails.
func (db *CompDB) ValidateEvent(compType, eventName string, params []string) error {
	def, ok := db.events[compType+"."+eventName]
	if !ok {
		return fmt.Errorf("component %s has no event %s", compType, eventName)
	}
	if len(params) != len(def.Params) {
		if len(def.Params) == 0 {
			return fmt.Errorf("event %s.%s takes no parameters", compType, eventName)
		}
		return fmt.Errorf("event %s.%s expects %d parameter(s) (%s) but got %d",
			compType, eventName,
			len(def.Params), strings.Join(def.Params, ", "),
			len(params))
	}
	return nil
}
