package config

import (
	"os"
	"strings"
)

type Output struct {
	Stdout *os.File
	Stderr *os.File
}

type Redirection struct {
	File   string
	Append bool
}

var Paths = strings.Split(os.Getenv("PATH"), ":")

var Commands map[string]func([]string, *Output)
var Builtin []string

var ShellHistory []string
var HistoryAppendIndex int

var HistFile = os.Getenv("HISTFILE")
