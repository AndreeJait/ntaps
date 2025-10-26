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

	importLine := fmt.Sprintf(`"%s/internal/adapters/inbound/http/%s"`, mod, pkg)
	if !strings.Contains(src, importLine) {
		src = util.InsertImport(src, importLine)
	}

	newItem := fmt.Sprintf("\t\t%s.New%sHandler(s.cfg, groupV1, s.uc),", pkg, util.ToPascalCase(pkg))
	open := "var handlers = []http.Handler{"
	start := strings.Index(src, open)
	if start == -1 {
		return fmt.Errorf("handlers slice not found in %s", path)
	}
	after := src[start:]
	end := strings.Index(after, "}")
	if end == -1 {
		return fmt.Errorf("handlers slice closing '}' not found in %s", path)
	}
	block := after[:end]
	if !strings.Contains(block, newItem) {
		insertAt := start + end
		src = src[:insertAt] + "\n" + newItem + "\n" + src[insertAt:]
	}

	return util.WriteGoFile(path, src)
}
