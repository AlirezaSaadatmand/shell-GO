# shell-GO

shell-GO is a simple Unix-like shell implemented in Go. It supports pipelines, built-in commands, input/output redirection, command history, and file-based history management.

## Features

✅ Execute external commands  
✅ Support for pipelines (`|`)  
✅ Input and output redirection (`>`, `>>`, `2>`, `2>>`)  
✅ Built-in commands:
- `cd`
- `echo`
- `exit`
- `pwd`
- `type`
- `history`

✅ Autocompletion (tab)  
✅ Command history in memory  
✅ Load/Write/Append command history from/to file  
✅ Automatically load from `$HISTFILE` on startup and save to it on exit  
✅ Proper handling of built-ins inside pipelines  

---

## Installation

Make sure you have Go installed (1.18+ recommended).

```bash
git clone https://github.com/AlirezaSaadatmand/shell-GO.git
cd shell-GO/app
go build main.go
```

Usage
-----
Run the shell:

    go run main.go

You can then type shell commands interactively, like:

    $ echo hello
    hello

    $ ls | grep go | wc


Environment Variables
---------------------
PATH:
    Used to locate executables.

HISTFILE:
    If set, GoShell will:
    - Load history from the file on startup.
    - Append new commands to the file on exit.


Built-in Commands
-----------------

### cd [dir]
    Change the current directory. Use "cd -" to switch to the last visited directory.

### pwd
    Print the current working directory.

### echo [args...]
    Print arguments to stdout.

### exit [status]
    Exit the shell with an optional status code. Saves history to file if HISTFILE is set.

### type <command>
    Tell whether a command is a shell built-in or an external executable.

### history
    Print command history with line numbers.

    Variants:
      history <n>         Show last n entries
      history -r <file>   Read history from file
      history -w <file>   Write current history to file (overwrite)
      history -a <file>   Append new history entries to file


Command Line Editing
--------------------
- Use arrow keys for history navigation.
- Press Tab for autocompletion.
    - Supports partial matches and common prefixes.
    - Pressing Tab twice shows all possible completions.


Redirection
-----------
\> or 1>      Redirect stdout (overwrite)

\>> or 1>>    Redirect stdout (append)

2>           Redirect stderr (overwrite)

2>>          Redirect stderr (append)


Examples
--------
    $ echo hello > out.txt
    $ cat out.txt | grep h
    hello

    $ ls -la / | grep bin | wc -l


Known Limitations
-----------------
- No support for job control (fg, bg, etc.)
- No support for command substitution ($(...))
- No background jobs (&)
- No environment variable setting/exporting
`