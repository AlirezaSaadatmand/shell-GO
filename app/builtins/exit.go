package builtins

import (
	"fmt"
	"os"
	"shell-GO/app/config"
	"shell-GO/app/internal/history"
	"strconv"
)

func Exit(args []string, out *config.Output) {
	status := 0
	if len(args) > 1 {
		fmt.Fprintln(out.Stderr, "Error: expected zero or one argument")
		return
	}
	if len(args) == 1 {
		var err error
		status, err = strconv.Atoi(args[0])
		if err != nil {
			fmt.Fprintln(out.Stderr, "Invalid number:", err)
			return
		}
	}
	history.WriteHistoryToFile()
	os.Exit(status)
}
