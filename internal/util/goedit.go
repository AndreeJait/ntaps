package util

import (
	"regexp"
	"strings"
)

// InsertImport ensures importLine is in the file's import block or creates one.
func InsertImport(src, importLine string) string {
	if strings.Contains(src, importLine) {
		return src
	}

	if strings.Contains(src, "import (") {
		return strings.Replace(src, "import (", "import (\n\t"+importLine+"\n", 1)
	}

	singleRe := regexp.MustCompile(`import\s+"[^"]+"`)
	if singleRe.MatchString(src) {
		old := singleRe.FindString(src)
		newBlock := "import (\n\t" + importLine + "\n\t" + strings.TrimPrefix(old, "import ") + "\n)"
		return strings.Replace(src, old, newBlock, 1)
	}

	idx := strings.Index(src, "\n")
	if idx == -1 {
		return src + "\nimport (\n\t" + importLine + "\n)\n"
	}
	return src[:idx+1] + "import (\n\t" + importLine + "\n)\n" + src[idx+1:]
}
