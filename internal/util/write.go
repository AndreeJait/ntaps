package util

import (
	"go/format"
	"os"
	"path/filepath"

	"golang.org/x/tools/imports"
)

// WriteGoFile formats + fixes imports, then writes.
// Falls back to raw write on parse failure so we never block scaffolding.
func WriteGoFile(path string, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	formatted, err := imports.Process(path, []byte(content), &imports.Options{
		Comments:   true,
		FormatOnly: false,
		TabIndent:  true,
		TabWidth:   8,
	})
	if err != nil {
		if f2, err2 := format.Source([]byte(content)); err2 == nil {
			formatted = f2
		} else {
			formatted = []byte(content)
		}
	}

	return os.WriteFile(path, formatted, 0o644)
}
