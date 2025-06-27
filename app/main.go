package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var COMMANDS map[string]func([]string)
var builtin []string

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
		args := strings.Split(input, " ")[1:]

		if _, ok := COMMANDS[command]; ok {
			COMMANDS[command](args)
		} else {
			fmt.Println(command + ": command not found")
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
		pathEnv := os.Getenv("PATH")
		paths := strings.Split(pathEnv, ":")

		for _, dir := range paths {
			fullPath := filepath.Join(dir, command)
			if info, err := os.Stat(fullPath); err == nil && info.Mode().IsRegular() && info.Mode().Perm()&0111 != 0 {
				fmt.Println(command + " is " + fullPath)
				return
			}
		}
		fmt.Println(args[0] + ": not found")
	}
}
