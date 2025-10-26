package repo

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/AndreeJait/ntaps/gen/usecase"
	"github.com/AndreeJait/ntaps/internal/paths"
)

func Run(pkg, method string, withParamRepo, withRespRepo, withTx bool, addToUC string) error {
	// 1. ensure repo package
	if err := ensureRepoPkgPostgres(pkg); err != nil {
		return err
	}

	// 2. DTO + method
	if err := ensureRepoDTO(pkg, method, withParamRepo, withRespRepo); err != nil {
		return err
	}
	if err := ensureRepoMethod(pkg, method, withParamRepo, withRespRepo, withTx); err != nil {
		return err
	}

	// 3. update postgres di.go
	if err := updatePostgresDI(pkg); err != nil {
		return err
	}

	// 4. wire initRepository() infra
	if err := updateInfraRepositoryInit(pkg); err != nil {
		return err
	}

	// 5. optionally wire into usecase (repo interface, struct field, ctor, init args)
	if addToUC != "" {
		ucDir := filepath.Join(paths.RootUsecaseDir, addToUC)
		if _, err := os.Stat(ucDir); errors.Is(err, os.ErrNotExist) {
			fmt.Println("ℹ️  repo not added to usecase because usecase is not found:", addToUC)
			return nil
		}
		// ensure a useCase exists at least, but don't create a new method
		if err := usecase.Run(addToUC, "", withParamRepo, withRespRepo); err != nil {
			// we try but don't die if adding fails
			fmt.Println("ℹ️  could not ensure usecase before wiring repo:", err)
		}

		if err := ensureUsecaseRepoInterface(addToUC, pkg, method, withParamRepo, withRespRepo, withTx); err != nil {
			return err
		}
		if err := ensureUsecaseHasRepo(addToUC, pkg); err != nil {
			return err
		}
		if err := updateInfraUsecaseInitArgs(addToUC, pkg); err != nil {
			return err
		}
	}

	return nil
}

// AddRepoToUsecase is the new feature.
// It wires an existing repo (repoPkg) into an existing usecase (ucPkg),
// and ensures the repo method signature is visible to that usecase.
func AddRepoToUsecase(repoPkg, ucPkg, method string, withParamRepo, withRespRepo, withTx bool) error {
	// Sanity: does the ucPkg actually exist?
	ucDir := filepath.Join(paths.RootUsecaseDir, ucPkg)
	if _, err := os.Stat(ucDir); errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("usecase package %q not found under %s", ucPkg, ucDir)
	}

	// 🔍 auto-detect repo signature if flags are all false
	if !withParamRepo && !withRespRepo && !withTx {
		dParam, dResp, dTx := detectRepoMethodSignature(repoPkg, method)
		withParamRepo, withRespRepo, withTx = dParam, dResp, dTx
		fmt.Printf("🔎 detected from %s.%s -> param=%v, resp=%v, tx=%v\n", repoPkg, method, withParamRepo, withRespRepo, withTx)
	}

	// 1) make sure repo package + method exist (idempotent)
	if err := ensureRepoPkgPostgres(repoPkg); err != nil {
		return err
	}
	if err := ensureRepoDTO(repoPkg, method, withParamRepo, withRespRepo); err != nil {
		return err
	}
	if err := ensureRepoMethod(repoPkg, method, withParamRepo, withRespRepo, withTx); err != nil {
		return err
	}

	// 2) make sure the repo is represented in postgres di and infra repo init
	if err := updatePostgresDI(repoPkg); err != nil {
		return err
	}
	if err := updateInfraRepositoryInit(repoPkg); err != nil {
		return err
	}

	// 3) Wire that repo into the given usecase
	// (a) make sure the UC exists and is "known" to ntaps so we don't miss imports, etc.
	// We call usecase.Run with dummy method info just to ensure port.go/usecase.go struct exist,
	// but *IMPORTANT*: we must NOT create a new UC method if the user didn't ask for it.
	//
	// So we do NOT call usecase.Run here, because that would try to add a new <method> to the usecase.
	// Instead we just assume the ucPkg's files are already in place.
	//
	// If you want auto-creation of ucPkg if missing, uncomment this:
	//    _ = usecase.Run(ucPkg, method, false, false)
	// but that would also add a method to that UC, which we don't want in this flow.

	if err := ensureUsecaseRepoInterface(ucPkg, repoPkg, method, withParamRepo, withRespRepo, withTx); err != nil {
		return err
	}

	if err := ensureUsecaseHasRepo(ucPkg, repoPkg); err != nil {
		return err
	}

	if err := updateInfraUsecaseInitArgs(ucPkg, repoPkg); err != nil {
		return err
	}

	// 4) ensure the DI layer knows about this usecase type at all.
	// (If the UC was never wired before, we need updateUsecaseDI + updateInfraInitUsecase once.)
	// BUT we must be careful not to inject a fresh duplicate s.uc.<Uc>Uc=... line.
	//
	// We'll reuse usecase.DI helpers, which now are safe (because we fixed updateInfraInitUsecase to not duplicate).
	if err := usecase.EnsureUsecaseKnownToDI(ucPkg); err != nil {
		return err
	}

	return nil
}

// detectRepoMethodSignature scans the repo impl.go for the method and infers
// whether it has param, response, or tx.
func detectRepoMethodSignature(repoPkg, method string) (hasParam, hasResp, hasTx bool) {
	filePath := filepath.Join(paths.RepoPgPath, repoPkg, "impl.go")

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filePath, nil, 0)
	if err != nil {
		// fallback: couldn't parse, return all false
		return false, false, false
	}

	// Look for func (r *repository) <Method>(
	for _, decl := range f.Decls {
		fd, ok := decl.(*ast.FuncDecl)
		if !ok || fd.Recv == nil || fd.Name.Name != method {
			continue
		}

		// Check params
		for _, field := range fd.Type.Params.List {
			typ := exprToString(field.Type)
			if strings.Contains(typ, "pgx.Tx") {
				hasTx = true
			}
			if strings.Contains(typ, method+"Param") {
				hasParam = true
			}
		}

		// Check results
		if fd.Type.Results != nil {
			for _, field := range fd.Type.Results.List {
				typ := exprToString(field.Type)
				if strings.Contains(typ, method+"Response") {
					hasResp = true
				}
			}
		}
		break
	}

	return hasParam, hasResp, hasTx
}

// exprToString converts a Go AST expression to a simple string
func exprToString(e ast.Expr) string {
	switch v := e.(type) {
	case *ast.Ident:
		return v.Name
	case *ast.SelectorExpr:
		return exprToString(v.X) + "." + v.Sel.Name
	case *ast.StarExpr:
		return exprToString(v.X)
	case *ast.ArrayType:
		return "[]" + exprToString(v.Elt)
	case *ast.MapType:
		return "map[" + exprToString(v.Key) + "]" + exprToString(v.Value)
	case *ast.InterfaceType:
		return "interface{}"
	default:
		return ""
	}
}
