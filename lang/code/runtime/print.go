package runtime

import "os"

// printLine writes s followed by a newline to stdout without importing fmt.
func printLine(s string) {
	os.Stdout.WriteString(s)
	os.Stdout.WriteString("\n")
}

// stub prints a warning that the named feature is not supported outside App Inventor.
func stub(feature string) {
	os.Stdout.WriteString("[stub] " + feature + " is not supported outside App Inventor\n")
}