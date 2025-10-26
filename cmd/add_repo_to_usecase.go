package cmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/AndreeJait/ntaps/gen/repo"
)

func runAddRepoToUsecaseCmd(args []string) {
	fs := flag.NewFlagSet("add-repo-to-usecase", flag.ExitOnError)

	var repoPkg, ucPkg, method string
	var withParamRepo, withRespRepo, withTx bool

	fs.StringVar(&repoPkg, "repoPkg", "", "repository package name (e.g. customer)")
	fs.StringVar(&ucPkg, "ucPkg", "", "usecase package name to inject into (e.g. send)")
	fs.StringVar(&method, "method", "", "repository method name in PascalCase (e.g. GetCustomerByID)")

	// these flags still exist for power users in non-interactive mode:
	fs.BoolVar(&withParamRepo, "withParamRepo", false, "repository method takes a <Method>Param struct")
	fs.BoolVar(&withRespRepo, "withResponseRepo", false, "repository method returns a <Method>Response struct")
	fs.BoolVar(&withTx, "withTx", false, "repository method takes a pgx.Tx")

	_ = fs.Parse(args)

	// Interactive mode if no flags OR interactive explicitly forced.
	if len(args) == 0 || os.Getenv("NTAPS_INTERACTIVE") == "1" {
		// interactive now ONLY asks the 3 required values
		interactiveAddRepoToUsecase(&repoPkg, &ucPkg, &method)
		// and we intentionally leave withParamRepo/withRespRepo/withTx alone.
		// they will still be false (their zero value) unless the caller passed flags.
	}

	if repoPkg == "" || ucPkg == "" || method == "" {
		exitErr("usage: ntaps add-repo-to-usecase --repoPkg=<repo> --ucPkg=<usecase> --method=<Pascal> [--withParamRepo] [--withResponseRepo] [--withTx]")
	}
	if !isPascalCase(method) {
		exitErr("--method must be PascalCase")
	}

	if err := repo.AddRepoToUsecase(repoPkg, ucPkg, method, withParamRepo, withRespRepo, withTx); err != nil {
		exitErr(err.Error())
	}

	fmt.Printf("✅ Wired repo=%s into usecase=%s (method=%s)\n", repoPkg, ucPkg, method)
}
