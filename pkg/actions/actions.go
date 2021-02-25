package actions

import (
	"os"
	"strings"
)

// GetInput gets the input by the given name.
func GetInput(name string) string {
	e := strings.ReplaceAll(name, " ", "_")
	e = strings.ToUpper(e)
	e = "INPUT_" + e

	return strings.TrimSpace(os.Getenv(e))
}
