package cmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/AndreeJait/ntaps/gen/outbound"
)

func runCreateOutboundCmd(args []string) {
	fs := flag.NewFlagSet("create-outbound", flag.ExitOnError)

	var pkg, method string
	var withParam, withResp bool

	fs.StringVar(&pkg, "pkg", "", "outbound package name (e.g., email)")
	fs.StringVar(&method, "method", "", "method name in PascalCase (e.g., SendEmailActivation)")
	fs.BoolVar(&withParam, "withParam", false, "generate <Method>Request")
	fs.BoolVar(&withResp, "withResp", false, "generate <Method>Response")
	_ = fs.Parse(args)

	if len(args) == 0 || os.Getenv("NTAPS_INTERACTIVE") == "1" {
		interactiveOutbound(&pkg, &method, &withParam, &withResp)
	}

	if pkg == "" {
		exitErr("usage: ntaps create-outbound --pkg=<name> [--method=<Pascal>] [--withParam] [--withResp]")
	}
	if method != "" && !isPascalCase(method) {
		exitErr("--method must be PascalCase")
	}

	if err := outbound.Run(pkg, method, withParam, withResp); err != nil {
		exitErr(err.Error())
	}

	if method == "" {
		fmt.Printf("✅ Done: outbound=%s created\n", pkg)
	} else {
		fmt.Printf("✅ Done: outbound=%s method=%s (withParam=%v, withResp=%v)\n", pkg, method, withParam, withResp)
	}
}
