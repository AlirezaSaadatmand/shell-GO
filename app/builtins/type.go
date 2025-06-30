package builtins

import (
	"fmt"
	"shell-GO/app/config"
	"shell-GO/app/internal/executor"
)

func Type_(args []string, out *config.Output) {

	command := args[0]
	var outputText string
	if _, exists := config.Commands[command]; exists {
		outputText = args[0] + " is a shell builtin"
	} else {
		fullPath := executor.FindExecutable(command, config.Paths)
		if fullPath != "" {
			outputText = command + " is " + fullPath
		} else {
			outputText = args[0] + ": not found"
		}
	}
	_, err := fmt.Fprintln(out.Stdout, outputText)
	if err != nil {
		fmt.Fprintln(out.Stderr, "Error writing to stdout:", err)
	}
}
