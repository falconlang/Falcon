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
	BlockProperties []scProperty `json:"blockProperties"`
	Methods         []scMethod   `json:"methods"`
	Events          []scEvent    `json:"events"`
}

type scProperty struct {
	Name   string    `json:"name"`
	Type   string    `json:"type"`
	Helper *scHelper `json:"helper,omitempty"`
}

type scMethod struct {
	Name         string    `json:"name"`
	Continuation bool      `json:"continuation,omitempty"`
	Params       []scParam `json:"params"`
	Helper       *scHelper `json:"helper,omitempty"`
}

type scEvent struct {
	Name   string    `json:"name"`
	Params []scParam `json:"params"`
}

type scParam struct {
	Name   string    `json:"name"`
	Type   string    `json:"type"`
	Helper *scHelper `json:"helper,omitempty"`
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

// ── CompDB ────────────────────────────────────────────────────────────────────

type CompDB struct {
	fqcn               map[string]string            // shortName → FQCN
	propType           map[string]map[string]string // shortName → propName → YAIL type
	methodContinuation map[string]bool              // "CompName.MethodName" → true
	optionLists        map[string]*OptionList       // option list key → OptionList
	events             map[string]EventDef          // "CompName.EventName" → EventDef
}

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
	}

	return db
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

// GetOptionList returns the OptionList for the given key, or nil if not found.
func (db *CompDB) GetOptionList(key string) *OptionList {
	return db.optionLists[key]
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
