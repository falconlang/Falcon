//go:build js && wasm
// +build js,wasm

package main

import (
	"Falcon/code/ast"
	"Falcon/code/ast/components"
	"Falcon/code/ast/variables"
	"Falcon/code/context"
	"Falcon/code/lex"
	"Falcon/code/runtime"
	"math"
	"sort"
	"strconv"
	"strings"
	"syscall/js"
	"time"
)

type simulationEventKey struct {
	component string
	event     string
}

type simulationGenericEventKey struct {
	componentType string
	event         string
}

type simulationSession struct {
	id          int
	interp      *runtime.Interpreter
	host        *simulationHost
	events      map[simulationEventKey]*components.Event
	generic     map[simulationGenericEventKey]*components.GenericEvent
	diagnostics []wasmDiagnostic
}

type simulationHost struct {
	state          map[string]map[string]runtime.Value
	statePatch     map[string]map[string]any
	effects        []map[string]any
	logs           []string
	unsupported    []map[string]any
	tinyDB         map[string]map[string]runtime.Value
	componentTypes map[string]string
	runEvent       func(componentName, componentType, eventName string, args []runtime.Value) bool

	// Property patches originating from browser input already carry their event.
	suppressTextChanged bool
}

var simulationSessions = map[int]*simulationSession{}
var nextSimulationSessionID = 1

func newSimulationHost(initial map[string]map[string]runtime.Value, componentTypes map[string]string, tinyDBInitial map[string]map[string]runtime.Value) *simulationHost {
	state := make(map[string]map[string]runtime.Value)
	for component, props := range initial {
		state[component] = make(map[string]runtime.Value)
		for prop, value := range props {
			state[component][prop] = value
		}
	}
	typeCopy := make(map[string]string)
	for component, componentType := range componentTypes {
		typeCopy[component] = componentType
	}
	tinyDB := make(map[string]map[string]runtime.Value)
	for namespace, store := range tinyDBInitial {
		tinyDB[namespace] = make(map[string]runtime.Value)
		for tag, value := range store {
			tinyDB[namespace][tag] = copyRuntimeValue(value)
		}
	}
	return &simulationHost{
		state:          state,
		statePatch:     map[string]map[string]any{},
		tinyDB:         tinyDB,
		componentTypes: typeCopy,
	}
}

func (h *simulationHost) GetProperty(componentName, componentType, property string) runtime.Value {
	componentType = h.componentType(componentName, componentType)
	if props, ok := h.state[componentName]; ok {
		if value, ok := props[property]; ok {
			return value
		}
	}
	switch componentType {
	case "DatePicker":
		if property == "MonthInText" || property == "Instant" {
			if year, month, day, ok := h.currentDateParts(componentName); ok {
				if property == "MonthInText" {
					return runtime.StrVal(monthName(month))
				}
				return runtime.StrVal(dateInstantString(year, month, day))
			}
		}
	case "TimePicker":
		if property == "Instant" {
			if hour, minute, ok := h.currentTimeParts(componentName); ok {
				return runtime.StrVal(timeInstantString(hour, minute))
			}
		}
	case "Screen", "Form":
		switch property {
		case "Platform":
			return runtime.StrVal("Android")
		case "PlatformVersion":
			return runtime.StrVal("simulator")
		}
	}
	return runtime.NullVal()
}

func (h *simulationHost) SetProperty(componentName, componentType, property string, value runtime.Value) {
	componentType = h.componentType(componentName, componentType)
	if property == "ElementsFromString" {
		h.setProperty(componentName, "ElementsFromString", value)
		h.setElements(componentName, componentType, runtime.ListVal(elementsFromString(value.AsStr())))
		return
	}
	if property == "Elements" {
		h.setElements(componentName, componentType, value)
		return
	}
	if property == "Selection" {
		h.setSelection(componentName, componentType, value)
		return
	}
	if property == "SelectionIndex" {
		h.setSelectionIndex(componentName, componentType, value)
		return
	}
	if componentType == "Slider" {
		switch property {
		case "MinValue", "MaxValue", "ThumbPosition", "NumberOfSteps":
			h.setSliderProperty(componentName, property, value)
			return
		}
	}
	if componentType == "DatePicker" {
		switch property {
		case "Year", "Month", "Day":
			h.setDateProperty(componentName, property, value)
			return
		}
	}
	if componentType == "TimePicker" {
		switch property {
		case "Hour", "Minute":
			h.setTimeProperty(componentName, property, value)
			return
		}
	}
	if (componentType == "TextBox" || componentType == "PasswordTextBox") && property == "Text" {
		changed := h.GetProperty(componentName, componentType, "Text").AsStr() != value.AsStr()
		h.setProperty(componentName, property, value)
		if changed && !h.suppressTextChanged && h.runEvent != nil {
			h.runEvent(componentName, componentType, "TextChanged", nil)
		}
		return
	}
	if componentType == "LinearProgress" && property == "Progress" {
		min := valueAsNumber(h.GetProperty(componentName, componentType, "Minimum"), 0)
		max := valueAsNumber(h.GetProperty(componentName, componentType, "Maximum"), 100)
		clamped := clampFloat(value.AsNum(), min, max)
		prev := valueAsNumber(h.GetProperty(componentName, componentType, "Progress"), 0)
		h.setProperty(componentName, property, runtime.NumVal(clamped))
		if clamped != prev && !h.suppressTextChanged && h.runEvent != nil {
			h.runEvent(componentName, componentType, "ProgressChanged", []runtime.Value{runtime.NumVal(clamped)})
		}
		return
	}
	if componentType == "WebViewer" && property == "HomeUrl" {
		h.setProperty(componentName, property, value)
		h.setProperty(componentName, "CurrentUrl", value)
		h.effects = append(h.effects, componentActionWith(componentName, "navigate", map[string]any{"url": value.AsStr()}))
		return
	}
	if componentType == "WebViewer" && property == "CurrentUrl" {
		h.setProperty(componentName, property, value)
		h.effects = append(h.effects, componentActionWith(componentName, "navigate", map[string]any{"url": value.AsStr()}))
		return
	}
	if (componentType == "CheckBox" && property == "Checked") || (componentType == "Switch" && property == "On") {
		changed := valueAsBool(h.GetProperty(componentName, componentType, property)) != valueAsBool(value)
		h.setProperty(componentName, property, value)
		if changed && !h.suppressTextChanged && h.runEvent != nil {
			h.runEvent(componentName, componentType, "Changed", nil)
		}
		return
	}
	h.setProperty(componentName, property, value)
}

func (h *simulationHost) componentType(componentName, explicit string) string {
	if explicit != "" {
		return explicit
	}
	return h.componentTypes[componentName]
}

func (h *simulationHost) setProperty(componentName, property string, value runtime.Value) {
	if _, ok := h.state[componentName]; !ok {
		h.state[componentName] = map[string]runtime.Value{}
	}
	h.state[componentName][property] = value
	if _, ok := h.statePatch[componentName]; !ok {
		h.statePatch[componentName] = map[string]any{}
	}
	h.statePatch[componentName][property] = runtimeValueToJS(value)
}

func (h *simulationHost) setElements(componentName, componentType string, value runtime.Value) {
	h.setProperty(componentName, "Elements", value)
	elements := valueList(value)
	currentIndex := int(valueAsNumber(h.GetProperty(componentName, componentType, "SelectionIndex"), 0))
	switch componentType {
	case "Spinner":
		if len(elements) == 0 {
			h.setSelectionIndex(componentName, componentType, runtime.NumVal(0))
		} else if currentIndex < 1 {
			// App Inventor's Spinner auto-selects the first item when populated
			// with no prior selection (no AfterSelecting fired).
			h.setSelectionIndex(componentName, componentType, runtime.NumVal(1))
		} else if currentIndex > len(elements) {
			h.setSelectionIndex(componentName, componentType, runtime.NumVal(float64(len(elements))))
		}
	case "ListView":
		h.setSelectionIndex(componentName, componentType, runtime.NumVal(0))
	case "ListPicker":
		currentSelection := h.GetProperty(componentName, componentType, "Selection").AsStr()
		if currentIndex < 1 || currentIndex > len(elements) || (currentSelection != "" && selectionIndexForValue(currentSelection, elements) == 0) {
			h.setSelectionIndex(componentName, componentType, runtime.NumVal(0))
		}
	}
}

func (h *simulationHost) setSelection(componentName, componentType string, value runtime.Value) {
	text := value.AsStr()
	elements := valueList(h.GetProperty(componentName, componentType, "Elements"))
	index := selectionIndexForValue(text, elements)
	switch componentType {
	case "Spinner":
		if index == 0 {
			h.setProperty(componentName, "Selection", runtime.StrVal(""))
			h.setProperty(componentName, "SelectionIndex", runtime.NumVal(0))
			return
		}
		h.setProperty(componentName, "Selection", runtime.StrVal(elements[index-1].AsStr()))
		h.setProperty(componentName, "SelectionIndex", runtime.NumVal(float64(index)))
	case "ListPicker":
		h.setProperty(componentName, "Selection", runtime.StrVal(text))
		h.setProperty(componentName, "SelectionIndex", runtime.NumVal(float64(index)))
	case "ListView":
		h.setProperty(componentName, "Selection", runtime.StrVal(text))
		h.setProperty(componentName, "SelectionIndex", runtime.NumVal(float64(index)))
	default:
		h.setProperty(componentName, "Selection", value)
	}
}

func (h *simulationHost) setSelectionIndex(componentName, componentType string, value runtime.Value) {
	index := int(value.AsNum())
	elements := valueList(h.GetProperty(componentName, componentType, "Elements"))
	if index < 1 || index > len(elements) {
		h.setProperty(componentName, "SelectionIndex", runtime.NumVal(0))
		h.setProperty(componentName, "Selection", runtime.StrVal(""))
		return
	}
	h.setProperty(componentName, "SelectionIndex", runtime.NumVal(float64(index)))
	if componentType == "ListView" {
		h.setProperty(componentName, "Selection", runtime.StrVal(listViewMainText(elements[index-1])))
		return
	}
	h.setProperty(componentName, "Selection", runtime.StrVal(elements[index-1].AsStr()))
}

func selectionIndexForValue(text string, elements []runtime.Value) int {
	for index, item := range elements {
		if item.AsStr() == text || listViewMainText(item) == text {
			return index + 1
		}
	}
	return 0
}

func valueAsNumber(value runtime.Value, fallback float64) float64 {
	if value.Type() == runtime.Null {
		return fallback
	}
	return value.AsNum()
}

func valueAsBool(value runtime.Value) bool {
	switch value.Type() {
	case runtime.Bool:
		return value.AsBool()
	case runtime.Number:
		return value.AsNum() != 0
	}
	switch strings.TrimSpace(strings.ToLower(value.AsStr())) {
	case "true", "t", "yes", "y", "1", "on":
		return true
	default:
		return false
	}
}

func valueAsString(value runtime.Value, fallback string) string {
	if value.Type() == runtime.Null {
		return fallback
	}
	text := value.AsStr()
	if text == "" {
		return fallback
	}
	return text
}

func argValue(args []runtime.Value, index int) runtime.Value {
	if index < 0 || index >= len(args) {
		return runtime.StrVal("")
	}
	return args[index]
}

func argString(args []runtime.Value, index int, fallback string) string {
	if index < 0 || index >= len(args) {
		return fallback
	}
	text := args[index].AsStr()
	if text == "" {
		return fallback
	}
	return text
}

func argBool(args []runtime.Value, index int, fallback bool) bool {
	if index < 0 || index >= len(args) {
		return fallback
	}
	if args[index].Type() == runtime.Bool {
		return args[index].AsBool()
	}
	text := strings.TrimSpace(strings.ToLower(args[index].AsStr()))
	switch text {
	case "true", "1", "yes":
		return true
	case "false", "0", "no":
		return false
	default:
		return fallback
	}
}

func (h *simulationHost) setSliderProperty(componentName, property string, value runtime.Value) {
	minValue := valueAsNumber(h.GetProperty(componentName, "Slider", "MinValue"), 10)
	maxValue := valueAsNumber(h.GetProperty(componentName, "Slider", "MaxValue"), 50)
	thumbPosition := valueAsNumber(h.GetProperty(componentName, "Slider", "ThumbPosition"), 30)

	switch property {
	case "MinValue":
		minValue = value.AsNum()
		if minValue > maxValue {
			maxValue = minValue
		}
		h.setProperty(componentName, "MinValue", runtime.NumVal(minValue))
		if valueAsNumber(h.GetProperty(componentName, "Slider", "MaxValue"), 50) != maxValue {
			h.setProperty(componentName, "MaxValue", runtime.NumVal(maxValue))
		}
	case "MaxValue":
		maxValue = value.AsNum()
		if maxValue < minValue {
			minValue = maxValue
		}
		h.setProperty(componentName, "MaxValue", runtime.NumVal(maxValue))
		if valueAsNumber(h.GetProperty(componentName, "Slider", "MinValue"), 10) != minValue {
			h.setProperty(componentName, "MinValue", runtime.NumVal(minValue))
		}
	case "ThumbPosition":
		thumbPosition = value.AsNum()
	case "NumberOfSteps":
		h.setProperty(componentName, "NumberOfSteps", value)
		return
	default:
		h.setProperty(componentName, property, value)
		return
	}

	clamped := clampFloat(thumbPosition, minValue, maxValue)
	if property == "ThumbPosition" || clamped != valueAsNumber(h.GetProperty(componentName, "Slider", "ThumbPosition"), 30) {
		h.setProperty(componentName, "ThumbPosition", runtime.NumVal(clamped))
	}
}

func clampFloat(value, minValue, maxValue float64) float64 {
	if minValue > maxValue {
		minValue, maxValue = maxValue, minValue
	}
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}

func (h *simulationHost) currentDateParts(componentName string) (int, int, int, bool) {
	year, okYear := intValue(h.GetProperty(componentName, "DatePicker", "Year"))
	month, okMonth := intValue(h.GetProperty(componentName, "DatePicker", "Month"))
	day, okDay := intValue(h.GetProperty(componentName, "DatePicker", "Day"))
	if !okYear || !okMonth || !okDay || !validDateParts(year, month, day) {
		return 0, 0, 0, false
	}
	return year, month, day, true
}

func (h *simulationHost) currentTimeParts(componentName string) (int, int, bool) {
	hour, okHour := intValue(h.GetProperty(componentName, "TimePicker", "Hour"))
	minute, okMinute := intValue(h.GetProperty(componentName, "TimePicker", "Minute"))
	if !okHour || !okMinute || !validTimeParts(hour, minute) {
		return 0, 0, false
	}
	return hour, minute, true
}

func (h *simulationHost) setDateProperty(componentName, property string, value runtime.Value) {
	year, month, day, ok := h.currentDateParts(componentName)
	if !ok {
		year, month, day = 1970, 1, 1
	}
	next, ok := intValue(value)
	if !ok {
		h.logs = append(h.logs, "DatePicker."+property+" ignored non-numeric value: "+value.AsStr())
		return
	}
	switch property {
	case "Year":
		year = next
	case "Month":
		month = next
	case "Day":
		day = next
	}
	if !validDateParts(year, month, day) {
		h.logs = append(h.logs, "DatePicker."+property+" ignored invalid date")
		return
	}
	h.setDateParts(componentName, year, month, day)
}

func (h *simulationHost) setDateToDisplay(componentName string, yearValue, monthValue, dayValue runtime.Value) {
	year, okYear := intValue(yearValue)
	month, okMonth := intValue(monthValue)
	day, okDay := intValue(dayValue)
	if !okYear || !okMonth || !okDay || !validDateParts(year, month, day) {
		h.logs = append(h.logs, "DatePicker.SetDateToDisplay ignored invalid date")
		return
	}
	h.setDateParts(componentName, year, month, day)
}

func (h *simulationHost) setDateParts(componentName string, year, month, day int) {
	h.setProperty(componentName, "Year", runtime.NumVal(float64(year)))
	h.setProperty(componentName, "Month", runtime.NumVal(float64(month)))
	h.setProperty(componentName, "Day", runtime.NumVal(float64(day)))
	h.setProperty(componentName, "MonthInText", runtime.StrVal(monthName(month)))
	h.setProperty(componentName, "Instant", runtime.StrVal(dateInstantString(year, month, day)))
}

func (h *simulationHost) setTimeProperty(componentName, property string, value runtime.Value) {
	hour, minute, ok := h.currentTimeParts(componentName)
	if !ok {
		hour, minute = 0, 0
	}
	next, ok := intValue(value)
	if !ok {
		h.logs = append(h.logs, "TimePicker."+property+" ignored non-numeric value: "+value.AsStr())
		return
	}
	switch property {
	case "Hour":
		hour = next
	case "Minute":
		minute = next
	}
	if !validTimeParts(hour, minute) {
		h.logs = append(h.logs, "TimePicker."+property+" ignored invalid time")
		return
	}
	h.setTimeParts(componentName, hour, minute)
}

func (h *simulationHost) setTimeToDisplay(componentName string, hourValue, minuteValue runtime.Value) {
	hour, okHour := intValue(hourValue)
	minute, okMinute := intValue(minuteValue)
	if !okHour || !okMinute || !validTimeParts(hour, minute) {
		h.logs = append(h.logs, "TimePicker.SetTimeToDisplay ignored invalid time")
		return
	}
	h.setTimeParts(componentName, hour, minute)
}

func (h *simulationHost) setTimeParts(componentName string, hour, minute int) {
	h.setProperty(componentName, "Hour", runtime.NumVal(float64(hour)))
	h.setProperty(componentName, "Minute", runtime.NumVal(float64(minute)))
	h.setProperty(componentName, "Instant", runtime.StrVal(timeInstantString(hour, minute)))
}

func intValue(value runtime.Value) (int, bool) {
	n, ok := runtime.CoerceNum(value)
	if !ok || math.IsNaN(n) || math.IsInf(n, 0) {
		return 0, false
	}
	return int(math.Trunc(n)), true
}

func validDateParts(year, month, day int) bool {
	if year < 1 || year > 9999 || month < 1 || month > 12 || day < 1 || day > 31 {
		return false
	}
	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	return t.Year() == year && int(t.Month()) == month && t.Day() == day
}

func validTimeParts(hour, minute int) bool {
	return hour >= 0 && hour <= 23 && minute >= 0 && minute <= 59
}

func monthName(month int) string {
	if month < 1 || month > 12 {
		return ""
	}
	return time.Month(month).String()
}

func dateInstantString(year, month, day int) string {
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)
}

func timeInstantString(hour, minute int) string {
	return time.Date(1970, 1, 1, hour, minute, 0, 0, time.UTC).Format(time.RFC3339)
}

func datePartsFromInstant(value runtime.Value) (int, int, int, bool) {
	switch value.Type() {
	case runtime.Number:
		t, ok := timeFromTimestamp(value.AsNum())
		if !ok {
			return 0, 0, 0, false
		}
		return t.Year(), int(t.Month()), t.Day(), true
	case runtime.String, runtime.Color:
		if t, ok := parseDateTimeString(value.AsStr()); ok {
			return t.Year(), int(t.Month()), t.Day(), true
		}
	case runtime.List, runtime.Matrix:
		values := valueList(value)
		if len(values) >= 3 {
			year, okYear := intValue(values[0])
			month, okMonth := intValue(values[1])
			day, okDay := intValue(values[2])
			if okYear && okMonth && okDay && validDateParts(year, month, day) {
				return year, month, day, true
			}
		}
	case runtime.Dict:
		if nested, ok := dictValueAny(value, "Instant", "instant", "Timestamp", "timestamp", "Millis", "millis"); ok {
			if year, month, day, ok := datePartsFromInstant(nested); ok {
				return year, month, day, true
			}
		}
		year, okYear := dictIntAny(value, "Year", "year")
		month, okMonth := dictIntAny(value, "Month", "month")
		day, okDay := dictIntAny(value, "Day", "day")
		if okYear && okMonth && okDay && validDateParts(year, month, day) {
			return year, month, day, true
		}
	}
	return 0, 0, 0, false
}

func timePartsFromInstant(value runtime.Value) (int, int, bool) {
	switch value.Type() {
	case runtime.Number:
		t, ok := timeFromTimestamp(value.AsNum())
		if !ok {
			return 0, 0, false
		}
		return t.Hour(), t.Minute(), true
	case runtime.String, runtime.Color:
		if t, ok := parseDateTimeString(value.AsStr()); ok {
			return t.Hour(), t.Minute(), true
		}
	case runtime.List, runtime.Matrix:
		values := valueList(value)
		if len(values) >= 2 {
			hour, okHour := intValue(values[0])
			minute, okMinute := intValue(values[1])
			if okHour && okMinute && validTimeParts(hour, minute) {
				return hour, minute, true
			}
		}
	case runtime.Dict:
		if nested, ok := dictValueAny(value, "Instant", "instant", "Timestamp", "timestamp", "Millis", "millis"); ok {
			if hour, minute, ok := timePartsFromInstant(nested); ok {
				return hour, minute, true
			}
		}
		hour, okHour := dictIntAny(value, "Hour", "hour")
		minute, okMinute := dictIntAny(value, "Minute", "minute")
		if !okMinute {
			minute, okMinute = dictIntAny(value, "Min", "min")
		}
		if okHour && okMinute && validTimeParts(hour, minute) {
			return hour, minute, true
		}
	}
	return 0, 0, false
}

func timeFromTimestamp(value float64) (time.Time, bool) {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return time.Time{}, false
	}
	absValue := math.Abs(value)
	if absValue > 1e11 {
		seconds := int64(value / 1000)
		millis := int64(math.Mod(value, 1000))
		return time.Unix(seconds, millis*int64(time.Millisecond)).UTC(), true
	}
	return time.Unix(int64(value), 0).UTC(), true
}

func parseDateTimeString(raw string) (time.Time, bool) {
	text := strings.TrimSpace(raw)
	if text == "" {
		return time.Time{}, false
	}
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04",
		"2006/01/02",
		"2006/01/02 15:04:05",
		"2006/01/02 15:04",
		"15:04:05",
		"15:04",
	}
	for _, layout := range layouts {
		if parsed, err := time.ParseInLocation(layout, text, time.UTC); err == nil {
			if layout == "15:04:05" || layout == "15:04" {
				return time.Date(1970, 1, 1, parsed.Hour(), parsed.Minute(), parsed.Second(), 0, time.UTC), true
			}
			return parsed.UTC(), true
		}
	}
	if n, err := strconv.ParseFloat(text, 64); err == nil {
		return timeFromTimestamp(n)
	}
	return time.Time{}, false
}

func dictValueAny(value runtime.Value, keys ...string) (runtime.Value, bool) {
	if value.Type() != runtime.Dict {
		return runtime.NullVal(), false
	}
	dict := value.AsDict()
	for _, key := range keys {
		if val, ok := dict.Get(key); ok {
			return val, true
		}
	}
	return runtime.NullVal(), false
}

func dictIntAny(value runtime.Value, keys ...string) (int, bool) {
	if val, ok := dictValueAny(value, keys...); ok {
		return intValue(val)
	}
	return 0, false
}

func (h *simulationHost) CallMethod(componentName, componentType, method string, args []runtime.Value) runtime.Value {
	componentType = h.componentType(componentName, componentType)
	switch componentType {
	case "TextBox", "PasswordTextBox":
		return h.callTextBoxMethod(componentName, method, args)
	case "Notifier":
		return h.callNotifierMethod(componentName, method, args)
	case "ListPicker":
		if method == "Open" {
			if h.runEvent != nil {
				h.runEvent(componentName, componentType, "BeforePicking", nil)
			}
			h.effects = append(h.effects, componentAction(componentName, "open"))
			return runtime.VoidVal()
		}
	case "Spinner":
		if method == "DisplayDropdown" {
			h.effects = append(h.effects, componentAction(componentName, "open"))
			return runtime.VoidVal()
		}
	case "DatePicker":
		switch method {
		case "LaunchPicker":
			if h.runEvent != nil {
				h.runEvent(componentName, componentType, "Click", nil)
			}
			h.effects = append(h.effects, componentAction(componentName, "open"))
			return runtime.VoidVal()
		case "SetDateToDisplay":
			if len(args) >= 3 {
				h.setDateToDisplay(componentName, args[0], args[1], args[2])
			}
			return runtime.VoidVal()
		case "SetDateToDisplayFromInstant":
			if len(args) >= 1 {
				if year, month, day, ok := datePartsFromInstant(args[0]); ok {
					h.setDateParts(componentName, year, month, day)
				} else {
					h.logs = append(h.logs, "DatePicker.SetDateToDisplayFromInstant ignored invalid instant: "+args[0].AsStr())
				}
			}
			return runtime.VoidVal()
		}
	case "TimePicker":
		switch method {
		case "LaunchPicker":
			if h.runEvent != nil {
				h.runEvent(componentName, componentType, "Click", nil)
			}
			h.effects = append(h.effects, componentAction(componentName, "open"))
			return runtime.VoidVal()
		case "SetTimeToDisplay":
			if len(args) >= 2 {
				h.setTimeToDisplay(componentName, args[0], args[1])
			}
			return runtime.VoidVal()
		case "SetTimeToDisplayFromInstant":
			if len(args) >= 1 {
				if hour, minute, ok := timePartsFromInstant(args[0]); ok {
					h.setTimeParts(componentName, hour, minute)
				} else {
					h.logs = append(h.logs, "TimePicker.SetTimeToDisplayFromInstant ignored invalid instant: "+args[0].AsStr())
				}
			}
			return runtime.VoidVal()
		}
	case "ListView":
		return h.callListViewMethod(componentName, componentType, method, args)
	case "TinyDB":
		return h.callTinyDBMethod(componentName, method, args)
	case "Screen", "Form":
		return h.callScreenMethod(componentName, method, args)
	case "EmailPicker":
		return h.callTextBoxMethod(componentName, method, args)
	case "ImagePicker", "FilePicker", "ContactPicker", "PhoneNumberPicker":
		return h.callPickerMethod(componentName, componentType, method, args)
	case "VideoPlayer":
		return h.callVideoPlayerMethod(componentName, method, args)
	case "WebViewer":
		return h.callWebViewerMethod(componentName, method, args)
	case "LinearProgress":
		if method == "IncrementProgressBy" {
			if len(args) >= 1 {
				cur := valueAsNumber(h.GetProperty(componentName, componentType, "Progress"), 0)
				min := valueAsNumber(h.GetProperty(componentName, componentType, "Minimum"), 0)
				max := valueAsNumber(h.GetProperty(componentName, componentType, "Maximum"), 100)
				next := clampFloat(cur+args[0].AsNum(), min, max)
				prev := cur
				h.setProperty(componentName, "Progress", runtime.NumVal(next))
				if next != prev && h.runEvent != nil {
					h.runEvent(componentName, componentType, "ProgressChanged", []runtime.Value{runtime.NumVal(next)})
				}
			}
			return runtime.VoidVal()
		}
	case "Canvas":
		return h.callCanvasMethod(componentName, method, args)
	case "Ball", "ImageSprite":
		return h.callSpriteMethod(componentName, componentType, method, args)
	case "Chart":
		return h.callChartMethod(componentName, method, args)
	case "ChartData2D":
		return h.callChartData2DMethod(componentName, method, args)
	case "Map":
		return h.callMapMethod(componentName, method, args)
	case "Marker", "Circle", "LineString", "Polygon", "Rectangle":
		return h.callMapFeatureMethod(componentName, componentType, method, args)
	case "FeatureCollection":
		if method == "LoadFromURL" || method == "FeatureFromDescription" {
			h.Unsupported("method", componentName+"."+method+" is not supported in the web simulator")
			return runtime.VoidVal()
		}
	}
	h.Unsupported("method", componentName+"."+method)
	return runtime.VoidVal()
}

func componentAction(componentName, action string) map[string]any {
	return map[string]any{
		"type":      "component-action",
		"component": componentName,
		"action":    action,
	}
}

func componentActionWith(componentName, action string, fields map[string]any) map[string]any {
	effect := componentAction(componentName, action)
	for key, value := range fields {
		effect[key] = value
	}
	return effect
}

func (h *simulationHost) callTextBoxMethod(componentName, method string, args []runtime.Value) runtime.Value {
	switch method {
	case "HideKeyboard":
		return runtime.VoidVal()
	case "RequestFocus":
		h.effects = append(h.effects, componentAction(componentName, "focus"))
	case "MoveCursorToStart":
		h.effects = append(h.effects, componentAction(componentName, "cursor-start"))
	case "MoveCursorToEnd":
		h.effects = append(h.effects, componentAction(componentName, "cursor-end"))
	case "MoveCursorTo":
		fields := map[string]any{}
		if len(args) >= 1 {
			fields["position"] = args[0].AsNum()
		}
		h.effects = append(h.effects, componentActionWith(componentName, "cursor-position", fields))
	default:
		h.Unsupported("method", componentName+"."+method)
	}
	return runtime.VoidVal()
}

func (h *simulationHost) callNotifierMethod(componentName, method string, args []runtime.Value) runtime.Value {
	switch method {
	case "ShowAlert":
		text := argString(args, 0, "")
		length := int(valueAsNumber(h.GetProperty(componentName, "Notifier", "NotifierLength"), 1))
		duration := 3500
		if length == 0 {
			duration = 2000
		}
		h.effects = append(h.effects, map[string]any{
			"type":            "notice",
			"component":       componentName,
			"text":            text,
			"duration":        duration,
			"backgroundColor": valueAsString(h.GetProperty(componentName, "Notifier", "BackgroundColor"), "&HFF444444"),
			"textColor":       valueAsString(h.GetProperty(componentName, "Notifier", "TextColor"), "&HFFFFFFFF"),
		})
	case "ShowMessageDialog":
		h.effects = append(h.effects, map[string]any{
			"type":       "dialog",
			"component":  componentName,
			"dialogType": "message",
			"message":    argString(args, 0, ""),
			"title":      argString(args, 1, ""),
			"buttonText": argString(args, 2, "OK"),
		})
	case "ShowChooseDialog":
		h.effects = append(h.effects, map[string]any{
			"type":        "dialog",
			"component":   componentName,
			"dialogType":  "choose",
			"message":     argString(args, 0, ""),
			"title":       argString(args, 1, ""),
			"button1Text": argString(args, 2, "OK"),
			"button2Text": argString(args, 3, "Cancel"),
			"cancelable":  argBool(args, 4, true),
		})
	case "ShowTextDialog", "ShowPasswordDialog":
		dialogType := "text"
		if method == "ShowPasswordDialog" {
			dialogType = "password"
		}
		h.effects = append(h.effects, map[string]any{
			"type":       "dialog",
			"component":  componentName,
			"dialogType": dialogType,
			"message":    argString(args, 0, ""),
			"title":      argString(args, 1, ""),
			"cancelable": argBool(args, 2, true),
		})
	case "ShowProgressDialog":
		h.effects = append(h.effects, map[string]any{
			"type":       "dialog",
			"component":  componentName,
			"dialogType": "progress",
			"message":    argString(args, 0, ""),
			"title":      argString(args, 1, ""),
		})
	case "DismissProgressDialog":
		h.effects = append(h.effects, map[string]any{
			"type":       "dialog-dismiss",
			"component":  componentName,
			"dialogType": "progress",
		})
	case "LogError", "LogWarning", "LogInfo":
		level := strings.TrimPrefix(method, "Log")
		h.logs = append(h.logs, "["+level+"] "+argString(args, 0, ""))
	default:
		h.Unsupported("method", componentName+"."+method)
	}
	return runtime.VoidVal()
}

func (h *simulationHost) callScreenMethod(componentName, method string, args []runtime.Value) runtime.Value {
	switch method {
	case "AskForPermission", "HideKeyboard":
		return runtime.VoidVal()
	case "OpenScreen", "OpenAnotherScreen", "OpenScreenWithStartValue", "OpenAnotherScreenWithStartValue",
		"CloseScreen", "CloseScreenWithValue", "CloseScreenWithPlainText", "CloseApplication", "CloseApp",
		"FinishActivity", "FinishActivityWithResult", "FinishApplication", "SwitchForm", "SwitchFormWithStartValue":
		h.Unsupported("navigation", componentName+"."+method+" is not supported in the web simulator")
		return runtime.VoidVal()
	default:
		h.Unsupported("method", componentName+"."+method)
		return runtime.VoidVal()
	}
}

func (h *simulationHost) callListViewMethod(componentName, componentType, method string, args []runtime.Value) runtime.Value {
	elements := valueList(h.GetProperty(componentName, componentType, "Elements"))
	switch method {
	case "AddItem":
		if len(args) >= 1 {
			elements = append(elements, h.listViewElement(componentName, args, 0))
			h.SetProperty(componentName, componentType, "Elements", runtime.ListVal(elements))
		}
	case "AddItemAtIndex":
		if len(args) >= 2 {
			index := int(args[0].AsNum()) - 1
			if index < 0 {
				index = 0
			}
			if index > len(elements) {
				index = len(elements)
			}
			elements = append(elements[:index], append([]runtime.Value{h.listViewElement(componentName, args, 1)}, elements[index:]...)...)
			h.SetProperty(componentName, componentType, "Elements", runtime.ListVal(elements))
		}
	case "AddItems":
		if len(args) >= 1 {
			elements = append(elements, valueList(args[0])...)
			h.SetProperty(componentName, componentType, "Elements", runtime.ListVal(elements))
		}
	case "AddItemsAtIndex":
		if len(args) >= 2 {
			index := int(args[0].AsNum()) - 1
			if index < 0 {
				index = 0
			}
			if index > len(elements) {
				index = len(elements)
			}
			items := valueList(args[1])
			next := make([]runtime.Value, 0, len(elements)+len(items))
			next = append(next, elements[:index]...)
			next = append(next, items...)
			next = append(next, elements[index:]...)
			h.SetProperty(componentName, componentType, "Elements", runtime.ListVal(next))
		}
	case "CreateElement":
		return createListViewElement(argValue(args, 0), argValue(args, 1), argValue(args, 2), true)
	case "GetMainText":
		if len(args) >= 1 {
			return runtime.StrVal(listViewMainText(args[0]))
		}
		return runtime.StrVal("")
	case "GetDetailText":
		if len(args) >= 1 {
			return runtime.StrVal(listViewDetailText(args[0]))
		}
		return runtime.StrVal("")
	case "GetImageName":
		if len(args) >= 1 {
			return runtime.StrVal(listViewImageName(args[0]))
		}
		return runtime.StrVal("")
	case "RemoveItemAtIndex":
		if len(args) >= 1 {
			index := int(args[0].AsNum()) - 1
			if index >= 0 && index < len(elements) {
				elements = append(elements[:index], elements[index+1:]...)
				h.SetProperty(componentName, componentType, "Elements", runtime.ListVal(elements))
			}
		}
	case "Refresh":
		return runtime.VoidVal()
	default:
		h.Unsupported("method", componentName+"."+method)
	}
	return runtime.VoidVal()
}

func (h *simulationHost) listViewElement(componentName string, args []runtime.Value, start int) runtime.Value {
	main := argValue(args, start)
	detail := argValue(args, start+1)
	image := argValue(args, start+2)
	forceDict := false
	if detail.AsStr() != "" || image.AsStr() != "" {
		forceDict = true
	}
	if valueAsNumber(h.GetProperty(componentName, "ListView", "ListViewLayout"), 0) > 0 {
		forceDict = true
	}
	for _, element := range valueList(h.GetProperty(componentName, "ListView", "Elements")) {
		if element.Type() == runtime.Dict {
			forceDict = true
			break
		}
	}
	return createListViewElement(main, detail, image, forceDict)
}

func createListViewElement(main, detail, image runtime.Value, forceDict bool) runtime.Value {
	if !forceDict {
		return main
	}
	dict := runtime.NewOrderedDict()
	dict.Set("Text1", copyRuntimeValue(main))
	dict.Set("Text2", copyRuntimeValue(detail))
	dict.Set("Image", copyRuntimeValue(image))
	return runtime.DictVal(dict)
}

func listViewMainText(value runtime.Value) string {
	if value.Type() != runtime.Dict {
		return value.AsStr()
	}
	if text, ok := dictStringAny(value, "Text1", "MainText", "mainText", "text1", "main"); ok {
		return text
	}
	return ""
}

func listViewDetailText(value runtime.Value) string {
	if value.Type() != runtime.Dict {
		return ""
	}
	if text, ok := dictStringAny(value, "Text2", "DetailText", "detailText", "text2", "detail"); ok {
		return text
	}
	return ""
}

func listViewImageName(value runtime.Value) string {
	if value.Type() != runtime.Dict {
		return ""
	}
	if text, ok := dictStringAny(value, "Image", "ImageName", "imageName", "image"); ok {
		return text
	}
	return ""
}

func dictStringAny(value runtime.Value, keys ...string) (string, bool) {
	if val, ok := dictValueAny(value, keys...); ok {
		return val.AsStr(), true
	}
	return "", false
}

func (h *simulationHost) tinyDBStore(componentName string) map[string]runtime.Value {
	namespace := h.tinyDBNamespace(componentName)
	if _, ok := h.tinyDB[namespace]; !ok {
		h.tinyDB[namespace] = map[string]runtime.Value{}
	}
	return h.tinyDB[namespace]
}

func (h *simulationHost) tinyDBNamespace(componentName string) string {
	namespace := h.GetProperty(componentName, "TinyDB", "Namespace").AsStr()
	if namespace == "" {
		namespace = "TinyDB1"
	}
	return namespace
}

func (h *simulationHost) callTinyDBMethod(componentName, method string, args []runtime.Value) runtime.Value {
	store := h.tinyDBStore(componentName)
	switch method {
	case "StoreValue":
		if len(args) >= 2 {
			store[args[0].AsStr()] = copyRuntimeValue(args[1])
		}
		return runtime.VoidVal()
	case "GetValue":
		if len(args) >= 1 {
			if value, ok := store[args[0].AsStr()]; ok {
				return copyRuntimeValue(value)
			}
		}
		if len(args) >= 2 {
			return copyRuntimeValue(args[1])
		}
		return runtime.NullVal()
	case "ClearTag":
		if len(args) >= 1 {
			delete(store, args[0].AsStr())
		}
		return runtime.VoidVal()
	case "ClearAll":
		h.tinyDB[h.tinyDBNamespace(componentName)] = map[string]runtime.Value{}
		return runtime.VoidVal()
	case "GetTags":
		names := make([]string, 0, len(store))
		for tag := range store {
			names = append(names, tag)
		}
		sort.Strings(names)
		tags := make([]runtime.Value, 0, len(names))
		for _, tag := range names {
			tags = append(tags, runtime.StrVal(tag))
		}
		return runtime.ListVal(tags)
	case "GetEntries":
		names := make([]string, 0, len(store))
		for tag := range store {
			names = append(names, tag)
		}
		sort.Strings(names)
		entries := runtime.NewOrderedDict()
		for _, tag := range names {
			entries.Set(tag, copyRuntimeValue(store[tag]))
		}
		return runtime.DictVal(entries)
	default:
		h.Unsupported("method", "TinyDB."+method)
		return runtime.VoidVal()
	}
}

func (h *simulationHost) callPickerMethod(componentName, componentType, method string, args []runtime.Value) runtime.Value {
	switch method {
	case "Open":
		if h.runEvent != nil {
			h.runEvent(componentName, componentType, "BeforePicking", nil)
		}
		h.effects = append(h.effects, componentAction(componentName, "open"))
	case "ViewContact":
		h.Unsupported("method", componentName+".ViewContact is not supported in the web simulator")
	default:
		h.Unsupported("method", componentName+"."+method)
	}
	return runtime.VoidVal()
}

func (h *simulationHost) callVideoPlayerMethod(componentName, method string, args []runtime.Value) runtime.Value {
	switch method {
	case "Start":
		h.effects = append(h.effects, componentAction(componentName, "play"))
	case "Pause":
		h.effects = append(h.effects, componentAction(componentName, "pause"))
	case "Stop":
		h.effects = append(h.effects, componentActionWith(componentName, "seek", map[string]any{"ms": float64(0)}))
		h.effects = append(h.effects, componentAction(componentName, "pause"))
	case "SeekTo":
		if len(args) >= 1 {
			h.effects = append(h.effects, componentActionWith(componentName, "seek", map[string]any{"ms": args[0].AsNum()}))
		}
	case "GetDuration":
		dur := valueAsNumber(h.GetProperty(componentName, "VideoPlayer", "Duration"), 0)
		return runtime.NumVal(dur)
	default:
		h.Unsupported("method", componentName+"."+method)
	}
	return runtime.VoidVal()
}

func (h *simulationHost) callWebViewerMethod(componentName, method string, args []runtime.Value) runtime.Value {
	switch method {
	case "GoToUrl":
		if len(args) >= 1 {
			url := args[0].AsStr()
			h.setProperty(componentName, "CurrentUrl", runtime.StrVal(url))
			h.effects = append(h.effects, componentActionWith(componentName, "navigate", map[string]any{"url": url}))
		}
	case "GoHome":
		home := h.GetProperty(componentName, "WebViewer", "HomeUrl").AsStr()
		h.setProperty(componentName, "CurrentUrl", runtime.StrVal(home))
		h.effects = append(h.effects, componentActionWith(componentName, "navigate", map[string]any{"url": home}))
	case "GoBack":
		h.effects = append(h.effects, componentAction(componentName, "goback"))
	case "GoForward":
		h.effects = append(h.effects, componentAction(componentName, "goforward"))
	case "Reload":
		h.effects = append(h.effects, componentAction(componentName, "reload"))
	case "CanGoBack":
		return runtime.BoolVal(false)
	case "CanGoForward":
		return runtime.BoolVal(false)
	case "RunJavaScript":
		h.Unsupported("method", componentName+".RunJavaScript requires same-origin page")
	case "ClearCaches", "ClearCookies", "ClearLocations", "StopLoading":
		// No-ops in the browser simulator
	default:
		h.Unsupported("method", componentName+"."+method)
	}
	return runtime.VoidVal()
}

func (h *simulationHost) callCanvasMethod(componentName, method string, args []runtime.Value) runtime.Value {
	switch method {
	case "Clear":
		h.effects = append(h.effects, componentActionWith(componentName, "canvas-clear", nil))
	case "DrawLine":
		if len(args) >= 4 {
			h.effects = append(h.effects, componentActionWith(componentName, "canvas-draw", map[string]any{
				"op": "line", "x1": args[0].AsNum(), "y1": args[1].AsNum(), "x2": args[2].AsNum(), "y2": args[3].AsNum(),
				"color":     h.GetProperty(componentName, "Canvas", "PaintColor").AsStr(),
				"lineWidth": valueAsNumber(h.GetProperty(componentName, "Canvas", "LineWidth"), 2),
			}))
		}
	case "DrawCircle":
		if len(args) >= 4 {
			h.effects = append(h.effects, componentActionWith(componentName, "canvas-draw", map[string]any{
				"op": "circle", "cx": args[0].AsNum(), "cy": args[1].AsNum(), "r": args[2].AsNum(),
				"fill":  args[3].AsBool(),
				"color": h.GetProperty(componentName, "Canvas", "PaintColor").AsStr(),
			}))
		}
	case "DrawPoint":
		if len(args) >= 2 {
			h.effects = append(h.effects, componentActionWith(componentName, "canvas-draw", map[string]any{
				"op": "point", "x": args[0].AsNum(), "y": args[1].AsNum(),
				"color":     h.GetProperty(componentName, "Canvas", "PaintColor").AsStr(),
				"lineWidth": valueAsNumber(h.GetProperty(componentName, "Canvas", "LineWidth"), 2),
			}))
		}
	case "DrawText":
		if len(args) >= 3 {
			h.effects = append(h.effects, componentActionWith(componentName, "canvas-draw", map[string]any{
				"op": "text", "text": args[0].AsStr(), "x": args[1].AsNum(), "y": args[2].AsNum(),
				"color":    h.GetProperty(componentName, "Canvas", "PaintColor").AsStr(),
				"fontSize": valueAsNumber(h.GetProperty(componentName, "Canvas", "FontSize"), 14),
			}))
		}
	case "DrawArc":
		if len(args) >= 8 {
			h.effects = append(h.effects, componentActionWith(componentName, "canvas-draw", map[string]any{
				"op": "arc", "left": args[0].AsNum(), "top": args[1].AsNum(), "right": args[2].AsNum(), "bottom": args[3].AsNum(),
				"startAngle": args[4].AsNum(), "sweepAngle": args[5].AsNum(),
				"useCenter": args[6].AsBool(), "fill": args[7].AsBool(),
				"color": h.GetProperty(componentName, "Canvas", "PaintColor").AsStr(),
			}))
		}
	case "DrawShape":
		if len(args) >= 2 {
			h.effects = append(h.effects, componentActionWith(componentName, "canvas-draw", map[string]any{
				"op": "shape", "points": runtimeValueToJS(args[0]), "fill": args[1].AsBool(),
				"color":     h.GetProperty(componentName, "Canvas", "PaintColor").AsStr(),
				"lineWidth": valueAsNumber(h.GetProperty(componentName, "Canvas", "LineWidth"), 2),
			}))
		}
	case "DrawTextAtAngle":
		if len(args) >= 4 {
			h.effects = append(h.effects, componentActionWith(componentName, "canvas-draw", map[string]any{
				"op": "textAngle", "text": args[0].AsStr(), "x": args[1].AsNum(), "y": args[2].AsNum(), "angle": args[3].AsNum(),
				"color":    h.GetProperty(componentName, "Canvas", "PaintColor").AsStr(),
				"fontSize": valueAsNumber(h.GetProperty(componentName, "Canvas", "FontSize"), 14),
			}))
		}
	case "SetBackgroundPixelColor":
		h.Unsupported("method", componentName+".SetBackgroundPixelColor is not supported")
	case "GetPixelColor", "GetBackgroundPixelColor":
		return runtime.NumVal(0)
	case "Save", "SaveAs":
		h.Unsupported("method", componentName+"."+method+" cannot write device storage in the web simulator")
		return runtime.StrVal("")
	default:
		h.Unsupported("method", componentName+"."+method)
	}
	return runtime.VoidVal()
}

func (h *simulationHost) callSpriteMethod(componentName, componentType, method string, args []runtime.Value) runtime.Value {
	switch method {
	case "MoveTo":
		if len(args) >= 2 {
			h.setProperty(componentName, "X", runtime.NumVal(args[0].AsNum()))
			h.setProperty(componentName, "Y", runtime.NumVal(args[1].AsNum()))
		}
	case "MoveToPoint":
		if len(args) >= 1 {
			coords := valueList(args[0])
			if len(coords) >= 2 {
				h.setProperty(componentName, "X", runtime.NumVal(coords[0].AsNum()))
				h.setProperty(componentName, "Y", runtime.NumVal(coords[1].AsNum()))
			}
		}
	case "PointInDirection":
		if len(args) >= 2 {
			dx := args[0].AsNum() - valueAsNumber(h.GetProperty(componentName, componentType, "X"), 0)
			dy := args[1].AsNum() - valueAsNumber(h.GetProperty(componentName, componentType, "Y"), 0)
			heading := math.Atan2(-dy, dx) * 180 / math.Pi
			if heading < 0 {
				heading += 360
			}
			h.setProperty(componentName, "Heading", runtime.NumVal(heading))
		}
	case "Bounce", "MoveIntoBounds", "PointTowards":
		h.Unsupported("method", componentName+"."+method+" requires a running animation loop")
	case "CollidingWith":
		return runtime.BoolVal(false)
	default:
		h.Unsupported("method", componentName+"."+method)
	}
	return runtime.VoidVal()
}

func (h *simulationHost) callChartMethod(componentName, method string, args []runtime.Value) runtime.Value {
	switch method {
	case "SetDomain":
		if len(args) >= 2 {
			h.setProperty(componentName, "XMin", runtime.NumVal(args[0].AsNum()))
			h.setProperty(componentName, "XMax", runtime.NumVal(args[1].AsNum()))
		}
	case "SetRange":
		if len(args) >= 2 {
			h.setProperty(componentName, "YMin", runtime.NumVal(args[0].AsNum()))
			h.setProperty(componentName, "YMax", runtime.NumVal(args[1].AsNum()))
		}
	case "ResetAxes":
		h.setProperty(componentName, "XMin", runtime.NullVal())
		h.setProperty(componentName, "XMax", runtime.NullVal())
		h.setProperty(componentName, "YMin", runtime.NullVal())
		h.setProperty(componentName, "YMax", runtime.NullVal())
	case "ExtendDomainToInclude", "ExtendRangeToInclude":
		// Client-side concern; just log
	default:
		h.Unsupported("method", componentName+"."+method)
	}
	return runtime.VoidVal()
}

func (h *simulationHost) callChartData2DMethod(componentName, method string, args []runtime.Value) runtime.Value {
	elements := valueList(h.GetProperty(componentName, "ChartData2D", "Elements"))
	switch method {
	case "AddEntry":
		if len(args) >= 2 {
			pair := runtime.ListVal([]runtime.Value{args[0], args[1]})
			elements = append(elements, pair)
			h.setProperty(componentName, "Elements", runtime.ListVal(elements))
		}
	case "RemoveEntry":
		if len(args) >= 2 {
			xS := args[0].AsStr()
			yS := args[1].AsStr()
			next := make([]runtime.Value, 0, len(elements))
			for _, el := range elements {
				pair := valueList(el)
				if len(pair) >= 2 && (pair[0].AsStr() == xS && pair[1].AsStr() == yS) {
					continue
				}
				next = append(next, el)
			}
			h.setProperty(componentName, "Elements", runtime.ListVal(next))
		}
	case "Clear":
		h.setProperty(componentName, "Elements", runtime.ListVal(nil))
	case "ImportFromList":
		if len(args) >= 1 {
			h.setProperty(componentName, "Elements", args[0])
		}
	case "ImportFromTinyDB":
		if len(args) >= 2 {
			namespace := args[0].AsStr()
			tag := args[1].AsStr()
			if store, ok := h.tinyDB[namespace]; ok {
				if val, ok := store[tag]; ok {
					h.setProperty(componentName, "Elements", val)
				}
			}
		}
	case "GetAllEntries":
		return h.GetProperty(componentName, "ChartData2D", "Elements")
	case "DoesEntryExist":
		if len(args) >= 2 {
			xS := args[0].AsStr()
			yS := args[1].AsStr()
			for _, el := range elements {
				pair := valueList(el)
				if len(pair) >= 2 && pair[0].AsStr() == xS && pair[1].AsStr() == yS {
					return runtime.BoolVal(true)
				}
			}
		}
		return runtime.BoolVal(false)
	case "GetEntriesWithXValue":
		if len(args) >= 1 {
			xS := args[0].AsStr()
			result := []runtime.Value{}
			for _, el := range elements {
				pair := valueList(el)
				if len(pair) >= 2 && pair[0].AsStr() == xS {
					result = append(result, el)
				}
			}
			return runtime.ListVal(result)
		}
		return runtime.ListVal(nil)
	case "GetEntriesWithYValue":
		if len(args) >= 1 {
			yS := args[0].AsStr()
			result := []runtime.Value{}
			for _, el := range elements {
				pair := valueList(el)
				if len(pair) >= 2 && pair[1].AsStr() == yS {
					result = append(result, el)
				}
			}
			return runtime.ListVal(result)
		}
		return runtime.ListVal(nil)
	case "HighlightDataPoints":
		h.setProperty(componentName, "HighlightColor", argValue(args, 1))
	case "ChangeDataSource", "RemoveDataSource", "ImportFromCloudDB", "ImportFromDataFile", "ImportFromSpreadsheet", "ImportFromWeb":
		h.Unsupported("method", componentName+"."+method+" — live data sources are not available in the web simulator")
	default:
		h.Unsupported("method", componentName+"."+method)
	}
	return runtime.VoidVal()
}

func (h *simulationHost) callMapMethod(componentName, method string, args []runtime.Value) runtime.Value {
	switch method {
	case "PanTo":
		if len(args) >= 2 {
			h.setProperty(componentName, "Latitude", runtime.NumVal(args[0].AsNum()))
			h.setProperty(componentName, "Longitude", runtime.NumVal(args[1].AsNum()))
			if len(args) >= 3 {
				h.setProperty(componentName, "ZoomLevel", runtime.NumVal(args[2].AsNum()))
			}
		}
	case "CreateMarker":
		if len(args) >= 2 {
			h.effects = append(h.effects, componentActionWith(componentName, "map-create-marker", map[string]any{
				"latitude": args[0].AsNum(), "longitude": args[1].AsNum(),
			}))
		}
		return runtime.NullVal()
	case "LoadFromURL":
		h.Unsupported("method", componentName+".LoadFromURL is subject to CORS restrictions in the web simulator")
	case "FeatureFromDescription":
		return runtime.NullVal()
	case "Save":
		h.Unsupported("method", componentName+".Save cannot write device storage in the web simulator")
	default:
		h.Unsupported("method", componentName+"."+method)
	}
	return runtime.VoidVal()
}

func (h *simulationHost) callMapFeatureMethod(componentName, componentType, method string, args []runtime.Value) runtime.Value {
	switch method {
	case "SetLocation":
		if len(args) >= 2 {
			h.setProperty(componentName, "Latitude", runtime.NumVal(args[0].AsNum()))
			h.setProperty(componentName, "Longitude", runtime.NumVal(args[1].AsNum()))
		}
	case "SetCenter":
		if len(args) >= 2 {
			lat := args[0].AsNum()
			lng := args[1].AsNum()
			north := valueAsNumber(h.GetProperty(componentName, componentType, "NorthLatitude"), 0)
			south := valueAsNumber(h.GetProperty(componentName, componentType, "SouthLatitude"), 0)
			east := valueAsNumber(h.GetProperty(componentName, componentType, "EastLongitude"), 0)
			west := valueAsNumber(h.GetProperty(componentName, componentType, "WestLongitude"), 0)
			halfLat := (north - south) / 2
			halfLng := (east - west) / 2
			h.setProperty(componentName, "NorthLatitude", runtime.NumVal(lat+halfLat))
			h.setProperty(componentName, "SouthLatitude", runtime.NumVal(lat-halfLat))
			h.setProperty(componentName, "EastLongitude", runtime.NumVal(lng+halfLng))
			h.setProperty(componentName, "WestLongitude", runtime.NumVal(lng-halfLng))
		}
	case "ShowInfobox":
		h.effects = append(h.effects, componentAction(componentName, "show-infobox"))
	case "HideInfobox":
		h.effects = append(h.effects, componentAction(componentName, "hide-infobox"))
	case "DistanceToPoint", "DistanceToFeature", "BearingToPoint", "BearingToFeature":
		h.Unsupported("method", componentName+"."+method+" — geo calculations not yet implemented in the simulator")
		return runtime.NumVal(0)
	case "Centroid":
		return runtime.ListVal(nil)
	case "Bounds":
		n := valueAsNumber(h.GetProperty(componentName, componentType, "NorthLatitude"), 0)
		s := valueAsNumber(h.GetProperty(componentName, componentType, "SouthLatitude"), 0)
		e := valueAsNumber(h.GetProperty(componentName, componentType, "EastLongitude"), 0)
		w := valueAsNumber(h.GetProperty(componentName, componentType, "WestLongitude"), 0)
		return runtime.ListVal([]runtime.Value{runtime.NumVal(n), runtime.NumVal(w), runtime.NumVal(s), runtime.NumVal(e)})
	case "Center":
		n := valueAsNumber(h.GetProperty(componentName, componentType, "NorthLatitude"), 0)
		s := valueAsNumber(h.GetProperty(componentName, componentType, "SouthLatitude"), 0)
		e := valueAsNumber(h.GetProperty(componentName, componentType, "EastLongitude"), 0)
		w := valueAsNumber(h.GetProperty(componentName, componentType, "WestLongitude"), 0)
		return runtime.ListVal([]runtime.Value{runtime.NumVal((n + s) / 2), runtime.NumVal((e + w) / 2)})
	default:
		h.Unsupported("method", componentName+"."+method)
	}
	return runtime.VoidVal()
}

func (h *simulationHost) Unsupported(kind, detail string) {
	h.unsupported = append(h.unsupported, map[string]any{
		"kind":   kind,
		"detail": detail,
	})
}

func (h *simulationHost) resetTransient() {
	h.statePatch = map[string]map[string]any{}
	h.effects = nil
	h.logs = nil
	h.unsupported = nil
}

func (h *simulationHost) payloadFields() map[string]any {
	unsupported := make([]any, len(h.unsupported))
	for i, entry := range h.unsupported {
		unsupported[i] = entry
	}
	effects := make([]any, len(h.effects))
	for i, effect := range h.effects {
		effects[i] = effect
	}
	logs := make([]any, len(h.logs))
	for i, line := range h.logs {
		logs[i] = line
	}
	return map[string]any{
		"statePatch":   jsStatePatch(h.statePatch),
		"effects":      effects,
		"logs":         logs,
		"unsupported":  unsupported,
		"tinyDBStores": jsTinyDBStoresSnapshot(h.tinyDB),
	}
}

func runtimeValueToJS(value runtime.Value) any {
	switch value.Type() {
	case runtime.Null:
		return nil
	case runtime.Bool:
		return value.AsBool()
	case runtime.Number:
		return value.AsNum()
	case runtime.String, runtime.Color:
		return value.AsStr()
	case runtime.List, runtime.Matrix:
		values := value.AsList()
		out := make([]any, len(*values))
		for i, item := range *values {
			out[i] = runtimeValueToJS(item)
		}
		return out
	case runtime.Dict:
		dict := value.AsDict()
		out := map[string]any{}
		for _, key := range dict.Keys() {
			keyText := key.AsStr()
			if item, ok := dict.Get(keyText); ok {
				out[keyText] = runtimeValueToJS(item)
			}
		}
		return out
	default:
		return value.String()
	}
}

func jsToRuntimeValue(value js.Value) runtime.Value {
	if value.IsUndefined() || value.IsNull() {
		return runtime.NullVal()
	}
	switch value.Type() {
	case js.TypeBoolean:
		return runtime.BoolVal(value.Bool())
	case js.TypeNumber:
		return runtime.NumVal(value.Float())
	case js.TypeString:
		return runtime.StrVal(value.String())
	default:
		if js.Global().Get("Array").Call("isArray", value).Bool() {
			values := make([]runtime.Value, value.Length())
			for i := 0; i < value.Length(); i++ {
				values[i] = jsToRuntimeValue(value.Index(i))
			}
			return runtime.ListVal(values)
		}
		keys := js.Global().Get("Object").Call("keys", value)
		if keys.Length() > 0 {
			dict := runtime.NewOrderedDict()
			for i := 0; i < keys.Length(); i++ {
				key := keys.Index(i).String()
				dict.Set(key, jsToRuntimeValue(value.Get(key)))
			}
			return runtime.DictVal(dict)
		}
		return runtime.StrVal(value.String())
	}
}

func copyRuntimeValue(value runtime.Value) runtime.Value {
	switch value.Type() {
	case runtime.List:
		values := value.AsList()
		out := make([]runtime.Value, len(*values))
		for i, item := range *values {
			out[i] = copyRuntimeValue(item)
		}
		return runtime.ListVal(out)
	case runtime.Matrix:
		values := value.AsList()
		out := make([]runtime.Value, len(*values))
		for i, item := range *values {
			out[i] = copyRuntimeValue(item)
		}
		return runtime.MatrixVal(out)
	case runtime.Dict:
		dict := runtime.NewOrderedDict()
		for _, key := range value.AsDict().Keys() {
			keyText := key.AsStr()
			if item, ok := value.AsDict().Get(keyText); ok {
				dict.Set(keyText, copyRuntimeValue(item))
			}
		}
		return runtime.DictVal(dict)
	default:
		return value
	}
}

func valueList(value runtime.Value) []runtime.Value {
	if value.Type() == runtime.List || value.Type() == runtime.Matrix {
		values := value.AsList()
		out := make([]runtime.Value, len(*values))
		copy(out, *values)
		return out
	}
	if value.Type() == runtime.Null {
		return []runtime.Value{}
	}
	return []runtime.Value{value}
}

func elementsFromString(text string) []runtime.Value {
	parts := strings.Split(text, ",")
	values := make([]runtime.Value, 0, len(parts))
	for _, part := range parts {
		clean := strings.TrimSpace(part)
		if clean != "" {
			values = append(values, runtime.StrVal(clean))
		}
	}
	return values
}

func jsInitialState(value js.Value) map[string]map[string]runtime.Value {
	state := map[string]map[string]runtime.Value{}
	if value.IsUndefined() || value.IsNull() {
		return state
	}
	keys := js.Global().Get("Object").Call("keys", value)
	for i := 0; i < keys.Length(); i++ {
		component := keys.Index(i).String()
		propsValue := value.Get(component)
		props := map[string]runtime.Value{}
		propKeys := js.Global().Get("Object").Call("keys", propsValue)
		for j := 0; j < propKeys.Length(); j++ {
			prop := propKeys.Index(j).String()
			props[prop] = jsToRuntimeValue(propsValue.Get(prop))
		}
		state[component] = props
	}
	return state
}

func jsTinyDBStores(value js.Value) map[string]map[string]runtime.Value {
	stores := map[string]map[string]runtime.Value{}
	if value.IsUndefined() || value.IsNull() {
		return stores
	}
	namespaces := js.Global().Get("Object").Call("keys", value)
	for i := 0; i < namespaces.Length(); i++ {
		namespace := namespaces.Index(i).String()
		storeValue := value.Get(namespace)
		if storeValue.IsUndefined() || storeValue.IsNull() {
			continue
		}
		store := map[string]runtime.Value{}
		tags := js.Global().Get("Object").Call("keys", storeValue)
		for j := 0; j < tags.Length(); j++ {
			tag := tags.Index(j).String()
			store[tag] = jsToRuntimeValue(storeValue.Get(tag))
		}
		stores[namespace] = store
	}
	return stores
}

func jsStatePatch(patch map[string]map[string]any) map[string]any {
	out := map[string]any{}
	for component, props := range patch {
		propOut := map[string]any{}
		for prop, value := range props {
			propOut[prop] = value
		}
		out[component] = propOut
	}
	return out
}

func jsTinyDBStoresSnapshot(stores map[string]map[string]runtime.Value) map[string]any {
	out := map[string]any{}
	for namespace, store := range stores {
		storeOut := map[string]any{}
		for tag, value := range store {
			storeOut[tag] = runtimeValueToJS(value)
		}
		out[namespace] = storeOut
	}
	return out
}

func simulationResult(ok bool, sessionID int, phase, raw string, diagnostics []wasmDiagnostic, fields map[string]any) js.Value {
	if fields == nil {
		fields = map[string]any{}
	}
	fields["sessionId"] = sessionID
	if _, ok := fields["statePatch"]; !ok {
		fields["statePatch"] = map[string]any{}
	}
	if _, ok := fields["effects"]; !ok {
		fields["effects"] = []any{}
	}
	if _, ok := fields["logs"]; !ok {
		fields["logs"] = []any{}
	}
	if _, ok := fields["unsupported"]; !ok {
		fields["unsupported"] = []any{}
	}
	if _, ok := fields["tinyDBStores"]; !ok {
		fields["tinyDBStores"] = map[string]any{}
	}
	return diagnosticResult(ok, phase, "simulation", raw, diagnostics, fields)
}

func parseSimulationSource(sourceCode string, componentContextMap map[string][]string, reverseComponentMap map[string]string) ([]ast.Expr, []wasmDiagnostic, string) {
	codeContext := &context.CodeContext{SourceCode: &sourceCode, FileName: "simulation"}
	tokens := lex.NewLexer(codeContext).Lex()
	langParser := newWasmLangParser(true, tokens)
	langParser.SetComponentDefinitions(componentContextMap, reverseComponentMap)
	var exprs []ast.Expr
	var raw string
	var diagnostics []wasmDiagnostic
	func() {
		defer func() {
			if r := recover(); r != nil {
				raw, diagnostics = diagnosticsForRecover(r, "compile", "simulation")
			}
		}()
		exprs = langParser.ParseAll()
	}()
	return exprs, diagnostics, raw
}

func createSimulationSession(this js.Value, p []js.Value) any {
	if len(p) < 3 {
		return simulationResult(false, 0, "compile", "createSimulationSession(source, componentDefs, initialState) not provided", nil, nil)
	}
	sourceCode := p[0].String()
	componentContextMap, reverseComponentMap := componentDefinitionsFromJS(p[1])
	exprs, diagnostics, raw := parseSimulationSource(sourceCode, componentContextMap, reverseComponentMap)
	if raw != "" {
		return simulationResult(false, 0, "compile", raw, diagnostics, nil)
	}

	tinyDBInitial := map[string]map[string]runtime.Value{}
	if len(p) >= 4 {
		tinyDBInitial = jsTinyDBStores(p[3])
	}
	host := newSimulationHost(jsInitialState(p[2]), reverseComponentMap, tinyDBInitial)
	interp := runtime.NewInterpreterWithOutput(func(line string) {
		host.logs = append(host.logs, line)
	})
	interp.SetComponentHost(host)
	session := &simulationSession{
		id:      nextSimulationSessionID,
		interp:  interp,
		host:    host,
		events:  map[simulationEventKey]*components.Event{},
		generic: map[simulationGenericEventKey]*components.GenericEvent{},
	}
	host.runEvent = func(componentName, componentType, eventName string, args []runtime.Value) bool {
		componentType = host.componentType(componentName, componentType)
		handled := false
		if event := session.events[simulationEventKey{component: componentName, event: eventName}]; event != nil {
			session.interp.RunBodyWithLocals(event.Body, event.Parameters, args)
			handled = true
		}
		if event := session.generic[simulationGenericEventKey{componentType: componentType, event: eventName}]; event != nil {
			genericArgs := append([]runtime.Value{
				runtime.StrVal(componentName),
				runtime.BoolVal(!handled),
			}, args...)
			session.interp.RunBodyWithLocals(event.Body, event.Parameters, genericArgs)
		}
		return handled
	}
	nextSimulationSessionID++

	for _, expr := range exprs {
		if !interp.RegisterTopLevelDefinition(expr) {
			return simulationResult(false, 0, "compile", "simulation v1 does not support free top-level executable statements", fallbackDiagnostic("simulation v1 does not support free top-level executable statements", "compile", "simulation"), nil)
		}
		if event, ok := expr.(*components.Event); ok {
			session.events[simulationEventKey{component: event.ComponentName, event: event.Event}] = event
		}
		if event, ok := expr.(*components.GenericEvent); ok {
			session.generic[simulationGenericEventKey{componentType: event.ComponentType, event: event.Event}] = event
		}
	}

	var runtimeRaw string
	var runtimeDiagnostics []wasmDiagnostic
	func() {
		defer func() {
			if r := recover(); r != nil {
				runtimeRaw = interp.FormatRuntimeError(r)
				runtimeDiagnostics = wasmDiagnosticsFromContext(interp.DiagnosticFromRuntimeError(r), "runtime", runtimeRaw, "simulation")
			}
		}()
		for _, expr := range exprs {
			if global, ok := expr.(*variables.Global); ok {
				interp.Eval(global)
			}
		}
		screenName := "Screen1"
		if names := componentContextMap["Screen"]; len(names) > 0 {
			screenName = names[0]
		}
		if event := session.events[simulationEventKey{component: screenName, event: "Initialize"}]; event != nil {
			interp.RunBody(event.Body)
		} else {
			for key, event := range session.events {
				if key.event == "Initialize" && (event.ComponentType == "Screen" || event.ComponentType == "Form") {
					interp.RunBody(event.Body)
					break
				}
			}
		}
	}()
	if runtimeRaw != "" {
		return simulationResult(false, 0, "runtime", runtimeRaw, runtimeDiagnostics, host.payloadFields())
	}

	simulationSessions[session.id] = session
	fields := host.payloadFields()
	host.resetTransient()
	return simulationResult(true, session.id, "runtime", "", nil, fields)
}

func setSimulationProperty(this js.Value, p []js.Value) any {
	if len(p) < 4 {
		return simulationResult(false, 0, "runtime", "setSimulationProperty(sessionId, component, property, value) not provided", nil, nil)
	}
	sessionID := p[0].Int()
	session := simulationSessions[sessionID]
	if session == nil {
		return simulationResult(false, sessionID, "runtime", "simulation session not found", nil, nil)
	}
	session.host.resetTransient()
	component := p[1].String()
	property := p[2].String()
	previousSuppressTextChanged := session.host.suppressTextChanged
	session.host.suppressTextChanged = true
	defer func() { session.host.suppressTextChanged = previousSuppressTextChanged }()
	session.host.SetProperty(component, "", property, jsToRuntimeValue(p[3]))
	fields := session.host.payloadFields()
	session.host.resetTransient()
	return simulationResult(true, sessionID, "runtime", "", nil, fields)
}

func dispatchSimulationEvent(this js.Value, p []js.Value) any {
	if len(p) < 3 {
		return simulationResult(false, 0, "runtime", "dispatchSimulationEvent(sessionId, component, event, args) not provided", nil, nil)
	}
	sessionID := p[0].Int()
	session := simulationSessions[sessionID]
	if session == nil {
		return simulationResult(false, sessionID, "runtime", "simulation session not found", nil, nil)
	}
	componentName := p[1].String()
	eventName := p[2].String()
	event := session.events[simulationEventKey{component: componentName, event: eventName}]
	componentType := session.host.componentType(componentName, "")
	genericEvent := session.generic[simulationGenericEventKey{componentType: componentType, event: eventName}]

	session.host.resetTransient()
	var runtimeRaw string
	var runtimeDiagnostics []wasmDiagnostic
	func() {
		defer func() {
			if r := recover(); r != nil {
				runtimeRaw = session.interp.FormatRuntimeError(r)
				runtimeDiagnostics = wasmDiagnosticsFromContext(session.interp.DiagnosticFromRuntimeError(r), "runtime", runtimeRaw, "simulation")
			}
		}()
		args := jsEventArgs(p)
		handled := false
		if event != nil {
			session.interp.RunBodyWithLocals(event.Body, event.Parameters, args)
			handled = true
		}
		if genericEvent != nil {
			genericArgs := append([]runtime.Value{
				runtime.StrVal(componentName),
				runtime.BoolVal(!handled),
			}, args...)
			session.interp.RunBodyWithLocals(genericEvent.Body, genericEvent.Parameters, genericArgs)
		}
	}()
	fields := session.host.payloadFields()
	session.host.resetTransient()
	if runtimeRaw != "" {
		return simulationResult(false, sessionID, "runtime", runtimeRaw, runtimeDiagnostics, fields)
	}
	return simulationResult(true, sessionID, "runtime", "", nil, fields)
}

func jsEventArgs(p []js.Value) []runtime.Value {
	if len(p) < 4 || p[3].IsUndefined() || p[3].IsNull() {
		return nil
	}
	argsValue := p[3]
	if !js.Global().Get("Array").Call("isArray", argsValue).Bool() {
		return []runtime.Value{jsToRuntimeValue(argsValue)}
	}
	args := make([]runtime.Value, argsValue.Length())
	for i := 0; i < argsValue.Length(); i++ {
		args[i] = jsToRuntimeValue(argsValue.Index(i))
	}
	return args
}

func disposeSimulationSession(this js.Value, p []js.Value) any {
	if len(p) < 1 {
		return simulationResult(false, 0, "runtime", "disposeSimulationSession(sessionId) not provided", nil, nil)
	}
	sessionIDText := strings.TrimSpace(p[0].String())
	sessionID, _ := strconv.Atoi(sessionIDText)
	if sessionID == 0 && p[0].Type() == js.TypeNumber {
		sessionID = p[0].Int()
	}
	delete(simulationSessions, sessionID)
	return simulationResult(true, sessionID, "runtime", "", nil, nil)
}
