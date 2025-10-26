package cmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/AndreeJait/ntaps/gen/repo"
)

func runCreateRepositoryCmd(args []string) {
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
	_ = fs.Parse(args)

	if len(args) == 0 || os.Getenv("NTAPS_INTERACTIVE") == "1" {
		interactiveRepository(&rtype, &pkg, &method, &withParam, &withResp, &withTx, &addToUC)
	}

	if rtype != "postgres" {
		exitErr("--type currently supports only 'postgres'")
	}
	if pkg == "" || method == "" {
		exitErr("usage: ntaps create-repository --type=postgres --pkg=<pkg> --method=<Pascal> [--withParamRepo] [--withResponseRepo] [--withTx] [--addToUC=<usecase>]")
	}
	if !isPascalCase(method) {
		exitErr("--method must be PascalCase")
	}

	if err := repo.Run(pkg, method, withParam, withResp, withTx, addToUC); err != nil {
		exitErr(err.Error())
	}

	fmt.Printf(
		"✅ Done: repository=%s method=%s (param=%v, resp=%v, tx=%v) wiredToUC=%s\n",
		pkg, method, withParam, withResp, withTx, addToUC,
	)
}
