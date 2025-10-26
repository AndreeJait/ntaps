package cmd

import (
	"fmt"
	"os"
	"strings"
)

func exitErr(msg string) {
	fmt.Fprintln(os.Stderr, "❌", msg)
	os.Exit(1)
}

func isPascalCase(s string) bool {
	if s == "" {
		return false
	}
	return strings.ToUpper(s[:1]) == s[:1]
}
