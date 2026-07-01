# AGENTS.md — GuanLan (观澜)

> 观水有术，必观其澜 — A single-user, local-first quantitative investment workstation.

## Project Identity

GuanLan is a local quantitative investment workstation (A-shares + US stocks). Three core domains: **market data** (OHLCV fetching/validation), **portfolio bookkeeping** (trades, dividends, P&L), and **quant tasks** (daily analysis, training, backtesting). Resource target: Jetson Orin Nano Super (6 ARM cores, 8 GB RAM).

Detailed docs: [docs/README.md](./docs/README.md) — start with [product scope](./docs/dev/01-product-scope.md), [architecture](./docs/dev/02-architecture.md), [domain model](./docs/dev/03-domain-model.md), [tech stack](./docs/dev/04-tech-stack.md).

## Architecture (Key Decisions)

```
React Frontend (:1420) ──HTTP/JSON──▶ Fiber Gateway (:8080) ──gRPC──▶ Go API Server (:50051)
                                                                              │
                                                                    ┌─────────┼─────────┐
                                                                    │ sql                 │ subprocess
                                                              DuckDB (file)     Python (uv run)
```

- **Go for API, Python for data/ML** — Go handles structured logic, validation, and orchestration. Python handles yfinance fetching, cleaning, validation, and ML workloads.
- **DuckDB is the only database** — Embedded, single-file at `data/guanlan.duckdb`. No separate DB server.
- **gRPC + gRPC-Gateway** — API defined in `proto/v1/api.proto`. `buf generate` produces Go stubs. Frontend only talks HTTP/JSON to the gateway.
- **uber-go/fx DI** — All Go wiring via `fx.Provide`/`fx.Invoke` in `cmd/api/main.go` and `cmd/gateway/main.go`.
- **Python as subprocess** — Go scheduler calls Python via `uv run python -m services.daily_data`. No gRPC bridge between the two (yet).
- **Task model is universal** — All operations (sync, analysis, training, backtest) are tasks with states: `pending → running → success/failed/partial_success/cancelled`.

## Quick Commands

### Go Backend

```bash
go mod tidy              # Resolve dependencies
go install tool          # Install protoc plugins, air, goreman
buf generate             # Regenerate protobuf stubs
go run ./cmd/api/        # gRPC server :50051
go run ./cmd/gateway/    # HTTP gateway :8080
goreman start            # Both + hot-reload (via air)
go test ./...            # Run Go tests
```

### Python Services

```bash
uv sync                                                       # Install Python deps
uv run python -m services.daily_data sync --stock 600519.SS --db data/guanlan.duckdb
uv run python -m services.daily_data daily-sync --db data/guanlan.duckdb
uv run python -m services.daily_data run-scheduler --db data/guanlan.duckdb
python scripts/import_stock_pool_csv.py --csv data/000510cons.csv --db data/guanlan.duckdb
```

### Web Frontend (`web/`)

```bash
pnpm dev          # Vite :1420, proxies /api → :8080
pnpm build        # tsc -b && vite build
pnpm lint         # ESLint
pnpm typecheck    # tsc --noEmit
```

## Conventions

| Area | Convention |
|------|-----------|
| **Go** | Idiomatic Go; compile-time interface checks (`var _ Interface = (*Impl)(nil)`); error wrapping with `%w`; `Repository` interface + `Store` impl pattern |
| **Python** | `from __future__ import annotations` in all files; `argparse` CLI with subcommands; `services.<domain>` module hierarchy |
| **Protobuf** | Package `guanlan.v1`; Go import path `github.com/gsxhnd/guanlan/internal/proto/v1;v1`; HTTP annotations via Google API HTTP pattern |
| **Frontend** | `@/` → `web/src/`; `kebab-case` filenames; Zustand stores; shadcn/ui components in `src/components/ui/`; i18n keys as `dot.notation` |
| **DuckDB** | Snake_case table names; `CREATE TABLE IF NOT EXISTS` DDL in both Go (`internal/data/schema.go`) and Python (`services/daily_data/schema.sql`) |
| **Docs** | Chinese; numbered prefixes; Mermaid diagrams |

## ⚠️ Pitfalls

1. **Dual schema definitions** — `internal/data/schema.go` and `services/daily_data/schema.sql` both define `stock_pool` and `daily_bars*`. Changes must be sync'd across both. The Python side checks for legacy column names and drops/recreates.
2. **DuckDB single-writer** — Both Go and Python open `data/guanlan.duckdb`. Only one writer at a time; the WAL file (`.duckdb.wal`) confirms active logging.
3. **No migration framework** — Schema evolves via `CREATE TABLE IF NOT EXISTS`. No versioned migrations. Be careful with ALTER/DROP operations.
4. **yfinance is unofficial** — No SLA, rate limits possible, breaking changes happen. See [open questions](./docs/dev/06-open-questions.md).
5. **Go tool chain** — `go.mod` uses `go 1.26.2` and `tool()` directive (requires Go ≥1.24). Protobuf generation uses local plugins (installed via `go install tool`).
6. **Frontend requires gateway** — Vite proxies `/api` to `:8080`. The gateway must be running for API calls to work.
