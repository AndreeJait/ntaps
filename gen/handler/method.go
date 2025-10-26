package handler

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AndreeJait/ntaps/internal/paths"
	"github.com/AndreeJait/ntaps/internal/util"
)

func ensureMethodAndRoute(
	pkg string,
	ucPkg string,
	endpointType string,
	endpoint string,
	handlerMethod string,
	ucMethodName string,
	withParamUc bool,
	withResponseUc bool,
	tag string,
	verb string,
) error {
	mod := util.ModulePathGuess()
	path := filepath.Join(paths.HandlerRootHTTPDir, pkg, paths.HandlerPkgFileName)

	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	src := string(raw)

	// imports needed
	requiredImports := []string{
		fmt.Sprintf(`"%s/internal/usecase/%s"`, mod, ucPkg),
		`"github.com/AndreeJait/go-utility/response"`,
		`"github.com/AndreeJait/go-utility/tracer"`,
		`"github.com/labstack/echo/v4"`,
	}
	for _, imp := range requiredImports {
		src = util.InsertImport(src, imp)
	}

	// ensure route registration in Handle()
	groupName := map[string]string{
		"public":   "groupPublic",
		"internal": "groupInternal",
		"private":  "groupPrivate",
	}[strings.ToLower(endpointType)]
	if groupName == "" {
		groupName = "groupPublic"
	}
	verbUpper := strings.ToUpper(verb)

	routeLine := fmt.Sprintf(`%s.%s("%s", h.%s)`, groupName, verbUpper, endpoint, handlerMethod)
	routeLineTabbed := "\t" + routeLine

	if !strings.Contains(src, "func (h *handler) Handle()") {
		return fmt.Errorf("Handle() not found in %s", path)
	}
	if !strings.Contains(src, routeLine) && !strings.Contains(src, routeLineTabbed) {
		marker := "// ntaps:routes"
		if idx := strings.Index(src, marker); idx != -1 {
			src = src[:idx] + routeLine + "\n\t" + src[idx:]
		} else {
			last := strings.LastIndex(src, "}")
			src = src[:last] + "\n" + routeLine + "\n" + src[last:]
		}
	}

	// ensure method body exists
	methodSig := fmt.Sprintf("func (h *handler) %s(", handlerMethod)
	if !strings.Contains(src, methodSig) {
		methodCode := buildHandlerMethod(
			handlerMethod,
			ucPkg,
			ucMethodName,
			endpointType,
			verbUpper,
			endpoint,
			withParamUc,
			withResponseUc,
			tag,
		)
		src += methodCode
	}

	// NEW FEATURE:
	// If this is a GET withParamUc and the path has params, update the <UcMethodName>Request DTO
	if strings.EqualFold(verbUpper, "GET") && withParamUc {
		_, pathParams := normalizePathParams(endpoint)
		if len(pathParams) > 0 {
			if err := ensureRequestDTOHasPathParams(
				ucPkg,
				ucMethodName,
				pathParams,
			); err != nil {
				// non-fatal: we still write handler; just surface the error
				fmt.Println("ℹ️  warning: could not enrich DTO with path params:", err)
			}
		}
	}

	return util.WriteGoFile(path, src)
}

func updateInfraHandlerInit(pkg string) error {
	mod := util.ModulePathGuess()
	path := paths.HandlerInfraInitPath

	raw, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("cannot find %s to register handler", path)
	}
	src := string(raw)

	// 1. ensure import "<module>/internal/adapters/inbound/http/<pkg>"
	importLine := fmt.Sprintf(`"%s/internal/adapters/inbound/http/%s"`, mod, pkg)
	if !strings.Contains(src, importLine) {
		src = util.InsertImport(src, importLine)
	}

	// 2. find handlers slice block: var handlers = []http.Handler{ ... }
	openMarker := "var handlers = []http.Handler{"
	start := strings.Index(src, openMarker)
	if start == -1 {
		return fmt.Errorf("handlers slice not found in %s", path)
	}

	after := src[start+len(openMarker):] // content after "{"
	endRel := strings.Index(after, "}")
	if endRel == -1 {
		return fmt.Errorf("handlers slice closing '}' not found in %s", path)
	}

	blockStart := start + len(openMarker)
	blockEnd := blockStart + endRel // position of '}'
	block := src[blockStart:blockEnd]

	// 3. build canonical ctor prefix
	ctorPrefix := fmt.Sprintf(`%s.New%sHandler(`, pkg, util.ToPascalCase(pkg))

	// If any line already calls <pkg>.New<Pkg>Handler(...), skip adding
	if strings.Contains(block, ctorPrefix) {
		return util.WriteGoFile(path, src)
	}

	// 4. otherwise insert default call before the end of slice
	newCall := fmt.Sprintf("\n\t\t%s.New%sHandler(s.cfg, groupV1, s.uc),",
		pkg,
		util.ToPascalCase(pkg),
	)

	src = src[:blockEnd] + newCall + src[blockEnd:]

	return util.WriteGoFile(path, src)
}
