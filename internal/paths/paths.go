package paths

const (
	RootUsecaseDir       = "internal/usecase"
	UsecaseDIPath        = "internal/usecase/di.go"
	InfraInitUsecasePath = "internal/infrastructure/di/usecase.go"

	HandlerRootHTTPDir   = "internal/adapters/inbound/http"
	HandlerPkgFileName   = "di.go"
	HandlerInfraInitPath = "internal/infrastructure/di/handler.go"

	RepoRootPath      = "internal/adapters/outbound/db"
	RepoPgPath        = "internal/adapters/outbound/db/postgres"
	PgDiPath          = "internal/adapters/outbound/db/di.go"
	InfraRepoInitPath = "internal/infrastructure/di/repository.go"

	OutboundRootPath = "internal/adapters/outbound"
)
