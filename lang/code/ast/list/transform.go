package list

import (
	"Falcon/code/ast"
	"Falcon/code/ast/control"
	"Falcon/code/ast/fundamentals"
	"Falcon/code/ast/variables"
	"Falcon/code/lex"
	"Falcon/code/sugar"
	"strconv"
)

type Transformer struct {
	Where       *lex.Token
	List        ast.Expr
	Name        string
	Args        []ast.Expr
	Names       []string
	Transformer ast.Expr
}

type TransformerSignature struct {
	ArgSize  int
	NameSize int
}

func makeSignature(argSize int, nameSize int) *TransformerSignature {
	return &TransformerSignature{ArgSize: argSize, NameSize: nameSize}
}

var transformers = map[string]*TransformerSignature{
	"map":       makeSignature(0, 1),
	"filter":    makeSignature(0, 1),
	"reduce":    makeSignature(1, 2),
	"sort":      makeSignature(0, 2),
	"sortByKey": makeSignature(0, 1),
	"min":       makeSignature(0, 2),
	"max":       makeSignature(0, 2),
}

func IsTransformer(name string, argCount int) bool {
	sig, ok := transformers[name]
	return ok && sig.ArgSize == argCount
}

func TestSignature(transformerName string, argsCount int, namesCount int) (string, *TransformerSignature) {
	signature, ok := transformers[transformerName]
	if !ok {
		return sugar.Format("Unknown list lambda! .% { }", transformerName), nil
	}
	if signature.ArgSize != argsCount {
		return sugar.Format("Expected % args but got % for transformer .% {",
			strconv.Itoa(signature.ArgSize), strconv.Itoa(argsCount), transformerName), nil
	}
	if signature.NameSize != namesCount {
		return sugar.Format("Expected % names but got % for transformer .% {",
			strconv.Itoa(signature.NameSize), strconv.Itoa(namesCount), transformerName), nil
	}
	return "", signature
}

func (t *Transformer) String() string {
	switch t.Transformer.(type) {
	case *control.Do, *variables.VarResult:
		return t.bodyTransformerString(t.Transformer)
	default:
		if sb, ok := t.Transformer.(*fundamentals.SmartBody); ok && len(sb.Body) == 1 {
			switch sb.Body[0].(type) {
			case *variables.VarResult, *variables.Var:
				return t.bodyTransformerString(sb.Body[0])
			}
		}
	}
	return t.singleExprTransformerString()
}

func (t *Transformer) singleExprTransformerString() string {
	if len(t.Args) == 0 {
		pFormat := "%\n  .% { % -> % }"
		if !t.List.Continuous() {
			pFormat = "(%)\n  .% { % -> % } "
		}
		return sugar.Format(pFormat,
			t.List.String(),
			t.Name,
			ast.JoinNames(", ", t.Names),
			t.Transformer.String())
	}
	pFormat := "%\n  .%(%) { % -> % }"
	if !t.List.Continuous() {
		pFormat = "(%)\n  .%(%) { % -> % }"
	}
	return sugar.Format(pFormat,
		t.List.String(),
		t.Name,
		ast.JoinExprs(", ", t.Args),
		ast.JoinNames(", ", t.Names),
		t.Transformer.String())
}

func (t *Transformer) bodyTransformerString(do ast.Expr) string {
	if len(t.Args) == 0 {
		pFormat := "%\n  .% { % -> \n%}"
		if !t.List.Continuous() {
			pFormat = "(%)\n  .% { % -> \n%} "
		}
		return sugar.Format(pFormat,
			t.List.String(),
			t.Name,
			ast.JoinNames(", ", t.Names),
			ast.PadDirect(ast.Pad(do.String())))
	}
	pFormat := "%\n  .%(%) { % -> \n%}"
	if !t.List.Continuous() {
		pFormat = "(%)\n  .%(%) { % -> \n%}"
	}
	return sugar.Format(pFormat,
		t.List.String(),
		t.Name,
		ast.JoinExprs(", ", t.Args),
		ast.JoinNames(", ", t.Names),
		ast.PadDirect(ast.Pad(do.String())))
}

func (t *Transformer) Blockly(flags ...bool) ast.Block {
	errorMessage, signature := TestSignature(t.Name, len(t.Args), len(t.Names))
	if signature == nil {
		t.Where.Error(errorMessage)
	}
	switch t.Name {
	case "map":
		return t.listMap()
	case "filter":
		return t.listFilter()
	case "reduce":
		return t.listReduce()
	case "sort":
		return t.listSort()
	case "sortByKey":
		return t.listSortByKey()
	case "min":
		return t.min()
	case "max":
		return t.max()
	default:
		t.Where.Error("Unknown list lambda! .% { }", t.Name)
		panic("Unreachable")
	}
}

func (t *Transformer) Continuous() bool {
	return true
}

func (t *Transformer) Consumable() bool {
	return true
}

func (t *Transformer) Signature() []ast.Signature {
	t.List.Signature()
	for _, arg := range t.Args {
		arg.Signature()
	}
	t.Transformer.Signature()
	listSigs := t.List.Signature()
	if !ast.HasSignature(listSigs, ast.SignList) {
		t.Where.TypeError("List transformer .% { } requires a list value, but got %", t.Name, ast.FormatSignatures(listSigs))
	}
	errorMessage, transformerSignature := TestSignature(t.Name, len(t.Args), len(t.Names))
	if transformerSignature == nil {
		t.Where.Error(errorMessage)
	}
	// TODO: this has to be improved when we are improving type safety
	if t.Name == "min" || t.Name == "max" || t.Name == "reduce" {
		return []ast.Signature{ast.SignAny}
	}
	return []ast.Signature{ast.SignList}
}

func (t *Transformer) max() ast.Block {
	return ast.Block{
		Type: "lists_maximum_value",
		Fields: []ast.Field{
			{Name: "VAR1", Value: t.Names[0]},
			{Name: "VAR2", Value: t.Names[1]},
		},
		Values: []ast.Value{
			{Name: "LIST", Block: t.List.Blockly(false)},
			{Name: "COMPARE", Block: t.Transformer.Blockly(false)},
		},
	}
}

func (t *Transformer) min() ast.Block {
	return ast.Block{
		Type: "lists_minimum_value",
		Fields: []ast.Field{
			{Name: "VAR1", Value: t.Names[0]},
			{Name: "VAR2", Value: t.Names[1]},
		},
		Values: []ast.Value{
			{Name: "LIST", Block: t.List.Blockly(false)},
			{Name: "COMPARE", Block: t.Transformer.Blockly(false)},
		},
	}
}

func (t *Transformer) listSortByKey() ast.Block {
	return ast.Block{
		Type:   "lists_sort_key",
		Fields: []ast.Field{{Name: "VAR", Value: t.Names[0]}},
		Values: []ast.Value{
			{Name: "LIST", Block: t.List.Blockly(false)},
			{Name: "KEY", Block: t.Transformer.Blockly(false)},
		},
	}
}

func (t *Transformer) listSort() ast.Block {
	return ast.Block{
		Type: "lists_sort_comparator",
		Fields: []ast.Field{
			{Name: "VAR1", Value: t.Names[0]},
			{Name: "VAR2", Value: t.Names[1]},
		},
		Values: []ast.Value{
			{Name: "LIST", Block: t.List.Blockly(false)},
			{Name: "COMPARE", Block: t.Transformer.Blockly(false)},
		},
	}
}

func (t *Transformer) listReduce() ast.Block {
	return ast.Block{
		Type: "lists_reduce",
		Fields: []ast.Field{
			{Name: "VAR1", Value: t.Names[0]},
			{Name: "VAR2", Value: t.Names[1]},
		},
		Values: []ast.Value{
			{Name: "LIST", Block: t.List.Blockly(false)},
			{Name: "INITANSWER", Block: t.Args[0].Blockly(false)},
			{Name: "COMBINE", Block: t.Transformer.Blockly(false)},
		},
	}
}

func (t *Transformer) listFilter() ast.Block {
	return ast.Block{
		Type:   "lists_filter",
		Fields: []ast.Field{{Name: "VAR", Value: t.Names[0]}},
		Values: []ast.Value{
			{Name: "LIST", Block: t.List.Blockly(false)},
			{Name: "TEST", Block: t.Transformer.Blockly(false)},
		},
	}
}

func (t *Transformer) listMap() ast.Block {
	return ast.Block{
		Type:   "lists_map",
		Fields: []ast.Field{{Name: "VAR", Value: t.Names[0]}},
		Values: []ast.Value{
			{Name: "LIST", Block: t.List.Blockly(false)},
			{Name: "TO", Block: t.Transformer.Blockly(false)},
		},
	}
}
