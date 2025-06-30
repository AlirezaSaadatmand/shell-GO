package completer

import (
	"fmt"
	"os"
	"shell-GO/app/config"
	"sort"
	"strings"
	"unicode"
)

type AutoCompleter struct {
	lastLine string
	lastPos  int
	tabCount int
}

func (a AutoCompleter) Do(line []rune, pos int) ([][]rune, int) {
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

func findCommandMatches(prefix string) []string {
	var matches []string
	seen := make(map[string]bool)

	for _, b := range config.Builtin {
		if strings.HasPrefix(b, prefix) && !seen[b] {
			seen[b] = true
			matches = append(matches, b)
		}
	}

	for _, dir := range config.Paths {
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
