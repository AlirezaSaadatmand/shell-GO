package main

import (
	"fmt"
	"log"
	"os"
	"shell-GO/app/builtins"
	"shell-GO/app/config"
	"shell-GO/app/internal/completer"
	"shell-GO/app/internal/executor"
	"shell-GO/app/internal/history"
	"shell-GO/app/internal/parser"
	"strings"

	"github.com/chzyer/readline"
)

func init() {
	config.Commands = map[string]func([]string, *config.Output){
		"exit":    builtins.Exit,
		"echo":    builtins.Echo,
		"type":    builtins.Type_,
		"pwd":     builtins.Pwd,
		"cd":      builtins.Cd,
		"history": builtins.History,
	}
	config.Builtin = []string{
		"exit",
		"echo",
		"type",
		"pwd",
		"cd",
		"history",
	}
}
func main() {
	if config.HistFile != "" {
		history.LoadHistoryFromFile(config.HistFile)
	}
	rl, err := readline.NewEx(&readline.Config{
		Prompt:          "$ ",
		HistoryFile:     "/tmp/readline.tmp",
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
		AutoComplete:    &completer.AutoCompleter{},
	})
	if err != nil {
		log.Fatal(err)
	}
	defer rl.Close()
	for {
		line, err := rl.Readline()
		if err != nil {
			break
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		config.ShellHistory = append(config.ShellHistory, line)
		if strings.Contains(line, "|") {
			ok := executor.ExecutePipeline(line)
			if !ok {
				fmt.Fprintln(os.Stderr, "pipeline execution failed")
			}
		} else {
			command, args := parser.SeparateCommandArgs(line)
			args, stdoutRedir, stderrRedir := parser.ParseRedirections(args)
			out, err := parser.SetupOutput(stdoutRedir, stderrRedir)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Redirection error:", err)
				continue
			}

			if builtinFunc, ok := config.Commands[command]; ok {
				builtinFunc(args, out)
			} else {
				if !executor.Execute(command, args, out) {
					fmt.Fprintln(os.Stderr, command+": command not found")
				}
			}
		}
	}
}