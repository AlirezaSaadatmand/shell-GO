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

func hasRdirection (args []string) ([]string, string) {
	for i, _ := range args {
		if args[i] == ">" || args[i] == "1>" {
			return args[:i], args[i + 1]
		}
	}
	return args, ""
}

func findExecutable(command string, paths []string) string {
	for _, dir := range paths {
		fullPath := filepath.Join(dir, command)
		if info, err := os.Stat(fullPath); err == nil && info.Mode().IsRegular() && info.Mode().Perm()&0111 != 0 {
			return fullPath
		}
	}
	return ""
}

func separateCommandArgs(input string) (string, []string) {
	var args []string
	var current strings.Builder
	inSingleQuote := false
	inDoubleQuote := false
	i := 0

	for i < len(input) {
		ch := input[i]

		switch ch {
		case '\'':
			if !inDoubleQuote {
				inSingleQuote = !inSingleQuote
			} else {
				current.WriteByte(ch)
			}
			i++
		case '"':
			if !inSingleQuote {
				inDoubleQuote = !inDoubleQuote
			} else {
				current.WriteByte(ch)
			}
			i++
		case '\\':
			if i+1 >= len(input) {
				current.WriteByte('\\')
				i++
				break
			}
			next := input[i+1]
			if inSingleQuote {
				current.WriteByte('\\')
				current.WriteByte(next)
				i += 2
			} else if inDoubleQuote {
				if next == '"' || next == '\\' || next == '$' || next == '`' {
					current.WriteByte(next)
				} else {
					current.WriteByte('\\')
					current.WriteByte(next)
				}
				i += 2
			} else {
				current.WriteByte(next)
				i += 2
			}
		case ' ', '\t':
			if inSingleQuote || inDoubleQuote {
				current.WriteByte(ch)
				i++
			} else {
				if current.Len() > 0 {
					args = append(args, current.String())
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
		args = append(args, current.String())
	}

	if len(args) == 0 {
		return "", []string{}
	}
	
	return args[0], args[1:]
}


var COMMANDS map[string]func([]string, string)
var builtin []string
var paths = strings.Split(os.Getenv("PATH"), ":")

func init() {
	COMMANDS = map[string]func([]string, string){
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
		if strings.TrimSpace(input) == "" {
			continue
		}
		command, args := separateCommandArgs(input[:len(input)-1])
		args, output := hasRdirection(args)
		if _, ok := COMMANDS[command]; ok {
			COMMANDS[command](args,output)
		} else {
			if !execute(command , args, output) {
				fmt.Fprintln(os.Stderr, command+": command not found")
			}
		}
	}
}

func exit(args []string, output string) {
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

func echo(args []string, output string) {
	if len(args) < 1 {
		fmt.Println("Error: not enough aguments")
		return
	}
	if output == "" {
		fmt.Fprint(os.Stdout, strings.Join(args, " "))
	} else {
		file, err := os.Create(output)
		if err != nil {
        	fmt.Println("Error:", err)
        	return
    	}
	    defer file.Close()
		file.WriteString(strings.Join(args, " ") + "\n")
	}
}

func type_(args []string, output string) {
	if len(args) != 1 {
		fmt.Println("Error: expected exactly one argument")
		return
	}
	command := args[0]
	var outputText string
	if _, exists := COMMANDS[command]; exists {
		outputText = args[0] + " is a shell builtin"
	} else {
		fullPath := findExecutable(command, paths)
		if fullPath != "" {
			outputText = command + " is " + fullPath
		} else {
			outputText = args[0] + ": not found"
		}
	}
	if output == "" {
		fmt.Fprintln(os.Stdout, outputText)
	} else {
		file, err := os.Create(output)
		if err != nil {
        	fmt.Println("Error:", err)
        	return
    	}
	    defer file.Close()
		file.WriteString(outputText + "\n")
	}
}

func execute(command string, args []string, output string) bool {
	fullPath := findExecutable(command, paths)
	if fullPath == "" {
		return false
	}

	var outFile *os.File
	var err error

	if output != "" {
		outFile, err = os.Create(output)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error creating output file:", err)
			return false
		}
		defer outFile.Close()
	}

	cmd := &exec.Cmd{
		Path:   fullPath,
		Args:   append([]string{command}, args...),
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	if output != "" {
		cmd.Stdout = outFile
	}

	if err := cmd.Run(); err != nil {
		fmt.Fprintln(os.Stdout, "Error:", err)
	}

	return true
}
