package blocklytoyail

import "Falcon/code/compdb"

// globalDB wraps the shared component DB for use within this package.
var globalDB = compdb.GlobalDB

// getDropdownYAIL returns the YAIL for a helpers_dropdown block. Always emits
// the REPL form: (protect-enum (static-field FQCN "optName") concreteValue).
// quoteStr is defined in parser.go and handles AppInventor REPL string escaping.
func getDropdownYAIL(key, optName string) string {
	ol := globalDB.GetOptionList(key)
	if ol == nil {
		return "(static-field " + key + " \"" + optName + "\")"
	}

	staticField := "(static-field " + ol.ClassName + " \"" + optName + "\")"

	concreteVal := ol.Options[optName]
	if concreteVal == "" {
		concreteVal = "0"
	}
	if ol.UnderlyingType == "java.lang.String" {
		concreteVal = quoteStr(concreteVal)
	}

	return "(protect-enum " + staticField + " " + concreteVal + ")"
}
