package usecase

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AndreeJait/ntaps/internal/util"
)

func createPort(dir, pkg, method string, withParam, withResp bool) error {
	path := filepath.Join(dir, "port.go")
	return util.WriteGoFile(path, renderPort(pkg, method, withParam, withResp))
}

func ensurePort(dir, pkg, method string, withParam, withResp bool) error {
	path := filepath.Join(dir, "port.go")

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return createPort(dir, pkg, method, withParam, withResp)
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	src := string(b)

	req := "ctx context.Context"
	if withParam {
		req += ", req " + method + "Request"
	}
	ret := "error"
	if withResp {
		ret = "(" + method + "Response, error)"
	}
	newLine := fmt.Sprintf("\t%s(%s) %s\n", method, req, ret)

	// ensure package line
	if !strings.Contains(src, "package "+pkg) {
		src = "package " + pkg + "\n\n" + src
	}

	// ensure import "context"
	if !strings.Contains(src, "import \"context\"") && !strings.Contains(src, "import (") {
		firstNL := strings.Index(src, "\n")
		src = src[:firstNL+1] + "import \"context\"\n" + src[firstNL+1:]
	}

	// ensure interface block + method
	start := strings.Index(src, "type UseCase interface {")
	if start == -1 {
		src = strings.TrimRight(src, "\n") +
			"\n\ntype UseCase interface {\n" +
			newLine +
			"}\n"
		return util.WriteGoFile(path, src)
	}

	end := strings.Index(src[start:], "}")
	if end == -1 {
		src += "\n}\n"
		end = strings.LastIndex(src, "}")
	}

	block := src[start : start+end]
	if strings.Contains(block, method+"(") {
		return util.WriteGoFile(path, src)
	}

	insertPos := start + end
	src = src[:insertPos] + "\n" + newLine + src[insertPos:]
	return util.WriteGoFile(path, src)
}

func renderPort(pkg, method string, withParam, withResp bool) string {
	req := "ctx context.Context"
	if withParam {
		req += ", req " + method + "Request"
	}
	ret := "error"
	if withResp {
		ret = "(" + method + "Response, error)"
	}
	return fmt.Sprintf(`package %s

import "context"

type UseCase interface {
	%s(%s) %s
}
`, pkg, method, req, ret)
}
