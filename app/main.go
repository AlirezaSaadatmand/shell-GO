package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"
)

func findExecutable(command string, paths []string) string {
	for _, dir := range paths {
		fullPath := filepath.Join(dir, command)
		if info, err := os.Stat(fullPath); err == nil && info.Mode().IsRegular() && info.Mode().Perm()&0111 != 0 {
			return fullPath
		}
	}
	return ""
}

func separateArgs(input string) []string {
	var result []string
	var current strings.Builder
	inSingleQuote := false
	inDoubleQuote := false
	i := 0

	for i < len(input) {
		ch := input[i]

		switch ch {
		case '\'':
			if !inSingleQuote && !inDoubleQuote {
				inSingleQuote = true
			} else if inSingleQuote {
				inSingleQuote = false
			} else if inDoubleQuote {
				current.WriteByte(ch)
			}
			i++
		case '"':
			if !inSingleQuote && !inDoubleQuote {
				inDoubleQuote = true
			} else if inDoubleQuote {
				inDoubleQuote = false
			} else if inSingleQuote {
				current.WriteByte(ch)
			}
			i++
		case '\\':
			if !inSingleQuote && !inDoubleQuote {
				current.WriteByte(input[i + 1])
				i += 2
			}
		case ' ', '\t':
			if inSingleQuote || inDoubleQuote{
				current.WriteByte(ch)
				i++
			} else {
				if current.Len() > 0 {
					result = append(result, current.String())
					current.Reset()
				}
				for i < len(input) && unicode.IsSpace(rune(input[i])) {
					i++
				}
			}

		default:
			current.WriteByte(ch)
			i++
		}
	}

	if current.Len() > 0 {
		result = append(result, current.String())
	}

	return result
}

var COMMANDS map[string]func([]string)
var builtin []string
var paths = strings.Split(os.Getenv("PATH"), ":")

func init() {
	COMMANDS = map[string]func([]string){
		"exit": exit,
		"echo": echo,
		"type": type_,
	}
	builtin = []string{
		"exit",
		"echo",
		"type",
	}
}
func main() {
	for {
		fmt.Fprint(os.Stdout, "$ ")

		input, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading input:", err)
			os.Exit(1)
		}
		input = input[:len(input)-1]
		command := strings.Split(input, " ")[0]
		args := separateArgs(strings.Join(strings.Split(input, " ")[1:], " "))

		if _, ok := COMMANDS[command]; ok {
			COMMANDS[command](args)
		} else {
			fullPath := findExecutable(command, paths)
			if fullPath != "" {
				cmd := exec.Command(command, args...)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr

				err := cmd.Run()
				if err != nil {
					fmt.Println("Error:", err)
				}
			} else {
				fmt.Println(command + ": command not found")
			}
		}
	}
}

func exit(args []string) {
	if len(args) != 1 {
		fmt.Println("Error: expected exactly one argument")
		return
	}

	status, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Println("Invalid number:", err)
		return
	}
	os.Exit(status)
}

func echo(args []string) {
	if len(args) < 1 {
		fmt.Println("Error: not enough aguments")
		return
	}
	fmt.Println(strings.Join(args, " "))
}

func type_(args []string) {
	if len(args) != 1 {
		fmt.Println("Error: expected exactly one argument")
		return
	}
	command := args[0]

	if _, exists := COMMANDS[command]; exists {
		fmt.Println(args[0] + " is a shell builtin")
	} else {
		fullPath := findExecutable(command, paths)
		if fullPath != "" {
			fmt.Println(command + " is " + fullPath)
			return
		}
		fmt.Println(args[0] + ": not found")
	}
}
