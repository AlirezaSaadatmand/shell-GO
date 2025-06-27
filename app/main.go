package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var COMMANDS map[string]func([]string)

func init() {
	COMMANDS = map[string]func([]string){
		"exit": exit,
		"echo": echo,
		"type": type_,
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
		input = input[:len(input) - 1]
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
	if len(args) > 1 {
		fmt.Println("Error: too many arguments")
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
		fmt.Println("Error: not enough arguments")
		return 
	}
	fmt.Println(strings.Join(args, " "))
}

func type_(args []string) {
	if len(args) != 1 {
		fmt.Println("Error: expected exactly one argument")
		return
	}

	if _, exists := COMMANDS[args[0]]; exists {
		fmt.Println(args[0] + " is a shell builtin")
	} else {
		fmt.Println(args[0] + ": not found")
	}
}