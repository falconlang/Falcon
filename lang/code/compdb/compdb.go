package compdb

import (
	"Falcon/code/jsonutil"
	_ "embed"
	"errors"
	"strconv"
	"strings"
)

//go:embed simple_components.json
var simpleComponentsJSON []byte

// ── JSON types ────────────────────────────────────────────────────────────────

type scComponent struct {
	Name            string
	Type            string
	Version         string
	CategoryString  string
	NonVisible      string
	HelpString      string
	ShowOnPalette   string
	BlockProperties []scProperty
	Properties      []scDesignerProperty
	Methods         []scMethod
	Events          []scEvent
}

type scProperty struct {
	Name        string
	Type        string
	RW          string
	Deprecated  string
	Description string
	Category    string
	Helper      *scHelper
}

type scDesignerProperty struct {
	Name         string
	EditorType   string
	DefaultValue string
}

type scMethod struct {
	Name         string
	Deprecated   string
	Description  string
	Continuation bool
	ReturnType   string
	Params       []scParam
	Helper       *scHelper
}

type scEvent struct {
	Name        string
	Deprecated  string
	Description string
	Params      []scParam
}

type scParam struct {
	Name        string
	Type        string
	Description string
	Helper      *scHelper
}

type scHelper struct {
	Type string
	Data scHelperData
}

type scHelperData struct {
	ClassName      string
	Key            string
	UnderlyingType string
	Options        []scOption
}

type scOption struct {
	Name  string
	Value string
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

// MethodParam holds the canonical name and YAIL type for a component method parameter.
type MethodParam struct {
	Name string
	Type string
}

// MethodDef holds the resolved metadata needed to validate a component method.
type MethodDef struct {
	Params     []MethodParam
	ReturnType string
}

// ── CompDB ────────────────────────────────────────────────────────────────────

type CompDB struct {
	fqcn               map[string]string
	propType           map[string]map[string]string
	methodContinuation map[string]bool
	optionLists        map[string]*OptionList
	events             map[string]EventDef
	components         map[string]scComponent
	rawJSON            map[string][]byte // cached serialised JSON per component for DescribeComponent
}

// GlobalDB is the singleton component database, initialized once at package load.
var GlobalDB = initCompDB()

func initCompDB() *CompDB {
	parsed, err := jsonutil.ParseAny(simpleComponentsJSON)
	if err != nil {
		panic("compdb: failed to parse simple_components.json: " + err.Error())
	}
	arr, ok := parsed.([]interface{})
	if !ok {
		panic("compdb: simple_components.json must be a JSON array")
	}

	db := &CompDB{
		fqcn:               make(map[string]string),
		propType:           make(map[string]map[string]string),
		methodContinuation: make(map[string]bool),
		optionLists:        make(map[string]*OptionList),
		events:             make(map[string]EventDef),
		components:         make(map[string]scComponent, len(arr)),
		rawJSON:            make(map[string][]byte, len(arr)),
	}

	for _, item := range arr {
		obj, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		comp := parseScComponent(obj)

		db.fqcn[comp.Name] = comp.Type

		pm := make(map[string]string, len(comp.BlockProperties))
		for _, prop := range comp.BlockProperties {
			pm[prop.Name] = prop.Type
			if prop.Helper != nil {
				db.indexHelper(prop.Helper)
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

		// Cache serialised JSON for DescribeComponent
		if raw, err := jsonutil.Marshal(obj); err == nil {
			db.rawJSON[comp.Name] = raw
		}
	}

	return db
}

// ── JSON parsing helpers ──────────────────────────────────────────────────────

func parseScComponent(obj map[string]interface{}) scComponent {
	c := scComponent{
		Name:           jsonutil.GetString(obj, "name"),
		Type:           jsonutil.GetString(obj, "type"),
		Version:        jsonutil.GetString(obj, "version"),
		CategoryString: jsonutil.GetString(obj, "categoryString"),
		NonVisible:     jsonutil.GetString(obj, "nonVisible"),
		HelpString:     jsonutil.GetString(obj, "helpString"),
		ShowOnPalette:  jsonutil.GetString(obj, "showOnPalette"),
	}
	for _, v := range jsonutil.GetArray(obj, "blockProperties") {
		if o, ok := v.(map[string]interface{}); ok {
			c.BlockProperties = append(c.BlockProperties, parseScProperty(o))
		}
	}
	for _, v := range jsonutil.GetArray(obj, "properties") {
		if o, ok := v.(map[string]interface{}); ok {
			c.Properties = append(c.Properties, parseScDesignerProperty(o))
		}
	}
	for _, v := range jsonutil.GetArray(obj, "methods") {
		if o, ok := v.(map[string]interface{}); ok {
			c.Methods = append(c.Methods, parseScMethod(o))
		}
	}
	for _, v := range jsonutil.GetArray(obj, "events") {
		if o, ok := v.(map[string]interface{}); ok {
			c.Events = append(c.Events, parseScEvent(o))
		}
	}
	return c
}

func parseScProperty(obj map[string]interface{}) scProperty {
	p := scProperty{
		Name:        jsonutil.GetString(obj, "name"),
		Type:        jsonutil.GetString(obj, "type"),
		RW:          jsonutil.GetString(obj, "rw"),
		Deprecated:  jsonutil.GetString(obj, "deprecated"),
		Description: jsonutil.GetString(obj, "description"),
		Category:    jsonutil.GetString(obj, "category"),
	}
	if h := jsonutil.GetObject(obj, "helper"); h != nil {
		helper := parseScHelper(h)
		p.Helper = &helper
	}
	return p
}

func parseScDesignerProperty(obj map[string]interface{}) scDesignerProperty {
	return scDesignerProperty{
		Name:         jsonutil.GetString(obj, "name"),
		EditorType:   jsonutil.GetString(obj, "editorType"),
		DefaultValue: jsonutil.GetString(obj, "defaultValue"),
	}
}

func parseScMethod(obj map[string]interface{}) scMethod {
	m := scMethod{
		Name:         jsonutil.GetString(obj, "name"),
		Deprecated:   jsonutil.GetString(obj, "deprecated"),
		Description:  jsonutil.GetString(obj, "description"),
		Continuation: jsonutil.GetBool(obj, "continuation"),
		ReturnType:   jsonutil.GetString(obj, "returnType"),
	}
	for _, v := range jsonutil.GetArray(obj, "params") {
		if o, ok := v.(map[string]interface{}); ok {
			m.Params = append(m.Params, parseScParam(o))
		}
	}
	if h := jsonutil.GetObject(obj, "helper"); h != nil {
		helper := parseScHelper(h)
		m.Helper = &helper
	}
	return m
}

func parseScEvent(obj map[string]interface{}) scEvent {
	e := scEvent{
		Name:        jsonutil.GetString(obj, "name"),
		Deprecated:  jsonutil.GetString(obj, "deprecated"),
		Description: jsonutil.GetString(obj, "description"),
	}
	for _, v := range jsonutil.GetArray(obj, "params") {
		if o, ok := v.(map[string]interface{}); ok {
			e.Params = append(e.Params, parseScParam(o))
		}
	}
	return e
}

func parseScParam(obj map[string]interface{}) scParam {
	p := scParam{
		Name:        jsonutil.GetString(obj, "name"),
		Type:        jsonutil.GetString(obj, "type"),
		Description: jsonutil.GetString(obj, "description"),
	}
	if h := jsonutil.GetObject(obj, "helper"); h != nil {
		helper := parseScHelper(h)
		p.Helper = &helper
	}
	return p
}

func parseScHelper(obj map[string]interface{}) scHelper {
	h := scHelper{
		Type: jsonutil.GetString(obj, "type"),
	}
	if d := jsonutil.GetObject(obj, "data"); d != nil {
		h.Data = scHelperData{
			ClassName:      jsonutil.GetString(d, "className"),
			Key:            jsonutil.GetString(d, "key"),
			UnderlyingType: jsonutil.GetString(d, "underlyingType"),
		}
		for _, v := range jsonutil.GetArray(d, "options") {
			if o, ok := v.(map[string]interface{}); ok {
				h.Data.Options = append(h.Data.Options, scOption{
					Name:  jsonutil.GetString(o, "name"),
					Value: jsonutil.GetString(o, "value"),
				})
			}
		}
	}
	return h
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

func normalizeComponentType(shortName string) string {
	if shortName == "Screen" {
		return "Form"
	}
	return shortName
}

// GetFQCN returns the fully-qualified class name for a component short name.
func (db *CompDB) GetFQCN(shortName string) string {
	shortName = normalizeComponentType(shortName)
	if fqcn := db.fqcn[shortName]; fqcn != "" {
		return fqcn
	}
	return shortName
}

// IsKnownComponent reports whether the component database has metadata for the short name.
func (db *CompDB) IsKnownComponent(shortName string) bool {
	shortName = normalizeComponentType(shortName)
	_, ok := db.components[shortName]
	return ok
}

// ComponentVersion returns the current App Inventor component version.
func (db *CompDB) ComponentVersion(shortName string) string {
	shortName = normalizeComponentType(shortName)
	if comp, ok := db.components[shortName]; ok && comp.Version != "" {
		return comp.Version
	}
	return "1"
}

// IsNonVisible reports whether the component is marked non-visible.
func (db *CompDB) IsNonVisible(shortName string) bool {
	shortName = normalizeComponentType(shortName)
	return db.components[shortName].NonVisible == "true"
}

// CategoryString returns the App Inventor palette category for a component.
func (db *CompDB) CategoryString(shortName string) string {
	shortName = normalizeComponentType(shortName)
	return db.components[shortName].CategoryString
}

// DesignerPropertyDefault returns a designer property's default value.
func (db *CompDB) DesignerPropertyDefault(compType, propName string) (string, bool) {
	compType = normalizeComponentType(compType)
	comp, ok := db.components[compType]
	if !ok {
		return "", false
	}
	for _, prop := range comp.Properties {
		if prop.Name == propName {
			return prop.DefaultValue, true
		}
	}
	return "", false
}

// GetPropType returns the YAIL type for a component property.
func (db *CompDB) GetPropType(compType, propName string) string {
	compType = normalizeComponentType(compType)
	if pm := db.propType[compType]; pm != nil {
		if t := pm[propName]; t != "" {
			return t
		}
	}
	return "any"
}

// IsContinuation reports whether a component method uses blocking-continuation.
func (db *CompDB) IsContinuation(compType, methodName string) bool {
	compType = normalizeComponentType(compType)
	return db.methodContinuation[compType+"."+methodName]
}

// GetEventParams returns the canonical parameter names for a component event.
func (db *CompDB) GetEventParams(compType, eventName string) []string {
	compType = normalizeComponentType(compType)
	def, ok := db.events[compType+"."+eventName]
	if !ok {
		return nil
	}
	params := make([]string, len(def.Params))
	copy(params, def.Params)
	return params
}

// GetMethodParams returns the parameters for a component method.
func (db *CompDB) GetMethodParams(compType, methodName string) []MethodParam {
	compType = normalizeComponentType(compType)
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

func (db *CompDB) getMethodDef(compType, methodName string) (MethodDef, bool) {
	compType = normalizeComponentType(compType)
	comp, ok := db.components[compType]
	if !ok {
		return MethodDef{}, false
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
		return MethodDef{Params: params, ReturnType: method.ReturnType}, true
	}
	return MethodDef{}, false
}

// GetOptionList returns the OptionList for the given key, or nil if not found.
func (db *CompDB) GetOptionList(key string) *OptionList {
	return db.optionLists[key]
}

// ValidateProperty checks that propName exists on compType.
func (db *CompDB) ValidateProperty(compType, propName string) error {
	compType = normalizeComponentType(compType)
	comp, known := db.components[compType]
	if !known {
		return nil
	}
	pm := db.propType[compType]
	if pm[propName] != "" {
		return nil
	}
	for _, prop := range comp.Properties {
		if prop.Name == propName {
			return nil
		}
	}
	lower := strings.ToLower(propName)
	for k := range pm {
		if strings.ToLower(k) == lower {
			return errors.New(compType + ": unknown property " + strconv.Quote(propName) + ", did you mean " + strconv.Quote(k) + "?")
		}
	}
	for _, prop := range comp.Properties {
		if strings.ToLower(prop.Name) == lower {
			return errors.New(compType + ": unknown property " + strconv.Quote(propName) + ", did you mean " + strconv.Quote(prop.Name) + "?")
		}
	}
	return errors.New(compType + ": unknown property " + strconv.Quote(propName))
}

// ValidateMethod checks that methodName exists on compType and receives the
// right number of arguments.
func (db *CompDB) ValidateMethod(compType, methodName string, argsCount int) error {
	compType = normalizeComponentType(compType)
	comp, known := db.components[compType]
	if !known {
		return nil
	}
	def, ok := db.getMethodDef(compType, methodName)
	if !ok {
		lower := strings.ToLower(methodName)
		for _, method := range comp.Methods {
			if strings.ToLower(method.Name) == lower {
				return errors.New(compType + ": unknown method " + strconv.Quote(methodName) + ", did you mean " + strconv.Quote(method.Name) + "?")
			}
		}
		return errors.New(compType + ": unknown method " + strconv.Quote(methodName))
	}
	if len(def.Params) != argsCount {
		if len(def.Params) == 0 {
			return errors.New("method " + compType + "." + methodName + " takes no arguments")
		}
		names := make([]string, len(def.Params))
		for i, param := range def.Params {
			names[i] = param.Name
		}
		return errors.New("method " + compType + "." + methodName + " expects " + strconv.Itoa(len(def.Params)) + " argument(s) (" + strings.Join(names, ", ") + ") but got " + strconv.Itoa(argsCount))
	}
	return nil
}

// DescribeComponent returns the JSON-encoded description of a component.
func (db *CompDB) DescribeComponent(name string) (string, bool) {
	name = normalizeComponentType(name)
	raw, ok := db.rawJSON[name]
	if !ok {
		return "", false
	}
	return string(raw), true
}

// ListComponentNames returns the names of all known component types.
func (db *CompDB) ListComponentNames() []string {
	names := make([]string, 0, len(db.components))
	for name := range db.components {
		names = append(names, name)
	}
	return names
}

// ValidateEvent checks that the given event exists on the component type.
func (db *CompDB) ValidateEvent(compType, eventName string, params []string) error {
	compType = normalizeComponentType(compType)
	def, ok := db.events[compType+"."+eventName]
	if !ok {
		return errors.New("component " + compType + " has no event " + eventName)
	}
	if len(params) != len(def.Params) {
		if len(def.Params) == 0 {
			return errors.New("event " + compType + "." + eventName + " takes no parameters")
		}
		return errors.New("event " + compType + "." + eventName + " expects " + strconv.Itoa(len(def.Params)) + " parameter(s) (" + strings.Join(def.Params, ", ") + ") but got " + strconv.Itoa(len(params)))
	}
	for i, param := range params {
		if param != def.Params[i] {
			return errors.New("event " + compType + "." + eventName + " parameter " + strconv.Itoa(i+1) + " must be " + strconv.Quote(def.Params[i]) + ", got " + strconv.Quote(param))
		}
	}
	return nil
}
