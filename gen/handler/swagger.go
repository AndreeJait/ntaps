package handler

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/AndreeJait/ntaps/internal/util"
)

// normalizePathParams turns /send/transaction/:transaction_code
// into /send/transaction/{transaction_code} for OpenAPI.
func normalizePathParams(endpoint string) (normalized string, params []string) {
	// case 1: Echo-style :param
	//   /foo/:bar -> /foo/{bar}
	reColon := regexp.MustCompile(`:([A-Za-z0-9_]+)`)
	out := reColon.ReplaceAllStringFunc(endpoint, func(m string) string {
		name := strings.TrimPrefix(m, ":")
		params = append(params, name)
		return "{" + name + "}"
	})

	// case 2: already {param}, collect them too
	reBrace := regexp.MustCompile(`\{([A-Za-z0-9_]+)\}`)
	matches := reBrace.FindAllStringSubmatch(out, -1)
	for _, m := range matches {
		if len(m) > 1 {
			name := m[1]
			// avoid duplicates
			found := false
			for _, p := range params {
				if p == name {
					found = true
					break
				}
			}
			if !found {
				params = append(params, name)
			}
		}
	}

	return out, params
}

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

	// security annotation
	security := ""
	switch strings.ToLower(endpointType) {
	case "internal":
		security = "// @Security BasicAuth\n"
	case "private":
		security = "// @Security BearerAuth\n"
	}

	human := util.HumanizePascal(ucMethodName)

	// figure out where the request comes from
	paramLoc := "body"
	if httpVerb == "GET" {
		paramLoc = "query"
	}

	// normalize :param -> {param} for swagger
	normEndpoint, pathParams := normalizePathParams(endpoint)

	// build swagger @Param lines for path params
	pathParamAnnots := ""
	for _, p := range pathParams {
		// standard: path params are always required and string (you can customize later)
		pathParamAnnots += fmt.Sprintf(`// @Param        %s path string true "%s"`+"\n", p, util.HumanizePascal(p))
	}

	// build request param annot (@Param request ...)
	paramAnnot := ""
	paramLine := ""
	callArgs := "ctx"

	if withParamUc {
		paramType := fmt.Sprintf("%s.%sRequest", ucPkg, ucMethodName)

		// Only include the body/query annotation if there will actually be a body/query bind.
		// Path params are documented separately via pathParamAnnots.
		paramAnnot = fmt.Sprintf(`// @Param       request %s %s %s true "%sRequest"
`, paramLoc, ucPkg, ucMethodName, ucMethodName)

		paramLine = fmt.Sprintf(`
	param := %s{}
	if err := c.Bind(&param); err != nil { return err }
`, paramType)

		callArgs = "ctx, param"
	}

	// build success response annotation and return body
	respType := ""
	retOK := `return response.SuccessOK(c, nil, "success ` + strings.ToLower(human) + `")`
	if withResponseUc {
		respType = fmt.Sprintf("%s.%sResponse", ucPkg, ucMethodName)
		retOK = `return response.SuccessOK(c, resp, "success ` + strings.ToLower(human) + `")`
	}

	// call the actual uc
	var callLine string
	if withResponseUc {
		callLine = fmt.Sprintf("resp, err := h.uc.%sUc.%s(%s)", util.ToPascalCase(ucPkg), ucMethodName, callArgs)
	} else {
		callLine = fmt.Sprintf("err := h.uc.%sUc.%s(%s)", util.ToPascalCase(ucPkg), ucMethodName, callArgs)
	}

	// choose full route including /send, /internal/send, etc.
	fullRoute := util.RouterPath(ucPkg, endpointType, normEndpoint)

	// final swagger block
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
		callLine, retOK,
	)
}
