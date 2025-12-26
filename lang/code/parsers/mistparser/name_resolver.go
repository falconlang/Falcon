package mistparser

import (
	"Falcon/code/sugar"
	"strconv"
)

type NameResolver struct {
	Procedures        map[string]*Procedure
	ComponentTypesMap map[string]string // Button1 -> Button
	ComponentNameMap  map[string][]string
}

type Procedure struct {
	Name       string
	Parameters []string
	Returning  bool
}

func (n *NameResolver) ResolveProcedure(name string, argsCount int) (string, *Procedure) {
	procedure, found := n.Procedures[name]
	if found {
		if len(procedure.Parameters) != argsCount {
			return sugar.Format(
				"Expected % args but got % for procedure %()",
				strconv.Itoa(len(procedure.Parameters)), strconv.Itoa(argsCount), name), nil
		}
		return "", procedure
	}
	return sugar.Format("Did not find procedure %()", name), nil
}
