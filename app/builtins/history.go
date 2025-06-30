package builtins

import (
	"fmt"
	"os"
	"shell-GO/app/config"
	"strconv"
	"strings"
)

func History(args []string, out *config.Output) {
	if len(args) == 0 {
		for i, entry := range config.ShellHistory {
			fmt.Fprintf(out.Stdout, "    %d  %s\n", i+1, entry)
		}
		return
	}

	if len(args) == 2 && args[0] == "-a" {
		path := args[1]
		file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Fprintln(out.Stderr, "history: cannot append to file:", err)
			return
		}
		defer file.Close()

		for _, entry := range config.ShellHistory[config.HistoryAppendIndex:] {
			fmt.Fprintln(file, entry)
		}

		config.HistoryAppendIndex = len(config.ShellHistory)
		return
	}

	if len(args) == 2 && args[0] == "-r" {
		// Read history from file
		path := args[1]
		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintln(out.Stderr, "history: cannot read file:", err)
			return
		}
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				config.ShellHistory = append(config.ShellHistory, line)
			}
		}
		return
	}

	if len(args) == 2 && args[0] == "-w" {
		// Write history to file
		path := args[1]
		file, err := os.Create(path)
		if err != nil {
			fmt.Fprintln(out.Stderr, "history: cannot write file:", err)
			return
		}
		defer file.Close()
		for _, entry := range config.ShellHistory {
			fmt.Fprintln(file, entry)
		}
		// Ensure final newline (Fprintln already does this per line)
		return
	}

	// Show history or history <n>
	total := len(config.ShellHistory)
	count := total
	if len(args) == 1 {
		n, err := strconv.Atoi(args[0])
		if err != nil || n < 0 {
			fmt.Fprintln(out.Stderr, "history: invalid number:", args[0])
			return
		}
		if n < count {
			count = n
		}
	}

	start := total - count
	for i := start; i < total; i++ {
		fmt.Fprintf(out.Stdout, "%5d  %s\n", i+1, config.ShellHistory[i])
	}
}
