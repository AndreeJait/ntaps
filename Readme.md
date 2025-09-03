# ntaps

Scaffold Go service boilerplate the fast way: usecases, handlers (Echo), repositories (Postgres/sqlc), and outbound adapters — all wired into your DI.

- Generates clean, formatted Go code (runs `goimports` under the hood).
- Idempotent: re-runs safely append to interfaces/impls and avoid dupes.
- **Interactive mode** for user-friendly prompts (or use flags if you prefer).

---

## 🚀 Install

Requires **Go 1.20+**

```bash
go install github.com/AndreeJait/ntaps@latest
# make sure $(go env GOPATH)/bin is on your PATH
```

---

## 📂 Project Layout

`ntaps` expects a hexagonal service layout. It auto-detects the module path from the first line of `go.mod`.

Expected directories (auto-created if missing):

```
internal/
  usecase/
  adapters/
    inbound/
      http/
    outbound/
      db/
        postgres/
      <other outbound packages…>
  infrastructure/
    di/
      handler.go
      repository.go
    config/
    db/
```

👉 Start from the template repo:  
[go-template-hexagonal](https://github.com/AndreeJait/go-template-hexagonal)

---

## 🌐 Global Conventions

- **Tracing**: boilerplate added to every generated function.
  ```go
  span, ctx := tracer.StartSpan(ctx, tracer.GetFuncName(<receiver>.<MethodName>))
  defer span.End()
  ```
- **Formatting & imports**: runs `goimports` + fallback `gofmt`.
- **Module detection**: reads `go.mod`, falls back to defaults if missing.

---

## 🛠 Commands

### 1) `create-usecase`

Scaffolds/extends a usecase.

```bash
ntaps create-usecase --pkg=send --method=SubmitCashToCash --withParam --withResponse
```

Generates:

- `internal/usecase/<pkg>/port.go`
- `internal/usecase/<pkg>/dto.go`
- `internal/usecase/<pkg>/usecase.go`
- DI wiring in `internal/usecase/di.go` + `internal/infrastructure/di/usecase.go`

✅ Idempotent: re-runs append safely.

---

### 2) `create-handler` (Echo)

Interactive (default):

```bash
ntaps create-handler
# prompts: pkg, ucPkg, endpointType, endpoint, verb, etc.
```

With flags:

```bash
ntaps create-handler \
  --pkg=send \
  --ucPkg=send \
  --endpointType=private \
  --endpoint=/submit/cash-to-cash \
  --withParamUc \
  --withResponseUc \
  --ucMethodName=SubmitCashToCash \
  --method=submitCashToCash \
  --tag=Send \
  --verb=POST
```

---

### 3) `create-repository` (Postgres/sqlc)

```bash
ntaps create-repository \
  --type=postgres \
  --pkg=user \
  --method=UpdateUserStatus \
  --withParamRepo \
  --withResponseRepo \
  --withTx \
  --addToUC=send
```

Generates:

- `internal/adapters/outbound/db/postgres/<pkg>/impl.go`
- `internal/adapters/outbound/db/postgres/<pkg>/dto.go`
- DI wiring in `internal/adapters/outbound/db/postgres/di.go`
- `internal/infrastructure/di/repository.go`

Optionally wires repo → usecase if `--addToUC` is given.

---

### 4) `create-outbound` (generic outbound adapter)

Interactive:

```bash
ntaps create-outbound
```

With flags:

```bash
ntaps create-outbound --pkg=email --method=SendEmailActivation --withParam --withResp
```

Generates:

- `internal/adapters/outbound/<pkg>/port.go`
- `internal/adapters/outbound/<pkg>/impl.go`
- `internal/adapters/outbound/<pkg>/dto.go`

---

## 💡 Interactive Mode Tips

- Running without flags starts prompts.
- `Enter` keeps defaults/skips.
- `create-handler` with only `--pkg` → skeleton handler + DI wiring, routes later.
- Force prompts with:
  ```bash
  NTAPS_INTERACTIVE=1 ntaps create-handler
  ```

---

## 🔖 Versioning & Release Workflow

This repo includes GitHub Actions for **version tagging**:

- Patch bumps:
  ```
  v.0.0.1 → v.0.0.2 → ... → v.0.0.9 → v.0.1.0
  ```
- You can bump patch/minor/major via workflow dispatch.

The workflow:

1. Creates an annotated tag & pushes it
2. Moves `latest` tag to same commit
3. Creates a GitHub Release

Install a version:

```bash
go install github.com/AndreeJait/ntaps@v.0.1.0
```

Install the latest:

```bash
go install github.com/AndreeJait/ntaps@latest
```

---

## 🐞 Troubleshooting

**“initUseCase/initRepository/handlers slice not found”**  
→ Ensure DI files exist with expected names:

- `internal/infrastructure/di/usecase.go` → `func (s wire) initUseCase(...)`
- `internal/infrastructure/di/repository.go` → `func (s wire) initRepository(...)`
- `internal/infrastructure/di/handler.go` → `var handlers = []http.Handler{...}`

**Module path looks wrong**  
→ Check first line of `go.mod`: `module <path>`

**Imports/formatting**  
→ Run manually:
```bash
gofmt -w . && goimports -w .
```

---

## 📦 Template Use

Best to start from:  
[go-template-hexagonal](https://github.com/AndreeJait/go-template-hexagonal)

---

## 📜 License

[MIT](./LICENSE) © AndreeJait
