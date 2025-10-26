package repo

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/AndreeJait/ntaps/internal/paths"
	"github.com/AndreeJait/ntaps/internal/util"
)

func ensureRepoPkgPostgres(pkg string) error {
	mod := util.ModulePathGuess()
	dir := filepath.Join(paths.RepoPgPath, pkg)

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

	impl := filepath.Join(dir, "impl.go")
	if _, err := os.Stat(impl); os.IsNotExist(err) {
		body := fmt.Sprintf(`package %s

import (
	"context"

	"%[3]s/internal/adapters/outbound/db/postgres/sqlc"
	"github.com/AndreeJait/go-utility/tracer"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	q *sqlc.Queries
}

func New%[2]sRepository(q *pgxpool.Pool) *Repository {
	return &Repository{q: sqlc.New(q)}
}
`, pkg, util.ToPascalCase(pkg), mod)
		return util.WriteGoFile(impl, body)
	}

	return nil
}
