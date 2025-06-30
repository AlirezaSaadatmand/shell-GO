package parser

import (
	"os"
	"shell-GO/app/config"
	"strings"
	"unicode"
)

func SeparateCommandArgs(input string) (string, []string) {
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

func ParseRedirections(args []string) (cleanArgs []string, stdoutRedir, stderrRedir *config.Redirection) {

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case ">", "1>":
			if i+1 < len(args) {
				stdoutRedir = &config.Redirection{File: args[i+1], Append: false}
				i++
			}
		case ">>", "1>>":
			if i+1 < len(args) {
				stdoutRedir = &config.Redirection{File: args[i+1], Append: true}
				i++
			}
		case "2>":
			if i+1 < len(args) {
				stderrRedir = &config.Redirection{File: args[i+1], Append: false}
				i++
			}
		case "2>>":
			if i+1 < len(args) {
				stderrRedir = &config.Redirection{File: args[i+1], Append: true}
				i++
			}
		default:
			cleanArgs = append(cleanArgs, args[i])
		}
	}
	return
}

func SetupOutput(stdoutRedir, stderrRedir *config.Redirection) (*config.Output, error) {
	out := &config.Output{
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
