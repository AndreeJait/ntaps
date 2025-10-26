package repo

import (
	"fmt"
	"os"
	"path/filepath"

	"strings"

	"github.com/AndreeJait/ntaps/internal/paths"
	"github.com/AndreeJait/ntaps/internal/util"
)

func ensureRepoMethod(pkg, method string, withParam, withResp, withTx bool) error {
	mod := util.ModulePathGuess()
	path := filepath.Join(paths.RepoPgPath, pkg, "impl.go")

	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	src := string(raw)

	// required imports
	reqImports := []string{
		`"context"`,
		fmt.Sprintf(`"%s/internal/adapters/outbound/db/postgres/sqlc"`, mod),
		`"github.com/AndreeJait/go-utility/tracer"`,
	}
	if withTx {
		reqImports = append(reqImports, `"github.com/jackc/pgx/v5"`)
	}
	for _, imp := range reqImports {
		src = util.InsertImport(src, imp)
	}

	// skip if method exists
	sig := fmt.Sprintf("func (r *Repository) %s(", method)
	if strings.Contains(src, sig) {
		return util.WriteGoFile(path, src)
	}

	// signature
	args := "ctx context.Context"
	if withParam {
		args += fmt.Sprintf(", param %sParam", method)
	}
	if withTx {
		args += ", tx pgx.Tx"
	}

	ret := "error"
	retBody := "\t// TODO: implement\n\t_ = q\n\treturn nil"
	if withResp {
		ret = fmt.Sprintf("(%sResponse, error)", method)
		retBody = fmt.Sprintf("\tvar resp %sResponse\n\t// TODO: implement\n\t_ = q\n\treturn resp, nil", method)
	}

	// q := r.q / tx override logic
	qSetup := "\n\tq := r.q\n"
	if withTx {
		qSetup = `
	q := r.q
	if tx != nil {
		q = sqlc.New(tx)
	}
`
	}

	methodCode := fmt.Sprintf(`

func (r *Repository) %s(%s) %s {
	span, ctx := tracer.StartSpan(ctx, tracer.GetFuncName(r.%s))
	defer span.End()
	%s
%s
}
`, method, args, ret, method, strings.TrimRight(qSetup, "\n"), retBody)

	src += methodCode
	return util.WriteGoFile(path, src)
}
