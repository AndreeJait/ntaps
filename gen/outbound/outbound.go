package outbound

import (
	"github.com/AndreeJait/ntaps/internal/util"
)

func Run(pkg, method string, withParam, withResp bool) error {
	// ensure base outbound pkg structure + port.go + impl.go
	if err := ensureOutboundPkg(pkg); err != nil {
		return err
	}

	// extend with method if provided
	if method != "" {
		if err := ensureOutboundPortHasMethod(pkg, method, withParam, withResp); err != nil {
			return err
		}
		if err := ensureOutboundDTO(pkg, method, withParam, withResp); err != nil {
			return err
		}
		if err := ensureOutboundImplHasMethod(pkg, method, withParam, withResp); err != nil {
			return err
		}
	}

	return nil
}

// helper for impl to get iface name
func ifaceName(pkg string) string {
	return util.ToPascalCase(pkg)
}
