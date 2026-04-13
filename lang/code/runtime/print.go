package runtime

import "os"

// printLine writes s followed by a newline. If the interpreter has an output
// callback (e.g. WASM streaming), it is called instead of writing to stdout.
func (i *Interpreter) printLine(s string) {
	if i.outputCallback != nil {
		i.outputCallback(s)
	} else {
		os.Stdout.WriteString(s)
		os.Stdout.WriteString("\n")
	}
}

// stub prints a warning that the named feature is not supported outside App Inventor.
func (i *Interpreter) stub(feature string) {
	msg := "[stub] " + feature + " is not supported outside App Inventor"
	if i.outputCallback != nil {
		i.outputCallback(msg)
	} else {
		os.Stdout.WriteString(msg + "\n")
	}
}
