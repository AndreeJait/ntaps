package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// shared reader
func rd() *bufio.Reader {
	return bufio.NewReader(os.Stdin)
}

func promptString(label, def string) string {
	r := rd()
	if def != "" {
		fmt.Printf("%s [%s]: ", label, def)
	} else {
		fmt.Printf("%s: ", label)
	}
	text, _ := r.ReadString('\n')
	text = strings.TrimSpace(text)
	if text == "" {
		return def
	}
	return text
}

func promptBool(label string, def bool) bool {
	r := rd()
	defStr := "false"
	if def {
		defStr = "true"
	}
	fmt.Printf("%s [%s]: ", label, defStr)
	text, _ := r.ReadString('\n')
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

/* ----- interactive prompts per command ----- */

func interactiveUsecase(pkg, method *string, withParam, withResp *bool) {
	fmt.Println("🛠  create-usecase (press Enter to keep defaults / leave empty)")
	*pkg = promptString("pkg", *pkg)
	*method = promptString("method (PascalCase)", *method)
	*withParam = promptBool("withParam", *withParam)
	*withResp = promptBool("withResponse", *withResp)
}

func interactiveHandler(
	pkg, ucPkg *string,
	withParamUc, withResponseUc *bool,
	ucMethodName, method, endpointType, endpoint, tag, verb *string,
) {
	fmt.Println("🛠  create-handler (press Enter to keep defaults / leave empty)")

	*pkg = promptString("pkg", *pkg)
	*ucPkg = promptString("ucPkg", *ucPkg)
	*withParamUc = promptBool("withParamUc", *withParamUc)
	*withResponseUc = promptBool("withResponseUc", *withResponseUc)
	*ucMethodName = promptString("ucMethodName (PascalCase)", *ucMethodName)
	*method = promptString("method (lowerCamel)", *method)

	defET := *endpointType
	if defET == "" {
		defET = "public"
	}
	*endpointType = promptString("endpointType [public|internal|private]", defET)
	*endpoint = promptString("endpoint (e.g., /submit/cash-to-cash or /transaction/:transaction_code)", *endpoint)
	*tag = promptString("tag (default CamelCase of pkg)", *tag)

	defVerb := *verb
	if defVerb == "" {
		defVerb = "POST"
	}
	*verb = promptString("verb [GET|POST|PUT|DELETE]", defVerb)
}

func interactiveRepository(
	rtype, pkg, method *string,
	withParam, withResp, withTx *bool,
	addToUC *string,
) {
	fmt.Println("🛠  create-repository (press Enter to keep defaults / leave empty)")

	if *rtype == "" {
		*rtype = "postgres"
	}
	*rtype = promptString("type [postgres]", *rtype)
	*pkg = promptString("pkg", *pkg)
	*method = promptString("method (PascalCase)", *method)
	*withParam = promptBool("withParamRepo", *withParam)
	*withResp = promptBool("withResponseRepo", *withResp)
	*withTx = promptBool("withTx", *withTx)
	*addToUC = promptString("addToUC (optional usecase pkg)", *addToUC)
}

func interactiveOutbound(pkg, method *string, withParam, withResp *bool) {
	fmt.Println("🛠  create-outbound (press Enter to keep defaults / leave empty)")
	*pkg = promptString("pkg", *pkg)
	*method = promptString("method (PascalCase; optional)", *method)
	*withParam = promptBool("withParam", *withParam)
	*withResp = promptBool("withResp", *withResp)
}

// NEW: interactive for add-repo-to-usecase
func interactiveAddRepoToUsecase(
	repoPkg, ucPkg, method *string,
) {
	fmt.Println("🛠  add-repo-to-usecase (press Enter to keep defaults / leave empty)")
	*repoPkg = promptString("repoPkg (repository package, e.g. customer)", *repoPkg)
	*ucPkg = promptString("ucPkg (usecase package, e.g. send)", *ucPkg)
	*method = promptString("method (PascalCase, e.g. GetCustomerByID)", *method)
}
