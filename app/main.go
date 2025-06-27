package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Fprint

func main() {
	for {
		// Uncomment this block to pass the first stage
		fmt.Fprint(os.Stdout, "$ ")
	
		// Wait for user input
		command, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading input:", err)
			os.Exit(1)
		}
	
		if command[:len(command) - 1] == "exit 0" {
			os.Exit(0)
		} else if strings.Split(command[:len(command) - 1], " ")[0] == "echo" {
			fmt.Println(strings.Join(strings.Split(command[:len(command) - 1], " ")[1:], " "))
		} else {
			fmt.Println(command[:len(command) - 1] + ": command not found")
		}
	}
}
