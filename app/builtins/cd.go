package builtins

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"shell-GO/app/config"
)

var lastDir string

func Cd(args []string, out *config.Output) {
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
