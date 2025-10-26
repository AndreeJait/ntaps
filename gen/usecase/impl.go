package usecase

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AndreeJait/ntaps/internal/util"
)

func createImpl(dir, pkg, method string, withParam, withResp bool) error {
	path := filepath.Join(dir, "usecase.go")
	body := renderImpl(pkg, method, withParam, withResp)
	return util.WriteGoFile(path, body)
}

func ensureImpl(dir, pkg, method string, withParam, withResp bool) error {
	path := filepath.Join(dir, "usecase.go")

	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return createImpl(dir, pkg, method, withParam, withResp)
	}
	if err != nil {
		return err
	}

	src := string(b)
	mod := util.ModulePathGuess()

	// required imports
	reqImports := []string{
		fmt.Sprintf(`"%s/internal/infrastructure/config"`, mod),
		fmt.Sprintf(`"%s/internal/infrastructure/db"`, mod),
		`"github.com/AndreeJait/go-utility/loggerw"`,
		`"github.com/AndreeJait/go-utility/tracer"`,
		`"context"`,
		`"time"`,
	}
	for _, imp := range reqImports {
		src = util.InsertImport(src, imp)
	}

	// ensure useCase struct & NewUseCase signature
	if !strings.Contains(src, "type useCase struct") {
		src += `
type useCase struct {
	cfg       *config.Config
	log       loggerw.Logger
	txManager *db.TxManager
}

func NewUseCase(cfg *config.Config, log loggerw.Logger, txManager *db.TxManager) UseCase {
	return &useCase{cfg: cfg, log: log, txManager: txManager}
}
`
	} else {
		// normalize txManager *db.TxManager
		src = strings.ReplaceAll(src, "txManager  db.TxManager", "txManager *db.TxManager")
		src = strings.ReplaceAll(src, "txManager db.TxManager", "txManager *db.TxManager")
		src = strings.ReplaceAll(
			src,
			"NewUseCase(cfg *config.Config, log loggerw.Logger, txManager db.TxManager)",
			"NewUseCase(cfg *config.Config, log loggerw.Logger, txManager *db.TxManager)",
		)
	}

	// ensure method exists
	methodSig := fmt.Sprintf("func (u *useCase) %s(", method)
	if !strings.Contains(src, methodSig) {
		sigIn := "ctx context.Context"
		if withParam {
			sigIn += ", req " + method + "Request"
		}
		ret := "error"
		var retBody string
		if withResp {
			ret = "(" + method + "Response, error)"
			retBody = "var resp " + method + "Response\n\t// TODO: implement\n\treturn resp, nil"
		} else {
			retBody = "// TODO: implement\n\treturn nil"
		}

		src += fmt.Sprintf(`

func (u *useCase) %s(%s) %s {
	span, ctx := tracer.StartSpan(ctx, tracer.GetFuncName(u.%s))
	defer span.End()

	%s
}
`, method, sigIn, ret, method, retBody)
	}

	return util.WriteGoFile(path, src)
}

func renderImpl(pkg, method string, withParam, withResp bool) string {
	mp := util.ModulePathGuess()

	sigIn := "ctx context.Context"
	if withParam {
		sigIn += ", req " + method + "Request"
	}
	ret := "error"
	retBody := "// TODO: implement\n\treturn nil"
	if withResp {
		ret = "(" + method + "Response, error)"
		retBody = "var resp " + method + "Response\n\t// TODO: implement\n\treturn resp, nil"
	}

	return fmt.Sprintf(`package %s

import (
	"context"
	"time"

	"github.com/AndreeJait/go-utility/loggerw"
	"github.com/AndreeJait/go-utility/tracer"

	"%s/internal/infrastructure/config"
	"%s/internal/infrastructure/db"
)

// useCase implements UseCase.
type useCase struct {
	cfg       *config.Config
	log       loggerw.Logger
	txManager *db.TxManager
}

func NewUseCase(cfg *config.Config, log loggerw.Logger, txManager *db.TxManager) UseCase {
	return &useCase{cfg: cfg, log: log, txManager: txManager}
}

func (u *useCase) %s(%s) %s {
	span, ctx := tracer.StartSpan(ctx, tracer.GetFuncName(u.%s))
	defer span.End()

	_ = time.Second // keep import 'time' warm; remove when used

	%s
}
`, pkg, mp, mp, method, sigIn, ret, method, retBody)
}

// ---- wiring ----
// these map your original updateUsecaseDI, updateInfraInitUsecase, updateInfraUsecaseInitArgs logic
