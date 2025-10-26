package cmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/AndreeJait/ntaps/gen/usecase"
)

func runCreateUsecaseCmd(args []string) {
	fs := flag.NewFlagSet("create-usecase", flag.ExitOnError)

	var pkg, method string
	var withParam, withResp bool

	fs.StringVar(&pkg, "pkg", "", "usecase package name (e.g., send)")
	fs.StringVar(&method, "method", "", "method name in PascalCase (e.g., SubmitCashToCash)")
	fs.BoolVar(&withParam, "withParam", false, "generate a Param struct <MethodName>Request")
	fs.BoolVar(&withResp, "withResponse", false, "generate a Response struct <MethodName>Response")
	_ = fs.Parse(args)

	if len(args) == 0 || os.Getenv("NTAPS_INTERACTIVE") == "1" {
		interactiveUsecase(&pkg, &method, &withParam, &withResp)
	}

	if pkg == "" || method == "" {
		exitErr("usage: ntaps create-usecase --pkg=<name> --method=<Pascal> [--withParam] [--withResponse]")
	}
	if !isPascalCase(method) {
		exitErr("method must be PascalCase")
	}

	if err := usecase.Run(pkg, method, withParam, withResp); err != nil {
		exitErr(err.Error())
	}

	fmt.Printf("✅ Done: usecase=%s method=%s (withParam=%v, withResponse=%v)\n", pkg, method, withParam, withResp)
}
