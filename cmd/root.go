package cmd

import (
	"fmt"
	"os"
)

func Execute() {
	if len(os.Args) < 2 {
		usageAndExit()
	}

	switch os.Args[1] {
	case "create-usecase":
		runCreateUsecaseCmd(os.Args[2:])
	case "create-handler":
		runCreateHandlerCmd(os.Args[2:])
	case "create-repository":
		runCreateRepositoryCmd(os.Args[2:])
	case "create-outbound":
		runCreateOutboundCmd(os.Args[2:])
	case "add-repo-to-usecase":
		runAddRepoToUsecaseCmd(os.Args[2:])
	default:
		usageAndExit()
	}
}

func usageAndExit() {
	fmt.Println(`ntaps <command> [flags]

Commands:
  create-usecase         scaffold/extend a usecase package & method (interactive if no flags)
  create-handler         scaffold/extend an inbound HTTP handler & route (interactive if no flags)
  create-repository      scaffold/extend a postgres repository and wire into DI (interactive if no flags)
  create-outbound        scaffold/extend an outbound adapter (interactive if no flags)
  add-repo-to-usecase    wire an existing repository into an existing usecase (interactive if no flags)

Interactive examples:
  ntaps create-usecase
  ntaps create-handler
  ntaps create-repository
  ntaps create-outbound
  ntaps add-repo-to-usecase

Flag examples:
  ntaps create-usecase --pkg=send --method=SubmitCashToCash --withParam --withResponse
  ntaps create-handler --pkg=send --ucPkg=send --endpointType=private --endpoint=/transaction/:transaction_code --withParamUc --withResponseUc --ucMethodName=GetTransactionDetailByCode --method=getTransactionDetailByCode --tag=Send --verb=GET
  ntaps create-repository --type=postgres --pkg=user --method=UpdateUserStatus --withParamRepo --withResponseRepo --withTx --addToUC=send
  ntaps create-outbound --pkg=email --method=SendEmailActivation --withParam --withResp
  ntaps add-repo-to-usecase --repoPkg=example --ucPkg=send --method=GetExample --withParamRepo --withResponseRepo --withTx`)
	os.Exit(2)
}
