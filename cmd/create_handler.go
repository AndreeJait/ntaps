package cmd

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/AndreeJait/ntaps/gen/handler"
	"github.com/AndreeJait/ntaps/internal/util"
)

func runCreateHandlerCmd(args []string) {
	fs := flag.NewFlagSet("create-handler", flag.ExitOnError)

	var pkg, ucPkg, endpointType, endpoint, ucMethodName, method, tag, verb string
	var withParamUc, withResponseUc bool

	fs.StringVar(&pkg, "pkg", "", "handler package name (e.g., send)")
	fs.StringVar(&ucPkg, "ucPkg", "", "usecase package to call (e.g., send)")
	fs.StringVar(&endpointType, "endpointType", "public", "public|internal|private")
	fs.StringVar(&endpoint, "endpoint", "", "endpoint path (e.g., /submit/cash-to-cash)")
	fs.BoolVar(&withParamUc, "withParamUc", false, "usecase method takes a Request")
	fs.BoolVar(&withResponseUc, "withResponseUc", false, "usecase method returns a Response")
	fs.StringVar(&ucMethodName, "ucMethodName", "", "usecase method name (PascalCase)")
	fs.StringVar(&method, "method", "", "handler method name (lowerCamel, e.g., submitCashToCash)")
	fs.StringVar(&tag, "tag", "", "swagger tag; default: CamelCase of --pkg")
	fs.StringVar(&verb, "verb", "POST", "HTTP verb: GET|POST|PUT|DELETE")
	_ = fs.Parse(args)

	if len(args) == 0 || os.Getenv("NTAPS_INTERACTIVE") == "1" {
		interactiveHandler(&pkg, &ucPkg, &withParamUc, &withResponseUc, &ucMethodName, &method, &endpointType, &endpoint, &tag, &verb)
	}

	// Skeleton mode: just create pkg & register
	if pkg != "" && ucPkg == "" && endpoint == "" && ucMethodName == "" && method == "" {
		if err := handler.EnsurePackageOnly(pkg); err != nil {
			exitErr(err.Error())
		}
		fmt.Printf("✅ Done: handler skeleton created & registered for pkg=%s\n", pkg)
		return
	}

	// Full mode
	if pkg == "" || ucPkg == "" || endpoint == "" || ucMethodName == "" || method == "" {
		exitErr("missing required values: need pkg, ucPkg, endpoint, ucMethodName, method (leave others empty ONLY for skeleton)")
	}
	if !isPascalCase(ucMethodName) {
		exitErr("--ucMethodName must be PascalCase")
	}
	if tag == "" {
		tag = util.ToPascalCase(pkg)
	}
	if endpoint != "" && endpoint[0] != '/' {
		endpoint = "/" + endpoint
	}
	verb = strings.ToUpper(strings.TrimSpace(verb))
	switch verb {
	case "GET", "POST", "PUT", "DELETE":
	default:
		exitErr("--verb must be one of GET|POST|PUT|DELETE")
	}

	if err := handler.Run(
		pkg,
		ucPkg,
		endpointType,
		endpoint,
		withParamUc,
		withResponseUc,
		ucMethodName,
		method,
		tag,
		verb,
	); err != nil {
		exitErr(err.Error())
	}

	fmt.Printf("✅ Done: handler=%s method=%s (%s %s) → uc=%s.%s\n", pkg, method, verb, endpointType, ucPkg, ucMethodName)
}
