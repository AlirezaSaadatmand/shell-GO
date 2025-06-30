package builtins

import (
	"fmt"
	"shell-GO/app/config"
	"strings"
)

func Echo(args []string, out *config.Output) {
	_, err := fmt.Fprintln(out.Stdout, strings.Join(args, " "))
	if err != nil {
		fmt.Fprintln(out.Stderr, "Error writing to stdout:", err)
	}
}
