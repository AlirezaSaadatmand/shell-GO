package executor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"shell-GO/app/config"
	"shell-GO/app/internal/parser"
	"strings"
	"sync"
)

func FindExecutable(command string, paths []string) string {
	for _, dir := range paths {
		fullPath := filepath.Join(dir, command)
		if info, err := os.Stat(fullPath); err == nil && info.Mode().IsRegular() && info.Mode().Perm()&0111 != 0 {
			return fullPath
		}
	}
	return ""
}

func ExecutePipeline(line string) bool {
	parts := strings.Split(line, "|")
	commands := make([][]string, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		cmd, args := parser.SeparateCommandArgs(part)
		args, _, _ = parser.ParseRedirections(args)
		commands = append(commands, append([]string{cmd}, args...))
	}

	numCmds := len(commands)
	var pipes []struct{ r, w *os.File }

	for i := 0; i < numCmds-1; i++ {
		r, w, err := os.Pipe()
		if err != nil {
			fmt.Fprintln(os.Stderr, "pipe error:", err)
			return false
		}
		pipes = append(pipes, struct{ r, w *os.File }{r, w})
	}

	var processes []*exec.Cmd
	var wg sync.WaitGroup

	for i, cmdArgs := range commands {
		cmdName := cmdArgs[0]
		args := cmdArgs[1:]

		var stdin *os.File = os.Stdin
		var stdout *os.File = os.Stdout

		if i > 0 {
			stdin = pipes[i-1].r
		}
		if i < numCmds-1 {
			stdout = pipes[i].w
		}

		if _, isBuiltin := config.Commands[cmdName]; isBuiltin {
			// Built-in: run in goroutine
			r := stdin
			w := stdout

			wg.Add(1)
			go func(name string, args []string, in, out *os.File) {
				defer wg.Done()
				RunBuiltin(name, args, &config.Output{
					Stdout: out,
					Stderr: os.Stderr,
				}, in)
				if in != os.Stdin {
					in.Close()
				}
				if out != os.Stdout {
					out.Close()
				}
			}(cmdName, args, r, w)

		} else {
			fullPath := FindExecutable(cmdName, config.Paths)
			if fullPath == "" {
				fmt.Fprintln(os.Stderr, cmdName+": command not found")
				return false
			}

			cmd := exec.Command(fullPath, args...)
			cmd.Stdin = stdin
			cmd.Stdout = stdout
			cmd.Stderr = os.Stderr

			if err := cmd.Start(); err != nil {
				fmt.Fprintln(os.Stderr, "start error:", err)
				return false
			}
			processes = append(processes, cmd)
		}
	}

	// Close all pipe ends in the parent process
	for _, pipe := range pipes {
		pipe.r.Close()
		pipe.w.Close()
	}

	// Wait for external commands
	for _, cmd := range processes {
		cmd.Wait()
	}

	// Wait for builtin goroutines
	wg.Wait()

	return true
}

func RunBuiltin(cmd string, args []string, out *config.Output, in *os.File) {
	// Temporarily override stdin/stdout/stderr
	oldStdin := os.Stdin
	oldStdout := os.Stdout
	oldStderr := os.Stderr

	if in != nil {
		os.Stdin = in
	}
	if out.Stdout != nil {
		os.Stdout = out.Stdout
	}
	if out.Stderr != nil {
		os.Stderr = out.Stderr
	}

	defer func() {
		os.Stdin = oldStdin
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	if builtinFunc, ok := config.Commands[cmd]; ok {
		builtinFunc(args, out)
	}
}

func Execute(command string, args []string, out *config.Output) bool {
	fullPath := FindExecutable(command, config.Paths)
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
