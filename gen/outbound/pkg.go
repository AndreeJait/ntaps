package outbound

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/AndreeJait/ntaps/internal/paths"
	"github.com/AndreeJait/ntaps/internal/util"
)

func ensureOutboundPkg(pkg string) error {
	mod := util.ModulePathGuess()
	dir := filepath.Join(paths.OutboundRootPath, pkg)

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

	// port.go
	if err := ensureOutboundPort(pkg); err != nil {
		return err
	}

	// impl.go
	implPath := filepath.Join(dir, "impl.go")
	if _, err := os.Stat(implPath); os.IsNotExist(err) {
		body := fmt.Sprintf(`package %s

import (
	"%s/internal/infrastructure/config"
	"github.com/AndreeJait/go-utility/loggerw"
)

type impl struct {
	logger loggerw.Logger
	cfg    *config.Config
}

func New%[2]sW(log loggerw.Logger, cfg *config.Config) %[2]s {
	return &impl{logger: log, cfg: cfg}
}
`, pkg, ifaceName(pkg), mod)
		return util.WriteGoFile(implPath, body)
	}
	return nil
}

func ensureOutboundPort(pkg string) error {
	path := filepath.Join(paths.OutboundRootPath, pkg, "port.go")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		body := fmt.Sprintf(`package %s

type %s interface {
}
`, pkg, ifaceName(pkg))
		return util.WriteGoFile(path, body)
	}
	return nil
}
