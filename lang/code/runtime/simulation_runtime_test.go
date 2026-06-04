package runtime

import (
	"Falcon/code/ast"
	"Falcon/code/ast/components"
	"Falcon/code/ast/fundamentals"
	"Falcon/code/ast/variables"
	"Falcon/code/compdb"
	"Falcon/code/context"
	"Falcon/code/lex"
	"Falcon/code/parsers/mistparser"
	"fmt"
	"sort"
	"strings"
	"testing"
)

type componentHostTestDouble struct {
	state          map[string]map[string]Value
	effects        []string
	unsupported    []string
	componentTypes map[string]string
	componentNames map[string][]string
}

var defaultSimulationRuntimeComponentTypes = map[string]string{
	"Screen1":             "Screen",
	"AddButton":           "Button",
	"Level":               "Slider",
	"firstNumberTextBox":  "TextBox",
	"secondNumberTextBox": "TextBox",
	"Notifier1":           "Notifier",
}

func (h *componentHostTestDouble) ComponentType(componentName string) string {
	if h.componentTypes != nil {
		return h.componentTypes[componentName]
	}
	return defaultSimulationRuntimeComponentTypes[componentName]
}

func (h *componentHostTestDouble) ComponentNames(componentType string) []string {
	if h.componentNames != nil {
		return append([]string(nil), h.componentNames[componentType]...)
	}
	types := h.componentTypes
	if types == nil {
		types = defaultSimulationRuntimeComponentTypes
	}
	names := make([]string, 0)
	for name, typ := range types {
		if typ == componentType {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

func (h *componentHostTestDouble) GetProperty(componentName, componentType, property string) Value {
	if props := h.state[componentName]; props != nil {
		if value, ok := props[property]; ok {
			return value
		}
	}
	return NullVal()
}

func (h *componentHostTestDouble) SetProperty(componentName, componentType, property string, value Value) {
	if h.state[componentName] == nil {
		h.state[componentName] = map[string]Value{}
	}
	h.state[componentName][property] = value
}

func (h *componentHostTestDouble) CallMethod(componentName, componentType, method string, args []Value) Value {
	if componentType == "Notifier" && method == "ShowAlert" {
		if len(args) > 0 {
			h.effects = append(h.effects, args[0].AsStr())
		}
		return VoidVal()
	}
	h.Unsupported("method", componentName+"."+method)
	return VoidVal()
}

func (h *componentHostTestDouble) Unsupported(kind, detail string) {
	h.unsupported = append(h.unsupported, kind+":"+detail)
}

func TestSimulationRuntimeGenericComponentBlocksDelegateToHost(t *testing.T) {
	host := &componentHostTestDouble{
		state: map[string]map[string]Value{
			"firstNumberTextBox": {"Text": StrVal("")},
			"Notifier1":          {},
		},
		componentNames: map[string][]string{
			"Button": {"AddButton"},
		},
	}
	interp := NewInterpreter()
	interp.SetComponentHost(host)

	interp.Eval(&components.GenericPropertySet{
		Component:     &fundamentals.Component{Name: "firstNumberTextBox", Type: "TextBox"},
		ComponentType: "TextBox",
		Property:      "Text",
		Value:         &fundamentals.Text{Content: "generic"},
	})
	got := interp.Eval(&components.GenericPropertyGet{
		Component:     &fundamentals.Component{Name: "firstNumberTextBox", Type: "TextBox"},
		ComponentType: "TextBox",
		Property:      "Text",
	})
	if got.AsStr() != "generic" {
		t.Fatalf("generic Text get = %q, want generic", got.AsStr())
	}

	interp.RunBodyWithLocals([]ast.Expr{
		&components.GenericPropertySet{
			Component:     &variables.Get{Name: "component"},
			ComponentType: "TextBox",
			Property:      "Text",
			Value:         &fundamentals.Text{Content: "from event component"},
		},
	}, []string{"component"}, []Value{StrVal("firstNumberTextBox")})
	if got := host.state["firstNumberTextBox"]["Text"].AsStr(); got != "from event component" {
		t.Fatalf("generic event component Text = %q, want from event component", got)
	}

	interp.Eval(&components.GenericMethodCall{
		Component:     &fundamentals.Component{Name: "Notifier1", Type: "Notifier"},
		ComponentType: "Notifier",
		Method:        "ShowAlert",
		Args:          []ast.Expr{&fundamentals.Text{Content: "hello"}},
	})
	if len(host.effects) != 1 || host.effects[0] != "hello" {
		t.Fatalf("generic method effects = %#v, want hello", host.effects)
	}

	componentsList := interp.Eval(&components.EveryComponent{Type: "Button"}).AsList()
	if len(*componentsList) != 1 || (*componentsList)[0].AsStr() != "AddButton" {
		t.Fatalf("every(Button) = %s, want [AddButton]", ListVal(*componentsList).String())
	}
}

func TestSimulationRuntimeGenericHelpersDelegateToHost(t *testing.T) {
	host := &componentHostTestDouble{state: map[string]map[string]Value{
		"firstNumberTextBox": {"Text": StrVal("")},
		"Level":              {"ThumbPosition": NumVal(7)},
		"Notifier1":          {},
	}}
	interp, events := newSimulationRuntimeTestSession(t, strings.Join([]string{
		`when AddButton.Click {`,
		`  set("TextBox", firstNumberTextBox, "Text", get("Slider", Level, "ThumbPosition"))`,
		`  call("Notifier", Notifier1, "ShowAlert", get("TextBox", firstNumberTextBox, "Text"))`,
		`}`,
	}, "\n"), host)

	interp.RunBody(events["AddButton.Click"].Body)

	if got := host.state["firstNumberTextBox"]["Text"].AsStr(); got != "7" {
		t.Fatalf("generic helper Text = %q, want 7", got)
	}
	if len(host.effects) != 1 || host.effects[0] != "7" {
		t.Fatalf("generic helper effects = %#v, want 7", host.effects)
	}
}

func TestSimulationRuntimeGenericBlockRejectsWrongComponentType(t *testing.T) {
	host := &componentHostTestDouble{state: map[string]map[string]Value{
		"firstNumberTextBox": {"Text": StrVal("")},
	}}
	interp := NewInterpreter()
	interp.SetComponentHost(host)

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("generic property get panic = nil, want component type mismatch")
		}
		if !strings.Contains(fmt.Sprint(r), "expected component of type Button but got TextBox") {
			t.Fatalf("generic property get panic = %v, want type mismatch", r)
		}
	}()

	interp.Eval(&components.GenericPropertyGet{
		Component:     &fundamentals.Component{Name: "firstNumberTextBox", Type: "TextBox"},
		ComponentType: "Button",
		Property:      "Text",
	})
}

func parseSimulationRuntimeTestSource(t *testing.T, src string) []ast.Expr {
	t.Helper()
	ctx := &context.CodeContext{SourceCode: &src, FileName: "simulation_runtime_test.mist"}
	parser := mistparser.NewLangParser(true, lex.NewLexer(ctx).Lex())
	parser.SetComponentDefinitions(
		map[string][]string{
			"Screen":   {"Screen1"},
			"Button":   {"AddButton"},
			"Slider":   {"Level"},
			"TextBox":  {"firstNumberTextBox", "secondNumberTextBox"},
			"Notifier": {"Notifier1"},
		},
		map[string]string{
			"Screen1":             "Screen",
			"AddButton":           "Button",
			"Level":               "Slider",
			"firstNumberTextBox":  "TextBox",
			"secondNumberTextBox": "TextBox",
			"Notifier1":           "Notifier",
		},
	)
	parser.SetEventValidator(compdb.GlobalDB.ValidateEvent)
	parser.SetPropertyValidator(compdb.GlobalDB.ValidateProperty)
	parser.SetMethodValidator(compdb.GlobalDB.ValidateMethod)
	return parser.ParseAll()
}

func newSimulationRuntimeTestSession(t *testing.T, src string, host *componentHostTestDouble) (*Interpreter, map[string]*components.Event) {
	t.Helper()
	exprs := parseSimulationRuntimeTestSource(t, src)
	interp := NewInterpreter()
	interp.SetComponentHost(host)
	events := map[string]*components.Event{}

	for _, expr := range exprs {
		if !interp.RegisterTopLevelDefinition(expr) {
			t.Fatalf("unexpected executable top-level expression %T", expr)
		}
		if event, ok := expr.(*components.Event); ok {
			events[event.ComponentName+"."+event.Event] = event
		}
	}
	for _, expr := range exprs {
		if _, ok := expr.(*variables.Global); ok {
			interp.Eval(expr)
		}
	}
	return interp, events
}

func TestSimulationRuntimeRegistersEventsWithoutExecuting(t *testing.T) {
	host := &componentHostTestDouble{state: map[string]map[string]Value{
		"firstNumberTextBox": {"Text": StrVal("")},
	}}
	_, events := newSimulationRuntimeTestSession(t, strings.Join([]string{
		`when AddButton.Click {`,
		`  firstNumberTextBox.Text = "clicked"`,
		`}`,
	}, "\n"), host)

	if events["AddButton.Click"] == nil {
		t.Fatal("AddButton.Click was not registered")
	}
	if got := host.state["firstNumberTextBox"]["Text"].AsStr(); got != "" {
		t.Fatalf("event executed during registration; Text = %q", got)
	}
}

func TestSimulationRuntimeButtonClickMutatesTextBoxText(t *testing.T) {
	host := &componentHostTestDouble{state: map[string]map[string]Value{
		"firstNumberTextBox":  {"Text": StrVal("2")},
		"secondNumberTextBox": {"Text": StrVal("3")},
	}}
	interp, events := newSimulationRuntimeTestSession(t, strings.Join([]string{
		`when AddButton.Click {`,
		`  firstNumberTextBox.Text = firstNumberTextBox.Text + secondNumberTextBox.Text`,
		`}`,
	}, "\n"), host)

	interp.RunBody(events["AddButton.Click"].Body)

	if got := host.state["firstNumberTextBox"]["Text"].AsStr(); got != "5" {
		t.Fatalf("Text = %q, want 5", got)
	}
}

func TestSimulationRuntimeScreenInitializeAndNotifierEffect(t *testing.T) {
	host := &componentHostTestDouble{state: map[string]map[string]Value{
		"firstNumberTextBox": {"Text": StrVal("")},
		"Notifier1":          {},
	}}
	interp, events := newSimulationRuntimeTestSession(t, strings.Join([]string{
		`when Screen1.Initialize {`,
		`  firstNumberTextBox.Text = "ready"`,
		`  Notifier1.ShowAlert("Started")`,
		`}`,
	}, "\n"), host)

	interp.RunBody(events["Screen1.Initialize"].Body)

	if got := host.state["firstNumberTextBox"]["Text"].AsStr(); got != "ready" {
		t.Fatalf("Text = %q, want ready", got)
	}
	if len(host.effects) != 1 || host.effects[0] != "Started" {
		t.Fatalf("effects = %#v, want Started", host.effects)
	}
}

func TestSimulationRuntimeEventArgumentsBindAsLocals(t *testing.T) {
	host := &componentHostTestDouble{state: map[string]map[string]Value{
		"firstNumberTextBox": {"Text": StrVal("")},
	}}
	interp, events := newSimulationRuntimeTestSession(t, strings.Join([]string{
		`when Level.PositionChanged(thumbPosition) {`,
		`  firstNumberTextBox.Text = thumbPosition`,
		`}`,
	}, "\n"), host)
	event := events["Level.PositionChanged"]
	if event == nil {
		t.Fatal("Level.PositionChanged was not registered")
	}

	interp.RunBodyWithLocals(event.Body, event.Parameters, []Value{NumVal(7)})

	if got := host.state["firstNumberTextBox"]["Text"].String(); got != "7" {
		t.Fatalf("Text = %q, want 7", got)
	}
}
