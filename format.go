package main

import (
	"go/format"
	"os"
	"path/filepath"

	"golang.org/x/tools/imports"
)

// writeGoFile formats + fixes imports, then writes.
// Falls back to raw write if formatting fails (so the tool still works).
func writeGoFile(path string, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	// Prefer goimports (formats AND fixes imports)
	//local := modulePathGuess() // keep your module grouped at bottom of imports
	formatted, err := imports.Process(path, []byte(content), &imports.Options{
		Comments:   true,
		FormatOnly: false,
		TabIndent:  true,
		TabWidth:   8,
	})
	if err != nil {
		// Fallback to go/format if goimports can’t parse (incomplete code, etc.)
		if f2, err2 := format.Source([]byte(content)); err2 == nil {
			formatted = f2
		} else {
			// last resort: write the raw content so the tool never blocks you
			formatted = []byte(content)
		}
	}

	return os.WriteFile(path, formatted, 0o644)
}
