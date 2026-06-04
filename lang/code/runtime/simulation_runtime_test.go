package runtime

import (
	"Falcon/code/ast"
	"Falcon/code/ast/components"
	"Falcon/code/ast/variables"
	"Falcon/code/compdb"
	"Falcon/code/context"
	"Falcon/code/lex"
	"Falcon/code/parsers/mistparser"
	"strings"
	"testing"
)

type componentHostTestDouble struct {
	state       map[string]map[string]Value
	effects     []string
	unsupported []string
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
