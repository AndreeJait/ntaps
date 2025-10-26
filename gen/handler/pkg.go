package handler

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/AndreeJait/ntaps/internal/paths"
	"github.com/AndreeJait/ntaps/internal/util"
)

func ensurePackageOnly(pkg string) error {
	mod := util.ModulePathGuess()

	dir := filepath.Join(paths.HandlerRootHTTPDir, pkg)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

	path := filepath.Join(dir, paths.HandlerPkgFileName)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		body := fmt.Sprintf(`package %s

import (
	http "%[2]s/internal/adapters/inbound/http"
	"%[2]s/internal/adapters/inbound/http/common/middleware"
	"%[2]s/internal/infrastructure/config"
	"%[2]s/internal/usecase"

	"github.com/AndreeJait/go-utility/response"
	"github.com/AndreeJait/go-utility/tracer"
	"github.com/labstack/echo/v4"
)

type handler struct {
	route *echo.Group
	uc    *usecase.UseCase
	cfg   *config.Config
}

func New%[3]sHandler(cfg *config.Config, route *echo.Group, uc *usecase.UseCase) http.Handler {
	return &handler{cfg: cfg, route: route, uc: uc}
}

// Handle registers routes for this module.
func (h *handler) Handle() {
	groupPublic := h.route.Group("/%[1]s")
	groupInternal := h.route.Group("/internal/%[1]s")
	groupPrivate := h.route.Group("/%[1]s")

	groupInternal.Use(middleware.BasicAuthLogged(h.cfg))
	groupPrivate.Use(middleware.MustLogged(h.cfg))

	// ntaps:routes
}
`, pkg, mod, util.ToPascalCase(pkg))
		return util.WriteGoFile(path, body)
	}

	return nil
}
