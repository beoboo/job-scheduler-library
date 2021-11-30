package helpers

import (
	"fmt"
	"strings"
)

func FormatCmdLine(executable string, args ...string) string {
	if len(args) == 0 {
		return executable
	}

	return fmt.Sprintf("%s %s", executable, strings.Join(args, " "))
}
