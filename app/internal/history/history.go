package history

import (
	"bufio"
	"fmt"
	"os"
	"shell-GO/app/config"
	"strings"
)

func WriteHistoryToFile() {
	if config.HistFile == "" {
		return
	}
	file, err := os.Create(config.HistFile)
	if err != nil {
		return
	}
	defer file.Close()

	for _, entry := range config.ShellHistory {
		fmt.Fprintln(file, entry)
	}
}

func LoadHistoryFromFile(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) != "" {
			config.ShellHistory = append(config.ShellHistory, line)
		}
	}
	config.HistoryAppendIndex = len(config.ShellHistory)
}
