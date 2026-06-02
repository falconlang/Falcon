package runtime

import (
	"Falcon/code/context"
	"Falcon/code/lex"
	"Falcon/code/parsers/mistparser"
	"strings"
	"testing"
)

func TestAnonymousProcedureRuntimeReturnCallForms(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want string
	}{
		{
			name: "direct call",
			src: strings.Join([]string{
				"local inc = func(x) = x + 1",
				"inc(4)",
			}, "\n"),
			want: "5",
		},
		{
			name: "input list call",
			src: strings.Join([]string{
				"local double = func(x) = x * 2",
				"double.call([5])",
			}, "\n"),
			want: "10",
		},
		{
			name: "immediate call",
			src:  "(func(x) = x * 3)(4)",
			want: "12",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := evalProcedureTestSource(t, tt.src)
			if got.String() != tt.want {
				t.Fatalf("result = %q, want %q", got.String(), tt.want)
			}
		})
	}
}

func TestAnonymousProcedureRuntimeVoidCalls(t *testing.T) {
	src := strings.Join([]string{
		`local greet = func(x) { println("Hello " _ x) }`,
		`greet.call(["Melon"])`,
	}, "\n")
	var output []string
	runProcedureTestSourceWithOutput(t, src, func(line string) {
		output = append(output, line)
	})
	if got := strings.Join(output, "\n"); got != "Hello Melon" {
		t.Fatalf("output = %q, want %q", got, "Hello Melon")
	}
}

func TestProcedureRuntimeHelpers(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want string
	}{
		{
			name: "num args",
			src: strings.Join([]string{
				"local add = func(a, b) = a + b",
				"add.numArgs()",
			}, "\n"),
			want: "2",
		},
		{
			name: "get with name",
			src: strings.Join([]string{
				"func add(x) = x + 1",
				`local f = getFunc("add")`,
				"f(9)",
			}, "\n"),
			want: "10",
		},
		{
			name: "get with dropdown",
			src: strings.Join([]string{
				"func add(x) = x + 1",
				"func.add(9)",
			}, "\n"),
			want: "10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := evalProcedureTestSource(t, tt.src)
			if got.String() != tt.want {
				t.Fatalf("result = %q, want %q", got.String(), tt.want)
			}
		})
	}
}

func TestAnonymousProcedureRuntimeCapturesLexicalScope(t *testing.T) {
	src := strings.Join([]string{
		"local makeAdder = func(x) = {",
		"  local add = func(y) = x + y",
		"  add",
		"}",
		"local addFive = makeAdder(5)",
		"addFive(7)",
	}, "\n")
	got := evalProcedureTestSource(t, src)
	if got.String() != "12" {
		t.Fatalf("closure result = %q, want 12", got.String())
	}
}

func TestAnonymousProcedureRuntimeErrors(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want string
	}{
		{
			name: "arity",
			src: strings.Join([]string{
				"local add = func(a, b) = a + b",
				"add(1)",
			}, "\n"),
			want: "expects 2 argument(s) but got 1",
		},
		{
			name: "non procedure call",
			src:  "5.call([1])",
			want: "expected a procedure value but got number 5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errMsg := recoverProcedureTestSource(t, tt.src)
			if !strings.Contains(errMsg, tt.want) {
				t.Fatalf("runtime error = %q, want substring %q", errMsg, tt.want)
			}
		})
	}
}

func evalProcedureTestSource(t *testing.T, src string) Value {
	t.Helper()
	return runProcedureTestSourceWithOutput(t, src, nil)
}

func runProcedureTestSourceWithOutput(t *testing.T, src string, output func(string)) Value {
	t.Helper()
	ctx := &context.CodeContext{SourceCode: &src, FileName: "procedures_test.mist"}
	parser := mistparser.NewLangParser(false, lex.NewLexer(ctx).Lex())
	interp := NewInterpreterWithOutput(output)
	return interp.RunGetLast(parser.ParseAll())
}

func recoverProcedureTestSource(t *testing.T, src string) (message string) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			message = NewInterpreter().FormatRuntimeError(r)
		}
	}()
	evalProcedureTestSource(t, src)
	t.Fatal("expected runtime panic")
	return ""
}
