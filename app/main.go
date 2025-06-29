package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/chzyer/readline"
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

func completeCommands(prefix string) [][]rune {
	var result [][]rune
	seen := make(map[string]bool)

	for _, b := range builtin {
		if strings.HasPrefix(b, prefix) && !seen[b] {
			seen[b] = true
			result = append(result, []rune(b))
		}
	}

	for _, dir := range paths {
		files, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, f := range files {
			name := f.Name()
			if strings.HasPrefix(name, prefix) && !seen[name] {
				seen[name] = true
				result = append(result, []rune(name))
			}
		}
	}
	return result
}

func findCommandMatches(prefix string) []string {
	var matches []string
	seen := make(map[string]bool)

	for _, b := range builtin {
		if strings.HasPrefix(b, prefix) && !seen[b] {
			seen[b] = true
			matches = append(matches, b)
		}
	}

	for _, dir := range paths {
		files, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, f := range files {
			name := f.Name()
			if strings.HasPrefix(name, prefix) && !seen[name] {
				seen[name] = true
				matches = append(matches, name)
			}
		}
	}

	sort.Strings(matches)
	return matches
}


func longestCommonPrefix(strs []string) string {
	if len(strs) == 0 {
		return ""
	}
	prefix := strs[0]
	for _, s := range strs[1:] {
		for !strings.HasPrefix(s, prefix) {
			prefix = prefix[:len(prefix)-1]
			if prefix == "" {
				return ""
			}
		}
	}
	return prefix
}

type AutoCompleter struct {
	lastLine  string
	lastPos   int
	tabCount  int
}

func (a *AutoCompleter) Do(line []rune, pos int) ([][]rune, int) {
	start := pos
	for start > 0 && !unicode.IsSpace(line[start-1]) {
		start--
	}
	current := string(line[start:pos])

	if current != a.lastLine || pos != a.lastPos {
		a.lastLine = current
		a.lastPos = pos
		a.tabCount = 0
	}

	matches := findCommandMatches(current)
	if len(matches) == 0 {
		fmt.Fprint(os.Stderr, "\a")
		a.tabCount = 0
		return nil, pos
	}

	if len(matches) == 1 {
		match := matches[0]
		suffix := match[len(current):] + " "
		a.tabCount = 0
		return [][]rune{[]rune(suffix)}, pos
	}

	lcp := longestCommonPrefix(matches)
	if lcp == current {
		a.tabCount++
		if a.tabCount == 1 {
			fmt.Fprint(os.Stderr, "\a")
		} else {
			fmt.Println()
			for _, m := range matches {
				fmt.Print(m + "  ")
			}
			fmt.Println()
			fmt.Print("$ " + current)
			a.tabCount = 0
		}
		return nil, pos
	}

	suffix := lcp[len(current):]
	a.lastLine = lcp
	a.tabCount = 0
	return [][]rune{[]rune(suffix)}, pos
}

func executePipeline(line string) bool {
	parts := strings.Split(line, "|")
	commands := make([][]string, 0, len(parts))

	// parse each part into command and args + redirections
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		cmd, args := separateCommandArgs(part)
		args, _, _ = parseRedirections(args)
		commands = append(commands, append([]string{cmd}, args...))
		// you can handle redirections later or inside execute function
	}

	// Prepare pipes
	// We'll create n-1 pipes for n commands
	cmds := make([]*exec.Cmd, len(commands))
	var pipes []struct{ r, w *os.File }

	for i := 0; i < len(commands)-1; i++ {
		r, w, err := os.Pipe()
		if err != nil {
			fmt.Fprintln(os.Stderr, "pipe error:", err)
			return false
		}
		pipes = append(pipes, struct{ r, w *os.File }{r, w})
	}

	// Setup each command's stdin/stdout
	for i, cmdArgs := range commands {
		cmdName := cmdArgs[0]
		cmdArgsSlice := cmdArgs[1:]

		fullPath := findExecutable(cmdName, paths)
		if fullPath == "" {
			fmt.Fprintln(os.Stderr, cmdName+": command not found")
			return false
		}

		cmd := exec.Command(fullPath, cmdArgsSlice...)

		// stdin
		if i == 0 {
			cmd.Stdin = os.Stdin
		} else {
			cmd.Stdin = pipes[i-1].r
		}

		// stdout
		if i == len(commands)-1 {
			cmd.Stdout = os.Stdout
		} else {
			cmd.Stdout = pipes[i].w
		}

		// stderr always to os.Stderr for simplicity, you can handle redirection if you want
		cmd.Stderr = os.Stderr

		cmds[i] = cmd
	}

	// Start all commands
	for i, cmd := range cmds {
		if err := cmd.Start(); err != nil {
			fmt.Fprintln(os.Stderr, "failed to start command:", err)
			return false
		}
		// Close pipe ends in parent as needed
		if i > 0 {
			pipes[i-1].r.Close()
		}
		if i < len(cmds)-1 {
			pipes[i].w.Close()
		}
	}

	// Wait all commands
	for _, cmd := range cmds {
		cmd.Wait()
	}

	return true
}


var COMMANDS map[string]func([]string, *Output)
var builtin []string
var paths = strings.Split(os.Getenv("PATH"), ":")

func init() {
	COMMANDS = map[string]func([]string, *Output){
		"exit": exit,
		"echo": echo,
		"type": type_,
		"pwd":  pwd,
		"cd":   cd,
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
	rl, err := readline.NewEx(&readline.Config{
		Prompt:          "$ ",
		HistoryFile:     "/tmp/readline.tmp",
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
		AutoComplete:    &AutoCompleter{},
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
		if strings.TrimSpace(line) == "" {
			continue
		}
		if strings.Contains(line, "|") {
			ok := executePipeline(line)
			if !ok {
				fmt.Fprintln(os.Stderr, "pipeline execution failed")
			}
		} else {
			command, args := separateCommandArgs(line)
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
}

func exit(args []string, out *Output) {
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
		if os.IsNotExist(err) {
			fmt.Fprintf(out.Stderr, "cd: %s: No such file or directory\n", targetDir)
		} else {
			fmt.Fprintf(out.Stderr, "cd: %s: %s\n", targetDir, err.Error())
		}
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
