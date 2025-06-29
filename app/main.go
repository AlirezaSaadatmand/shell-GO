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

type Redirection struct {
	File   string
	Append bool
}

func parseRedirections(args []string) (cleanArgs []string, stdoutRedir, stderrRedir *Redirection) {
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case ">", "1>":
			if i+1 < len(args) {
				stdoutRedir = &Redirection{File: args[i+1], Append: false}
				i++
			}
		case ">>", "1>>":
			if i+1 < len(args) {
				stdoutRedir = &Redirection{File: args[i+1], Append: true}
				i++
			}
		case "2>":
			if i+1 < len(args) {
				stderrRedir = &Redirection{File: args[i+1], Append: false}
				i++
			}
		case "2>>":
			if i+1 < len(args) {
				stderrRedir = &Redirection{File: args[i+1], Append: true}
				i++
			}
		default:
			cleanArgs = append(cleanArgs, args[i])
		}
	}
	return
}

type Output struct {
	Stdout *os.File
	Stderr *os.File
}

func setupOutput(stdoutRedir, stderrRedir *Redirection) (*Output, error) {
	out := &Output{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	if stdoutRedir != nil {
		flag := os.O_CREATE | os.O_WRONLY
		if stdoutRedir.Append {
			flag |= os.O_APPEND
		} else {
			flag |= os.O_TRUNC
		}
		f, err := os.OpenFile(stdoutRedir.File, flag, 0644)
		if err != nil {
			return nil, err
		}
		out.Stdout = f
	}

	if stderrRedir != nil {
		flag := os.O_CREATE | os.O_WRONLY
		if stderrRedir.Append {
			flag |= os.O_APPEND
		} else {
			flag |= os.O_TRUNC
		}
		f, err := os.OpenFile(stderrRedir.File, flag, 0644)
		if err != nil {
			if out.Stdout != os.Stdout {
				out.Stdout.Close()
			}
			return nil, err
		}
		out.Stderr = f
	}

	return out, nil
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

var COMMANDS map[string]func([]string, *Output)
var builtin []string
var paths = strings.Split(os.Getenv("PATH"), ":")

func init() {
	COMMANDS = map[string]func([]string, *Output){
		"exit": exit,
		"echo": echo,
		"type": type_,
		"pwd": pwd,
		"cd": cd,
	}
	builtin = []string{
		"exit",
		"echo",
		"type",
		"pwd",
		"cd",	
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
		command, args := separateCommandArgs(strings.TrimSpace(input))
		args, stdoutRedir, stderrRedir := parseRedirections(args)

		out, err := setupOutput(stdoutRedir, stderrRedir)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Redirection error:", err)
			continue
		}

		if builtinFunc, ok := COMMANDS[command]; ok {
			builtinFunc(args, out)
		} else {
			if !execute(command, args, out) {
				fmt.Fprintln(os.Stderr, command+": command not found")
			}
		}
	}
}

func exit(args []string, out *Output) {
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

func echo(args []string, out *Output) {
	_, err := fmt.Fprintln(out.Stdout, strings.Join(args, " "))
	if err != nil {
		fmt.Fprintln(out.Stderr, "Error writing to stdout:", err)
	}
}

func type_(args []string, out *Output) {

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
	_, err := fmt.Fprintln(out.Stdout, outputText)
	if err != nil {
		fmt.Fprintln(out.Stderr, "Error writing to stdout:", err)
	}
}

func pwd(args []string, out *Output) {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(out.Stderr, "Error writing to stdout:", err)
	} else {
		fmt.Fprintln(out.Stdout, dir)
	}
}

var lastDir string

func cd(args []string, out *Output) {
	var targetDir string

	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(out.Stderr, "cd: failed to get current directory:", err)
		return
	}

	if len(args) == 0 {
		home := os.Getenv("HOME")
		if home == "" {
			fmt.Fprintln(out.Stderr, "cd: HOME not set")
			return
		}
		targetDir = home
	} else if args[0] == "-" {
		if lastDir == "" {
			fmt.Fprintln(out.Stderr, "cd: OLDPWD not set")
			return
		}
		targetDir = lastDir
		fmt.Fprintln(out.Stdout, targetDir)
	} else if strings.HasPrefix(args[0], "~") {
		home := os.Getenv("HOME")
		if home == "" {
			fmt.Fprintln(out.Stderr, "cd: HOME not set")
			return
		}
		targetDir = filepath.Join(home, strings.TrimPrefix(args[0], "~"))
	} else {
		targetDir = args[0]
	}

	if err := os.Chdir(targetDir); err != nil {
		fmt.Fprintln(out.Stderr, "cd:", err.(*os.PathError).Err)
		return
	}

	lastDir = currentDir
}

func execute(command string, args []string, out *Output) bool {
	fullPath := findExecutable(command, paths)
	if fullPath == "" {
		return false
	}

	cmd := &exec.Cmd{
		Path:   fullPath,
		Args:   append([]string{command}, args...),
		Stdin:  os.Stdin,
		Stdout: out.Stdout,
		Stderr: out.Stderr,
	}

	cmd.Run()

	if out.Stdout != os.Stdout {
		out.Stdout.Close()
	}
	if out.Stderr != os.Stderr {
		out.Stderr.Close()
	}

	return true
}
