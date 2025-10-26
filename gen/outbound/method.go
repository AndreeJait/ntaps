package outbound

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AndreeJait/ntaps/internal/paths"
	"github.com/AndreeJait/ntaps/internal/util"
)

func ensureOutboundPortHasMethod(pkg, method string, withParam, withResp bool) error {
	iface := ifaceName(pkg)
	path := filepath.Join(paths.OutboundRootPath, pkg, "port.go")

	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	src := string(b)

	// add context import
	if !strings.Contains(src, "import (") && !strings.Contains(src, `import "context"`) {
		nl := strings.Index(src, "\n")
		src = src[:nl+1] + `import "context"` + "\n" + src[nl+1:]
	} else if strings.Contains(src, "import (") && !strings.Contains(src, `"context"`) {
		src = util.InsertImport(src, `"context"`)
	}

	// method signature
	req := "ctx context.Context"
	if withParam {
		req += ", req " + method + "Request"
	}
	ret := "error"
	if withResp {
		ret = "(" + method + "Response, error)"
	}
	newLine := fmt.Sprintf("\t%s(%s) %s\n", method, req, ret)

	start := strings.Index(src, "type "+iface+" interface {")
	if start == -1 {
		src = strings.TrimRight(src, "\n") + "\n" +
			fmt.Sprintf("type %s interface {\n%s}\n", iface, newLine)
		return util.WriteGoFile(path, src)
	}

	end := strings.Index(src[start:], "}")
	if end == -1 {
		src += "\n}\n"
		end = strings.LastIndex(src, "}")
	}
	block := src[start : start+end]
	if !strings.Contains(block, method+"(") {
		insertAt := start + end
		src = src[:insertAt] + "\n" + newLine + src[insertAt:]
	}

	return util.WriteGoFile(path, src)
}

func ensureOutboundImplHasMethod(pkg, method string, withParam, withResp bool) error {
	mod := util.ModulePathGuess()
	path := filepath.Join(paths.OutboundRootPath, pkg, "impl.go")

	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	src := string(raw)

	// ensure imports
	required := []string{
		`"context"`,
		`"github.com/AndreeJait/go-utility/tracer"`,
		fmt.Sprintf(`"%s/internal/infrastructure/config"`, mod),
		`"github.com/AndreeJait/go-utility/loggerw"`,
	}
	for _, imp := range required {
		src = util.InsertImport(src, imp)
	}

	sig := fmt.Sprintf("func (i *impl) %s(", method)
	if strings.Contains(src, sig) {
		return util.WriteGoFile(path, src)
	}

	args := "ctx context.Context"
	if withParam {
		args += ", req " + method + "Request"
	}
	ret := "error"
	retBody := "// TODO: implement\n\treturn nil"
	if withResp {
		ret = "(" + method + "Response, error)"
		retBody = "var resp " + method + "Response\n\t// TODO: implement\n\treturn resp, nil"
	}

	methodSrc := fmt.Sprintf(`

func (i *impl) %s(%s) %s {
	span, ctx := tracer.StartSpan(ctx, tracer.GetFuncName(i.%s))
	defer span.End()

	%s
}
`, method, args, ret, method, retBody)

	src += methodSrc
	return util.WriteGoFile(path, src)
}
