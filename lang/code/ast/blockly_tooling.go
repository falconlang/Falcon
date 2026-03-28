package ast

// DependsOnVariables checks if the expression references any of the variables in the list
func DependsOnVariables(e Expr, variables []string) bool {
	var references []string
	getReferencedVariables(e.Blockly(false), &references)

	for _, reference := range references {
		for _, match := range variables {
			if reference == match {
				return true
			}
		}
	}
	return false
}

func getReferencedVariables(bky Block, currReferences *[]string) {
	if bky.Type == "lexical_variable_get" {
		*currReferences = append(*currReferences, bky.SingleField())
	}

	for _, value := range bky.Values {
		getReferencedVariables(value.Block, currReferences)
	}

	for _, statement := range bky.Statements {
		getReferencedVariables(*statement.Block, currReferences)
	}
}
