package util

import (
	"os"
	"strings"
)

// ModulePathGuess tries to read first line of go.mod.
func ModulePathGuess() string {
	b, err := os.ReadFile("go.mod")
	if err != nil {
		return "go-template-hexagonal"
	}
	lines := strings.Split(string(b), "\n")
	if len(lines) > 0 && strings.HasPrefix(strings.TrimSpace(lines[0]), "module ") {
		return strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(lines[0]), "module"))
	}
	return "go-template-hexagonal"
}
