package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
)

/* ===== shared constants (used by other files in same package) ===== */

const (
	rootUsecaseDir       = "internal/usecase"
	usecaseDIPath        = "internal/usecase/di.go"
	infraInitUsecasePath = "internal/infrastructure/di/usecase.go"

	handlerRootHTTPDir   = "internal/adapters/inbound/http"
	handlerPkgFileName   = "di.go"
	handlerInfraInitPath = "internal/infrastructure/di/handler.go"

	repoRootPath      = "internal/adapters/outbound/db"
	repoPgPath        = "internal/adapters/outbound/db/postgres"
	pgDiPath          = "internal/adapters/outbound/db/di.go"
	infraRepoInitPath = "internal/infrastructure/di/repository.go"
)

/* ============================ entry point ============================ */

func main() {
	if len(os.Args) < 2 {
		usage()
	}

	switch os.Args[1] {
	case "create-usecase":
		fs := flag.NewFlagSet("create-usecase", flag.ExitOnError)
		var pkg, method string
		var withParam, withResp bool
		fs.StringVar(&pkg, "pkg", "", "usecase package name (e.g., send)")
		fs.StringVar(&method, "method", "", "method name in PascalCase (e.g., SubmitCashToCash)")
		fs.BoolVar(&withParam, "withParam", false, "generate a Param struct <MethodName>Request")
		fs.BoolVar(&withResp, "withResponse", false, "generate a Response struct <MethodName>Response")
		_ = fs.Parse(os.Args[2:])

		if pkg == "" || method == "" {
			exitErr("usage: ntaps create-usecase --pkg=<name> --method=<Pascal> [--withParam] [--withResponse]")
		}
		if strings.ToUpper(method[:1]) != method[:1] {
			exitErr("method must be PascalCase")
		}
		if err := runCreateUsecase(pkg, method, withParam, withResp); err != nil {
			exitErr(err.Error())
		}
		fmt.Printf("✅ Done: usecase=%s method=%s (withParam=%v, withResponse=%v)\n", pkg, method, withParam, withResp)

	case "create-handler":
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
		_ = fs.Parse(os.Args[2:])

		if pkg == "" || ucPkg == "" || endpoint == "" || ucMethodName == "" || method == "" {
			exitErr("usage: ntaps create-handler --pkg=<pkg> --ucPkg=<ucPkg> --endpointType=public|internal|private --endpoint=/path --withParamUc --withResponseUc --ucMethodName=<Pascal> --method=<lowerCamel> [--tag=Tag] [--verb=GET|POST|PUT|DELETE]")
		}
		if strings.ToUpper(ucMethodName[:1]) != ucMethodName[:1] {
			exitErr("--ucMethodName must be PascalCase")
		}
		if tag == "" {
			tag = toExport(pkg)
		}
		if endpoint[0] != '/' {
			endpoint = "/" + endpoint
		}
		verb = strings.ToUpper(verb)
		switch verb {
		case "GET", "POST", "PUT", "DELETE":
		default:
			exitErr("--verb must be one of GET|POST|PUT|DELETE")
		}

		if err := runCreateHandler(pkg, ucPkg, endpointType, endpoint, withParamUc, withResponseUc, ucMethodName, method, tag, verb); err != nil {
			exitErr(err.Error())
		}
		fmt.Printf("✅ Done: handler=%s method=%s (%s %s) → uc=%s.%s\n", pkg, method, verb, endpointType, ucPkg, ucMethodName)

	case "create-repository":
		fs := flag.NewFlagSet("create-repository", flag.ExitOnError)
		var rtype, pkg, method, addToUC string
		var withParam, withResp, withTx bool
		fs.StringVar(&rtype, "type", "postgres", "repository backend type (postgres)")
		fs.StringVar(&pkg, "pkg", "", "repository package (e.g., user)")
		fs.StringVar(&method, "method", "", "repository method name (PascalCase, e.g., UpdateUserStatus)")
		fs.BoolVar(&withParam, "withParamRepo", false, "generate <Method>Param")
		fs.BoolVar(&withResp, "withResponseRepo", false, "generate <Method>Response")
		fs.BoolVar(&withTx, "withTx", false, "include tx pgx.Tx parameter")
		fs.StringVar(&addToUC, "addToUC", "", "usecase pkg to wire this repo into (e.g., send)")
		_ = fs.Parse(os.Args[2:])

		if rtype != "postgres" {
			exitErr("--type currently supports only 'postgres'")
		}
		if pkg == "" || method == "" {
			exitErr("usage: ntaps create-repository --type=postgres --pkg=<pkg> --method=<Pascal> [--withParamRepo] [--withResponseRepo] [--withTx] [--addToUC=<usecase>]")
		}
		if strings.ToUpper(method[:1]) != method[:1] {
			exitErr("--method must be PascalCase")
		}
		if err := runCreateRepository(pkg, method, withParam, withResp, withTx, addToUC); err != nil {
			exitErr(err.Error())
		}
		fmt.Printf("✅ Done: repository=%s method=%s (param=%v, resp=%v, tx=%v) wiredToUC=%s\n", pkg, method, withParam, withResp, withTx, addToUC)

	default:
		usage()
	}
}

func usage() {
	fmt.Println(`ntaps <command> [flags]

Commands:
  create-usecase     scaffold/extend a usecase package & method
  create-handler     scaffold/extend an inbound HTTP handler & route
  create-repository  scaffold/extend a postgres repository and wire into DI

Examples:
  ntaps create-usecase --pkg=send --method=SubmitCashToCash --withParam --withResponse
  ntaps create-handler --pkg=send --ucPkg=send --endpointType=private --endpoint=/submit/cash-to-cash --withParamUc --withResponseUc --ucMethodName=SubmitCashToCash --method=submitCashToCash --tag=Send --verb=POST
  ntaps create-repository --type=postgres --pkg=user --method=UpdateUserStatus --withParamRepo --withResponseRepo --withTx --addToUC=send`)
	os.Exit(2)
}

/* ============================ shared helpers ============================ */

func insertImport(body, importLine string) string {
	if strings.Contains(body, "import (") {
		return strings.Replace(body, "import (", "import (\n\t"+importLine+"\n", 1)
	}
	re := regexp.MustCompile(`import\s+"[^"]+"`)
	if re.MatchString(body) {
		old := re.FindString(body)
		newBlock := "import (\n\t" + importLine + "\n\t" + strings.TrimPrefix(old, "import ") + "\n)"
		return strings.Replace(body, old, newBlock, 1)
	}
	idx := strings.Index(body, "\n")
	return body[:idx+1] + "import (\n\t" + importLine + "\n)\n" + body[idx+1:]
}

// Reads ONLY first line of go.mod. Fallback to module name if missing.
func modulePathGuess() string {
	b, err := os.ReadFile("go.mod")
	if err != nil {
		return "go-template-hexagonal"
	}
	lines := strings.Split(string(b), "\n")
	if len(lines) > 0 && strings.HasPrefix(strings.TrimSpace(lines[0]), "module ") {
		return strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(lines[0]), "module"))
	}
	return "go-template-hexagonal"
}

func toExport(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func lowerFirst(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToLower(s[:1]) + s[1:]
}

func exitErr(msg string) {
	fmt.Fprintln(os.Stderr, "❌", msg)
	os.Exit(1)
}
