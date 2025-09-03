package main

import (
	"bufio"
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

	outboundRootPath = "internal/adapters/outbound"
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

		// Interactive mode if no flags (or forced)
		if len(os.Args) == 2 || os.Getenv("NTAPS_INTERACTIVE") == "1" {
			interactiveCreateUsecase(&pkg, &method, &withParam, &withResp)
		}

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

		if len(os.Args) == 2 || os.Getenv("NTAPS_INTERACTIVE") == "1" {
			interactiveCreateHandler(&pkg, &ucPkg, &withParamUc, &withResponseUc, &ucMethodName, &method, &endpointType, &endpoint, &tag, &verb)
		}

		// Skeleton mode: only pkg → create package and register handler, skip method/route
		if pkg != "" && ucPkg == "" && endpoint == "" && ucMethodName == "" && method == "" {
			if err := ensureHandlerPkg(pkg); err != nil {
				exitErr(err.Error())
			}
			if err := updateInfraHandlerInit(pkg); err != nil {
				exitErr(err.Error())
			}
			fmt.Printf("✅ Done: handler skeleton created & registered for pkg=%s\n", pkg)
			return
		}

		// Full validation for method/route
		if pkg == "" || ucPkg == "" || endpoint == "" || ucMethodName == "" || method == "" {
			exitErr("missing required values: need pkg, ucPkg, endpoint, ucMethodName, method (leave others empty ONLY for skeleton)")
		}
		if strings.ToUpper(ucMethodName[:1]) != ucMethodName[:1] {
			exitErr("--ucMethodName must be PascalCase")
		}
		if tag == "" {
			tag = toExport(pkg)
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

		if len(os.Args) == 2 || os.Getenv("NTAPS_INTERACTIVE") == "1" {
			interactiveCreateRepository(&rtype, &pkg, &method, &withParam, &withResp, &withTx, &addToUC)
		}

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

	case "create-outbound":
		fs := flag.NewFlagSet("create-outbound", flag.ExitOnError)
		var pkg, method string
		var withParam, withResp bool
		fs.StringVar(&pkg, "pkg", "", "outbound package name (e.g., email)")
		fs.StringVar(&method, "method", "", "method name in PascalCase (e.g., SendEmailActivation)")
		fs.BoolVar(&withParam, "withParam", false, "generate <Method>Request")
		fs.BoolVar(&withResp, "withResp", false, "generate <Method>Response")
		_ = fs.Parse(os.Args[2:])

		if len(os.Args) == 2 || os.Getenv("NTAPS_INTERACTIVE") == "1" {
			interactiveCreateOutbound(&pkg, &method, &withParam, &withResp)
		}

		if pkg == "" {
			exitErr("usage: ntaps create-outbound --pkg=<name> [--method=<Pascal>] [--withParam] [--withResp]")
		}
		if method != "" && strings.ToUpper(method[:1]) != method[:1] {
			exitErr("--method must be PascalCase")
		}
		if err := runCreateOutbound(pkg, method, withParam, withResp); err != nil {
			exitErr(err.Error())
		}
		if method == "" {
			fmt.Printf("✅ Done: outbound=%s created\n", pkg)
		} else {
			fmt.Printf("✅ Done: outbound=%s method=%s (withParam=%v, withResp=%v)\n", pkg, method, withParam, withResp)
		}

	default:
		usage()
	}
}

func usage() {
	fmt.Println(`ntaps <command> [flags]

Commands:
  create-usecase     scaffold/extend a usecase package & method (interactive if no flags)
  create-handler     scaffold/extend an inbound HTTP handler & route (interactive if no flags)
  create-repository  scaffold/extend a postgres repository and wire into DI (interactive if no flags)
  create-outbound    scaffold/extend an outbound adapter (interactive if no flags)

Interactive examples:
  ntaps create-usecase
  ntaps create-handler
  ntaps create-repository
  ntaps create-outbound

Flag examples:
  ntaps create-usecase --pkg=send --method=SubmitCashToCash --withParam --withResponse
  ntaps create-handler --pkg=send --ucPkg=send --endpointType=private --endpoint=/submit/cash-to-cash --withParamUc --withResponseUc --ucMethodName=SubmitCashToCash --method=submitCashToCash --tag=Send --verb=POST
  ntaps create-repository --type=postgres --pkg=user --method=UpdateUserStatus --withParamRepo --withResponseRepo --withTx --addToUC=send
  ntaps create-outbound --pkg=email --method=SendEmailActivation --withParam --withResp`)
	os.Exit(2)
}

/* ============================ interactive helpers ============================ */

func interactiveCreateUsecase(pkg, method *string, withParam, withResp *bool) {
	rd := bufio.NewReader(os.Stdin)
	fmt.Println("🛠  create-usecase (press Enter to keep defaults / leave empty)")

	*pkg = promptString(rd, "pkg", *pkg)
	*method = promptString(rd, "method (PascalCase)", *method)
	*withParam = promptBool(rd, "withParam", *withParam)
	*withResp = promptBool(rd, "withResponse", *withResp)
}

func interactiveCreateHandler(pkg, ucPkg *string, withParamUc, withResponseUc *bool, ucMethodName, method, endpointType, endpoint, tag, verb *string) {
	rd := bufio.NewReader(os.Stdin)
	fmt.Println("🛠  create-handler (press Enter to keep defaults / leave empty)")

	*pkg = promptString(rd, "pkg", *pkg)
	*ucPkg = promptString(rd, "ucPkg", *ucPkg)
	*withParamUc = promptBool(rd, "withParamUc", *withParamUc)
	*withResponseUc = promptBool(rd, "withResponseUc", *withResponseUc)
	*ucMethodName = promptString(rd, "ucMethodName (PascalCase)", *ucMethodName)
	*method = promptString(rd, "method (lowerCamel)", *method)

	defET := *endpointType
	if defET == "" {
		defET = "public"
	}
	*endpointType = promptString(rd, "endpointType [public|internal|private]", defET)
	*endpoint = promptString(rd, "endpoint (e.g., /submit/cash-to-cash)", *endpoint)
	*tag = promptString(rd, "tag (default CamelCase of pkg)", *tag)

	defVerb := *verb
	if defVerb == "" {
		defVerb = "POST"
	}
	*verb = promptString(rd, "verb [GET|POST|PUT|DELETE]", defVerb)
}

func interactiveCreateRepository(rtype, pkg, method *string, withParam, withResp, withTx *bool, addToUC *string) {
	rd := bufio.NewReader(os.Stdin)
	fmt.Println("🛠  create-repository (press Enter to keep defaults / leave empty)")

	defType := *rtype
	if defType == "" {
		defType = "postgres"
	}
	*rtype = promptString(rd, "type [postgres]", defType)
	*pkg = promptString(rd, "pkg", *pkg)
	*method = promptString(rd, "method (PascalCase)", *method)
	*withParam = promptBool(rd, "withParamRepo", *withParam)
	*withResp = promptBool(rd, "withResponseRepo", *withResp)
	*withTx = promptBool(rd, "withTx", *withTx)
	*addToUC = promptString(rd, "addToUC (optional usecase pkg)", *addToUC)
}

func interactiveCreateOutbound(pkg, method *string, withParam, withResp *bool) {
	rd := bufio.NewReader(os.Stdin)
	fmt.Println("🛠  create-outbound (press Enter to keep defaults / leave empty)")

	*pkg = promptString(rd, "pkg", *pkg)
	*method = promptString(rd, "method (PascalCase; optional)", *method)
	*withParam = promptBool(rd, "withParam", *withParam)
	*withResp = promptBool(rd, "withResp", *withResp)
}

func promptString(rd *bufio.Reader, label, def string) string {
	if def != "" {
		fmt.Printf("%s [%s]: ", label, def)
	} else {
		fmt.Printf("%s: ", label)
	}
	text, _ := rd.ReadString('\n')
	text = strings.TrimSpace(text)
	if text == "" {
		return def
	}
	return text
}

func promptBool(rd *bufio.Reader, label string, def bool) bool {
	defStr := "false"
	if def {
		defStr = "true"
	}
	fmt.Printf("%s [%s]: ", label, defStr)
	text, _ := rd.ReadString('\n')
	text = strings.TrimSpace(strings.ToLower(text))
	if text == "" {
		return def
	}
	switch text {
	case "1", "t", "true", "y", "yes":
		return true
	case "0", "f", "false", "n", "no":
		return false
	default:
		fmt.Println("  (invalid bool, keeping default)")
		return def
	}
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

var wordRe = regexp.MustCompile(`[A-Za-z0-9]+`)

func toPascalCase(s string) string {
	if s == "" {
		return ""
	}
	parts := wordRe.FindAllString(s, -1) // splits on non-alnum: "_", "-", ".", etc.
	if len(parts) == 0 {
		return ""
	}
	for i, p := range parts {
		if len(p) == 1 {
			parts[i] = strings.ToUpper(p)
			continue
		}
		parts[i] = strings.ToUpper(p[:1]) + strings.ToLower(p[1:])
	}
	return strings.Join(parts, "")
}

// lowerCamelCase
func toCamelCase(s string) string {
	p := toPascalCase(s)
	if p == "" {
		return ""
	}
	return strings.ToLower(p[:1]) + p[1:]
}

// Backward-compat shim: everywhere you used toExport(...), you now get PascalCase.
func toExport(s string) string { return toPascalCase(s) }
