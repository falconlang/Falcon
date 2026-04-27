package method

import (
	"Falcon/code/ast"
	"Falcon/code/fzf"
)

// flattenChain flattens a nested method chain into left-to-right call order
// and returns the root receiver (the non-Call expression at the base).
//
// For s.a().b().c() (stored as c{b{a{s}}}), returns ([a, b, c], s).
func flattenChain(call *Call) ([]*Call, ast.Expr) {
	var chain []*Call
	curr := ast.Expr(call)
	for {
		if c, ok := curr.(*Call); ok {
			chain = append(chain, c)
			curr = c.On
		} else {
			break
		}
	}
	for i, j := 0, len(chain)-1; i < j; i, j = i+1, j-1 {
		chain[i], chain[j] = chain[j], chain[i]
	}
	return chain, curr
}

// safeRootSignature returns the root expression's output signature,
// falling back to [SignAny] if Signature() panics.
func safeRootSignature(expr ast.Expr) (sigs []ast.Signature) {
	defer func() {
		if recover() != nil {
			sigs = []ast.Signature{ast.SignAny}
		}
	}()
	return expr.Signature()
}

// moduleMatchesSig reports whether the given module ("text"/"list"/"dict")
// is compatible with the provided signatures.
func moduleMatchesSig(module string, sigs []ast.Signature) bool {
	switch module {
	case "text":
		return ast.HasSignature(sigs, ast.SignText)
	case "list":
		return ast.HasSignature(sigs, ast.SignList)
	case "dict":
		return ast.HasSignature(sigs, ast.SignDict)
	}
	return false
}

// signatureForModule converts a module name to its corresponding Signature constant.
func signatureForModule(module string) ast.Signature {
	switch module {
	case "text":
		return ast.SignText
	case "list":
		return ast.SignList
	case "dict":
		return ast.SignDict
	}
	return ast.SignAny
}

// minCorrectionScore is the minimum fzf.Score a candidate must reach to be
// considered as a correction. Using a lower threshold than fzf.Rank (0.2) so
// that prefix-match cases like "upper" → "uppercase" (score ≈ 0.18) are accepted.
const minCorrectionScore = 0.1

// findBestCorrection returns the best replacement method name for wrongName such
// that the replacement accepts inputSigs and (when neededOutput is non-nil) produces
// a compatible output type. Falls back to ignoring the output constraint if the
// constrained search yields nothing.
func findBestCorrection(wrongName string, inputSigs []ast.Signature, neededOutput *ast.Signature) string {
	bestScore := -1.0
	bestName := ""
	for name, sig := range signatures {
		if name == wrongName {
			continue
		}
		if !moduleMatchesSig(sig.Module, inputSigs) {
			continue
		}
		if neededOutput != nil && sig.Signature != *neededOutput && sig.Signature != ast.SignAny {
			continue
		}
		if s := fzf.Score(wrongName, name); s > bestScore {
			bestScore = s
			bestName = name
		}
	}
	if bestScore >= minCorrectionScore {
		return bestName
	}
	if neededOutput != nil {
		return findBestCorrection(wrongName, inputSigs, nil)
	}
	return ""
}

// tryDecomposeAnd attempts to split name into a sequence of valid method names
// using camelCase "And" (capital A) as the joiner delimiter.
//
// For example, "allButLastAndListLen" with inputSig=[SignList] returns
// ["allButLast", "listLen"] because the types chain correctly.
// Returns nil when no valid decomposition is found.
func tryDecomposeAnd(name string, inputSig []ast.Signature) []string {
	// Base case: name is already a valid single method for this inputSig.
	if sig, ok := signatures[name]; ok && moduleMatchesSig(sig.Module, inputSig) {
		return []string{name}
	}
	// Try each "And" split point (capital A, preceded by a lowercase letter).
	for i := 1; i+3 <= len(name); i++ {
		if name[i] != 'A' || name[i+1] != 'n' || name[i+2] != 'd' {
			continue
		}
		if name[i-1] < 'a' || name[i-1] > 'z' {
			continue
		}
		rightRaw := name[i+3:]
		if len(rightRaw) == 0 {
			continue
		}
		left := name[:i]
		// Lowercase the first character of the right part (join convention:
		// "allButLast" + "listLen" → "allButLastAndListLen", so "ListLen" → "listLen").
		firstChar := rightRaw[0]
		if firstChar >= 'A' && firstChar <= 'Z' {
			firstChar += 'a' - 'A'
		}
		right := string(firstChar) + rightRaw[1:]

		leftSig, leftOk := signatures[left]
		if !leftOk || !moduleMatchesSig(leftSig.Module, inputSig) {
			continue
		}
		leftOutput := []ast.Signature{leftSig.Signature}
		rest := tryDecomposeAnd(right, leftOutput)
		if rest != nil {
			return append([]string{left}, rest...)
		}
	}
	return nil
}

// applyDecomposition rewrites call c in-place to represent the last method in
// parts, inserting new Call nodes for all earlier parts between c.On and c.
//
// For parts = ["allButLast", "listLen"] and c originally being the merged call:
//
//	c.On stays pointing to the original receiver
//	a new Call{On: c.On, Name: "allButLast"} is created
//	c.On is updated to the new inner call, c.Name = "listLen"
func applyDecomposition(c *Call, parts []string) {
	inner := c.On
	for _, part := range parts[:len(parts)-1] {
		inner = &Call{Where: c.Where, On: inner, Name: part, Args: []ast.Expr{}}
	}
	c.On = inner
	c.Name = parts[len(parts)-1]
}

// CorrectChain inspects the full method chain rooted at call and rewrites any
// method names that are mismatched with their receiver types. Two strategies
// are applied in order for each invalid call:
//
//  1. And-decomposition: a merged name like "allButLastAndListLen" is split
//     into two properly chained calls using "And" as the camelCase joiner.
//  2. Fuzzy rename: an unknown or wrong-module name like "upper" or "reverse"
//     (on a list) is replaced with the closest matching valid method.
//
// Returns true if at least one correction was made.
func CorrectChain(call *Call) bool {
	chain, root := flattenChain(call)
	if len(chain) == 0 {
		return false
	}

	inputSig := safeRootSignature(root)
	corrected := false

	for i, c := range chain {
		sig, exists := signatures[c.Name]
		isValid := exists && moduleMatchesSig(sig.Module, inputSig)

		if !isValid {
			// Strategy 1: try to split a merged "And"-joined name into two calls.
			if parts := tryDecomposeAnd(c.Name, inputSig); len(parts) >= 2 {
				applyDecomposition(c, parts)
				sig = signatures[c.Name]
				exists = sig != nil
				corrected = true
			} else {
				// Strategy 2: fuzzy rename — find the closest valid single method.
				// Scan forward past any consecutive invalid calls to find the first
				// valid one; its input module constrains what this call must output.
				var neededOutput *ast.Signature
				for j := i + 1; j < len(chain); j++ {
					if nextSig, ok := signatures[chain[j].Name]; ok {
						needed := signatureForModule(nextSig.Module)
						neededOutput = &needed
						break
					}
				}
				c.hintOutput = neededOutput // retained for error message generation if correction fails
				if bestName := findBestCorrection(c.Name, inputSig, neededOutput); bestName != "" {
					c.Name = bestName
					sig = signatures[bestName]
					exists = true
					corrected = true
				}
			}
		}

		if exists && sig != nil {
			inputSig = []ast.Signature{sig.Signature}
		} else {
			inputSig = []ast.Signature{ast.SignAny}
		}
	}

	return corrected
}
