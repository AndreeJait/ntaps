package handler

// Run is the full flow: ensure usecase exists, ensure handler pkg, add method+route, wire DI.
import (
	"github.com/AndreeJait/ntaps/gen/usecase"
)

func Run(
	pkg string,
	ucPkg string,
	endpointType string,
	endpoint string,
	withParamUc bool,
	withResponseUc bool,
	ucMethodName string,
	handlerMethod string,
	tag string,
	verb string,
) error {
	// 1) make sure the usecase + method exist
	if err := usecase.Run(ucPkg, ucMethodName, withParamUc, withResponseUc); err != nil {
		return err
	}

	// 2) ensure handler pkg skeleton
	if err := ensurePackageOnly(pkg); err != nil {
		return err
	}

	// 3) ensure route + method in that handler
	if err := ensureMethodAndRoute(
		pkg,
		ucPkg,
		endpointType,
		endpoint,
		handlerMethod,
		ucMethodName,
		withParamUc,
		withResponseUc,
		tag,
		verb,
	); err != nil {
		return err
	}

	// 4) wire handler into infra
	if err := updateInfraHandlerInit(pkg); err != nil {
		return err
	}

	return nil
}

// EnsurePackageOnly creates handler package + registers it without adding new route.
// It's what we call for "skeleton mode".
func EnsurePackageOnly(pkg string) error {
	if err := ensurePackageOnly(pkg); err != nil {
		return err
	}
	if err := updateInfraHandlerInit(pkg); err != nil {
		return err
	}
	return nil
}
