package handler

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/AndreeJait/ntaps/internal/util"
)

// normalizePathParams turns /send/transaction/:transaction_code
// into /send/transaction/{transaction_code} for swagger,
// and returns ["transaction_code"].
func normalizePathParams(endpoint string) (normalized string, params []string) {
	// convert :param -> {param}
	reColon := regexp.MustCompile(`:([A-Za-z0-9_]+)`)
	out := reColon.ReplaceAllStringFunc(endpoint, func(m string) string {
		name := strings.TrimPrefix(m, ":")
		params = append(params, name)
		return "{" + name + "}"
	})

	// also collect already-braced params {foo}
	reBrace := regexp.MustCompile(`\{([A-Za-z0-9_]+)\}`)
	matches := reBrace.FindAllStringSubmatch(out, -1)
	for _, m := range matches {
		if len(m) > 1 {
			name := m[1]
			dup := false
			for _, p := range params {
				if p == name {
					dup = true
					break
				}
			}
			if !dup {
				params = append(params, name)
			}
		}
	}

	return out, params
}

// buildHandlerMethod generates the full handler method (swagger block + func body).
func buildHandlerMethod(
	handlerMethod string,
	ucPkg string,
	ucMethodName string,
	endpointType string,
	httpVerb string,
	endpoint string,
	withParamUc bool,
	withResponseUc bool,
	tag string,
) string {

	// Security annotation
	security := ""
	switch strings.ToLower(endpointType) {
	case "internal":
		security = "// @Security BasicAuth\n"
	case "private":
		security = "// @Security BearerAuth\n"
	}

	human := util.HumanizePascal(ucMethodName)

	// GET => request comes from query, others => body
	paramLoc := "body"
	if strings.EqualFold(httpVerb, "GET") {
		paramLoc = "query"
	}

	// Normalize /foo/:code -> /foo/{code} and collect ["code"]
	normEndpoint, pathParams := normalizePathParams(endpoint)

	// Path param swagger lines
	// @Param transaction_code path string true "Transaction Code"
	pathParamAnnots := ""
	for _, p := range pathParams {
		pathParamAnnots += fmt.Sprintf(
			`// @Param        %s path string true "%s"`+"\n",
			p,
			util.HumanizePascal(p),
		)
	}

	// Body/query param swagger line (+ bind code)
	paramAnnot := ""
	paramLine := ""
	callArgs := "ctx"

	if withParamUc {
		paramType := fmt.Sprintf("%s.%sRequest", ucPkg, ucMethodName)

		// ✅ FIXED: single schema token transfer.InsertIntoManualTransferHistoriesRequest
		paramAnnot = fmt.Sprintf(`// @Param       request %s %s true "%sRequest"
`, paramLoc, paramType, ucMethodName)

		paramLine = fmt.Sprintf(`
	param := %s{}
	if err := c.Bind(&param); err != nil { return err }
`, paramType)

		callArgs = "ctx, param"
	}

	// Response wiring
	respType := ""
	returnOK := `return response.SuccessOK(c, nil, "success ` + strings.ToLower(human) + `")`

	if withResponseUc {
		respType = fmt.Sprintf("%s.%sResponse", ucPkg, ucMethodName)
		returnOK = `return response.SuccessOK(c, resp, "success ` + strings.ToLower(human) + `")`
	}

	// Call the UC
	var callLine string
	if withResponseUc {
		callLine = fmt.Sprintf(
			"resp, err := h.uc.%sUc.%s(%s)",
			util.ToPascalCase(ucPkg),
			ucMethodName,
			callArgs,
		)
	} else {
		callLine = fmt.Sprintf(
			"err := h.uc.%sUc.%s(%s)",
			util.ToPascalCase(ucPkg),
			ucMethodName,
			callArgs,
		)
	}

	// Swagger @Router path needs prefix (/pkg or /internal/pkg)
	fullRoute := util.RouterPath(ucPkg, endpointType, normEndpoint)

	return fmt.Sprintf(`

// %s godoc
// @Summary      %s
// @Description  %s
// @Tags         %s
// @Accept       json
// @Produce      json
%s%s%s// @Success     200 {object} response.Response%s "success %s"
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
		pathParamAnnots,
		paramAnnot,
		util.Conditional(withResponseUc, "{data="+respType+"}"),
		strings.ToLower(human),
		fullRoute,
		strings.ToLower(httpVerb),
		handlerMethod, handlerMethod, paramLine,
		callLine, returnOK,
	)
}
