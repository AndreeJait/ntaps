package usecase

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/AndreeJait/ntaps/internal/paths"
)

// Run creates or extends a usecase package and wires DI.
func Run(pkg, method string, withParam, withResp bool) error {
	pkgDir := filepath.Join(paths.RootUsecaseDir, pkg)

	if _, err := os.Stat(pkgDir); errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(pkgDir, 0o755); err != nil {
			return fmt.Errorf("mkdir %s: %w", pkgDir, err)
		}
		if err := createPort(pkgDir, pkg, method, withParam, withResp); err != nil {
			return err
		}
		if err := createDTO(pkgDir, pkg, method, withParam, withResp); err != nil {
			return err
		}
		if err := createImpl(pkgDir, pkg, method, withParam, withResp); err != nil {
			return err
		}
	} else {
		if err := ensurePort(pkgDir, pkg, method, withParam, withResp); err != nil {
			return err
		}
		if err := ensureDTO(pkgDir, pkg, method, withParam, withResp); err != nil {
			return err
		}
		if err := ensureImpl(pkgDir, pkg, method, withParam, withResp); err != nil {
			return err
		}
	}

	if err := updateUsecaseDI(pkg); err != nil {
		return err
	}
	if err := updateInfraInitUsecase(pkg); err != nil {
		return err
	}

	return nil
}
