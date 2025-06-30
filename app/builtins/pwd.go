package builtins

import (
	"fmt"
	"os"
	"shell-GO/app/config"
)

func Pwd(args []string, out *config.Output) {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(out.Stderr, "Error writing to stdout:", err)
	} else {
		fmt.Fprintln(out.Stdout, dir)
	}
}
