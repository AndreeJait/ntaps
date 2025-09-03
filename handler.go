package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

/* ============================ create-handler ============================ */

func runCreateHandler(pkg, ucPkg, endpointType, endpoint string, withParamUc, withResponseUc bool, ucMethodName, handlerMethod, tag, verb string) error {
	// 1) ensure usecase/method exists
	if err := runCreateUsecase(ucPkg, ucMethodName, withParamUc, withResponseUc); err != nil {
		return err
	}

	// 2) ensure handler package skeleton
	if err := ensureHandlerPkg(pkg); err != nil {
		return err
	}

	// 3) ensure method + route
	if err := ensureHandlerMethodAndRoute(pkg, ucPkg, endpointType, endpoint, handlerMethod, ucMethodName, withParamUc, withResponseUc, tag, verb); err != nil {
		return err
	}

	// 4) register in infrastructure di/handler.go
	if err := updateInfraHandlerInit(pkg); err != nil {
		return err
	}
	return nil
}

/* ---------- handler package skeleton ---------- */

func ensureHandlerPkg(pkg string) error {
	mod := modulePathGuess()
	dir := filepath.Join(handlerRootHTTPDir, pkg)
	if _, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

	path := filepath.Join(dir, handlerPkgFileName)
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		body := fmt.Sprintf(`package %s

import (
	http "%[2]s/internal/adapters/inbound/http"
	"%[2]s/internal/adapters/inbound/http/common/middleware"
	"%[2]s/internal/infrastructure/config"
	"%[2]s/internal/usecase"

	"github.com/AndreeJait/go-utility/response"
	"github.com/AndreeJait/go-utility/tracer"
	"github.com/labstack/echo/v4"
)

type handler struct {
	route *echo.Group
	uc    *usecase.UseCase
	cfg   *config.Config
}

func New%[3]sHandler(cfg *config.Config, route *echo.Group, uc *usecase.UseCase) http.Handler {
	return &handler{cfg: cfg, route: route, uc: uc}
}

// Handle registers routes for this module.
func (h *handler) Handle() {
	groupPublic := h.route.Group("/%[1]s")
	groupInternal := h.route.Group("/internal/%[1]s")
	groupPrivate := h.route.Group("/%[1]s")

	groupInternal.Use(middleware.BasicAuthLogged(h.cfg))
	groupPrivate.Use(middleware.MustLogged(h.cfg))

	// ntaps:routes
}
`, pkg, mod, toExport(pkg))
		return writeGoFile(path, body)
	}
	return nil
}

/* ---------- method & route generation ---------- */

func ensureHandlerMethodAndRoute(pkg, ucPkg, endpointType, endpoint, handlerMethod, ucMethodName string, withParamUc, withResponseUc bool, tag, verb string) error {
	mod := modulePathGuess()
	path := filepath.Join(handlerRootHTTPDir, pkg, handlerPkgFileName)

	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	src := string(b)

	// imports required for method
	required := []string{
		fmt.Sprintf(`"%s/internal/usecase/%s"`, mod, ucPkg),
		`"github.com/AndreeJait/go-utility/response"`,
		`"github.com/AndreeJait/go-utility/tracer"`,
		`"github.com/labstack/echo/v4"`,
	}
	for _, imp := range required {
		if !strings.Contains(src, imp) {
			src = insertImport(src, imp)
		}
	}

	// ensure route line in Handle()
	groupName := map[string]string{"public": "groupPublic", "internal": "groupInternal", "private": "groupPrivate"}[strings.ToLower(endpointType)]
	if groupName == "" {
		groupName = "groupPublic"
	}
	verbUpper := strings.ToUpper(verb)
	routeLine := fmt.Sprintf(`%s.%s("%s", h.%s)`, groupName, verbUpper, endpoint, handlerMethod) // no leading tab
	routeLineTabbed := "\t" + routeLine                                                          // in case user already has a tabbed version

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

	// ensure handler method exists
	methodSig := fmt.Sprintf("func (h *handler) %s(", handlerMethod)
	if !strings.Contains(src, methodSig) {
		// swagger bits
		security := ""
		switch strings.ToLower(endpointType) {
		case "internal":
			security = "// @Security BasicAuth\n"
		case "private":
			security = "// @Security BearerAuth\n"
		}

		human := humanize(ucMethodName)

		// Param location depends on verb
		paramLoc := "body"
		if verbUpper == "GET" {
			paramLoc = "query"
		}

		// Optional swagger @Param line (only when withParamUc)
		paramAnnot := ""
		if withParamUc {
			paramAnnot = fmt.Sprintf(`// @Param       request %s %s.%sRequest true "%sRequest"
`, paramLoc, ucPkg, ucMethodName, ucMethodName)
		}

		paramLine := ""
		callArgs := "ctx"
		respType := ""
		retOK := `return response.SuccessOK(c, nil, "success ` + strings.ToLower(human) + `")`

		if withParamUc {
			paramType := fmt.Sprintf("%s.%sRequest", ucPkg, ucMethodName)
			paramLine = fmt.Sprintf(`
	param := %s{}
	if err := c.Bind(&param); err != nil { return err }
`, paramType)
			callArgs = "ctx, param"
		}
		if withResponseUc {
			respType = fmt.Sprintf("%s.%sResponse", ucPkg, ucMethodName)
			retOK = `return response.SuccessOK(c, resp, "success ` + strings.ToLower(human) + `")`
		}

		var callLine string
		if withResponseUc {
			callLine = fmt.Sprintf("resp, err := h.uc.%sUc.%s(%s)", toExport(ucPkg), ucMethodName, callArgs)
		} else {
			callLine = fmt.Sprintf("err := h.uc.%sUc.%s(%s)", toExport(ucPkg), ucMethodName, callArgs)
		}

		method := fmt.Sprintf(`

// %s godoc
// @Summary      %s
// @Description  %s
// @Tags         %s
// @Accept       json
// @Produce      json
%s%s// @Success     200 {object} response.Response%s "%s"
// @Failure      400 {object} response.ErrorResponse "validation/bind error"
// @Failure      500 {object} response.ErrorResponse "internal error"
// @Router       %s [%s]
func (h *handler) %s(c echo.Context) error {
	ctx := c.Request().Context()
	span, ctx := tracer.StartSpan(ctx, tracer.GetFuncName(h.%s))
	defer span.End()%s

	%s
	if err != nil { return err }
	%s
}
`,
			handlerMethod,
			human, human, tag,
			security,
			paramAnnot,
			conditional(withResponseUc, "{data="+respType+"}"), "success "+strings.ToLower(human),
			routerPath(pkg, endpointType, endpoint), strings.ToLower(verbUpper),
			handlerMethod, handlerMethod, paramLine,
			callLine, retOK,
		)
		src += method
	}

	return writeGoFile(path, src)
}

/* ---------- infra di/handler.go registration ---------- */

func updateInfraHandlerInit(pkg string) error {
	mod := modulePathGuess()
	path := handlerInfraInitPath
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("cannot find %s to register handler", path)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	src := string(b)

	importLine := fmt.Sprintf(`"%s/internal/adapters/inbound/http/%s"`, mod, pkg)
	if !strings.Contains(src, importLine) {
		src = insertImport(src, importLine)
	}

	newItem := fmt.Sprintf("\t\t%s.New%sHandler(s.cfg, groupV1, s.uc),", pkg, toExport(pkg))
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
	return writeGoFile(path, src)
}

/* ---------- small helpers (handler-specific) ---------- */

func humanize(s string) string {
	re := regexp.MustCompile(`([a-z0-9])([A-Z])`)
	return strings.TrimSpace(re.ReplaceAllString(s, "$1 $2"))
}

func routerPath(pkg, endpointType, endpoint string) string {
	switch strings.ToLower(endpointType) {
	case "internal":
		return "/internal/" + pkg + endpoint
	default: // public or private share /{pkg}
		return "/" + pkg + endpoint
	}
}

func conditional(ok bool, s string) string {
	if ok {
		return s
	}
	return ""
}
