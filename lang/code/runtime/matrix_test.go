package runtime

import (
	"Falcon/code/context"
	"Falcon/code/lex"
	"Falcon/code/parsers/mistparser"
	"strings"
	"testing"
)

func TestMatrixCellRuntimeGet(t *testing.T) {
	got := evalMatrixTestSource(t, "[[1, 2], [3, 4]][[2, 1]]")
	if got.String() != "3" {
		t.Fatalf("matrix cell get = %q, want 3", got.String())
	}
}

func TestMatrixCellRuntimeSet(t *testing.T) {
	src := strings.Join([]string{
		"global m = [[1, 2], [3, 4]]",
		"m[[2, 1]] = 9",
		"m[[2, 1]]",
	}, "\n")

	got := evalMatrixTestSource(t, src)
	if got.String() != "9" {
		t.Fatalf("matrix cell set = %q, want 9", got.String())
	}
}

func TestMatrixCreateRuntime(t *testing.T) {
	got := evalMatrixTestSource(t, "matrix[[1, 2], [3, 4]]")
	if got.String() != "[[1, 2], [3, 4]]" {
		t.Fatalf("matrix create = %q, want [[1, 2], [3, 4]]", got.String())
	}
}

func TestMatrixArithmeticRuntime(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want string
	}{
		{
			name: "add",
			src:  "matrix[[1, 2], [3, 4]] [+] matrix[[5, 6], [7, 8]]",
			want: "[[6, 8], [10, 12]]",
		},
		{
			name: "subtract",
			src:  "matrix[[5, 6], [7, 8]] [-] matrix[[1, 2], [3, 4]]",
			want: "[[4, 4], [4, 4]]",
		},
		{
			name: "multiply",
			src:  "matrix[[1, 2], [3, 4]] [*] matrix[[5, 6], [7, 8]]",
			want: "[[19, 22], [43, 50]]",
		},
		{
			name: "scalar multiply",
			src:  "matrix[[1, 2], [3, 4]] [*] 2",
			want: "[[2, 4], [6, 8]]",
		},
		{
			name: "power",
			src:  "matrix[[1, 2], [3, 4]] [^] 2",
			want: "[[7, 10], [15, 22]]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := evalMatrixTestSource(t, tt.src)
			if got.String() != tt.want {
				t.Fatalf("matrix %s = %q, want %q", tt.name, got.String(), tt.want)
			}
		})
	}
}

func TestMatrixMethodsRuntime(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want string
	}{
		{
			name: "row",
			src:  "matrix[[1, 2], [3, 4]].row(2)",
			want: "[3, 4]",
		},
		{
			name: "column",
			src:  "matrix[[1, 2], [3, 4]].col(1)",
			want: "[1, 3]",
		},
		{
			name: "dimensions",
			src:  "matrix[[1, 2], [3, 4]].dimension()",
			want: "[2, 2]",
		},
		{
			name: "transpose",
			src:  "matrix[[1, 2], [3, 4]].transpose()",
			want: "[[1, 3], [2, 4]]",
		},
		{
			name: "rotate left",
			src:  "matrix[[1, 2], [3, 4]].rotateLeft()",
			want: "[[2, 4], [1, 3]]",
		},
		{
			name: "rotate right",
			src:  "matrix[[1, 2], [3, 4]].rotateRight()",
			want: "[[3, 1], [4, 2]]",
		},
		{
			name: "inverse",
			src:  "matrix[[4, 7], [2, 6]].inverse()",
			want: "[[0.6, -0.7], [-0.2, 0.4]]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := evalMatrixTestSource(t, tt.src)
			if got.String() != tt.want {
				t.Fatalf("matrix %s = %q, want %q", tt.name, got.String(), tt.want)
			}
		})
	}
}

func TestMatrixCreateMultidimRuntime(t *testing.T) {
	got := evalMatrixTestSource(t, "makeNdArray([2, 2, 2], 9)")
	if got.String() != "[[[9, 9], [9, 9]], [[9, 9], [9, 9]]]" {
		t.Fatalf("makeNdArray = %q, want 2x2x2 matrix of 9s", got.String())
	}
}

func TestMatrixQuestionRuntime(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want string
	}{
		{name: "matrix", src: "matrix[[1, 2], [3, 4]] ? matrix", want: "true"},
		{name: "multidim", src: "makeNdArray([2, 2], 0) ? matrix", want: "true"},
		{name: "plain list", src: "[[1, 2], [3, 4]] ? matrix", want: "false"},
		{name: "ragged", src: "[[1], [2, 3]] ? matrix", want: "false"},
		{name: "non numeric", src: `[[1], ["x"]] ? matrix`, want: "false"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := evalMatrixTestSource(t, tt.src)
			if got.String() != tt.want {
				t.Fatalf("matrix question = %q, want %q", got.String(), tt.want)
			}
		})
	}
}

func evalMatrixTestSource(t *testing.T, src string) Value {
	t.Helper()
	ctx := &context.CodeContext{SourceCode: &src, FileName: "matrix_test.mist"}
	parser := mistparser.NewLangParser(false, lex.NewLexer(ctx).Lex())
	return NewInterpreter().RunGetLast(parser.ParseAll())
}
