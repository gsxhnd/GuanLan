# AGENTS.md — GuanLan (观澜)

> 观水有术，必观其澜 — A single-user, local-first quantitative investment workstation.

## Project Identity

GuanLan is a local quantitative investment workstation (A-shares + US stocks). Three core domains: **market data** (OHLCV fetching/validation), **portfolio bookkeeping** (trades, dividends, P&L), and **quant tasks** (daily analysis, training, backtesting). Resource target: Jetson Orin Nano Super (6 ARM cores, 8 GB RAM).

Detailed docs: [docs/README.md](./docs/README.md) — start with [product scope](./docs/dev/01-product-scope.md), [architecture](./docs/dev/02-architecture.md), [domain model](./docs/dev/03-domain-model.md), [tech stack](./docs/dev/04-tech-stack.md).

```
React Frontend (:1420) ──HTTP/JSON──▶ cmd/api (:8080) ──DuckDB──▶ data/guanlan.duckdb
 │
 └──gRPC──▶ quant.crawler (:50061)
 └──gRPC──▶ quant.ml.serving (:50062)
```

- **Go for API, Python for data/ML** — Go handles structured logic, validation, and orchestration. Python crawler / prediction are **stateless** (no DuckDB).
- **DuckDB is the only database** — Embedded, single-file at `data/guanlan.duckdb`. Only `cmd/api` opens it.
- **HTTP via Kratos go-http** — API defined in `proto/quant/v1/api.proto` (`quant.v1`). Frontend talks HTTP/JSON to `:8080`.
- **uber-go/fx DI** — Wiring via `fx.Provide`/`fx.Invoke` in `cmd/api/main.go` (no `kratos.App`).
- **Orchestrator + cron** — `internal/orchestrator` claims tasks and calls crawler gRPC; `robfig/cron` enqueues daily sync.
- **Task model is universal** — All operations (sync, analysis, training, backtest) are tasks with states: `pending → running → success/failed/partial_success/cancelled`.

## Quick Commands

### Go Backend

```bash
go mod tidy              # Resolve dependencies
go install tool          # Install protoc plugins, air, goreman
buf generate             # Regenerate protobuf stubs (Go + Python + OpenAPI)
go run ./cmd/api/        # HTTP + DuckDB + orchestrator + cron :8080
goreman start            # api + crawler + prediction
go test ./...            # Run Go tests
```

### Python Services

```bash
uv sync                                                       # Install Python deps
uv run python -m quant.crawler.server --addr :50061           # crawler gRPC
uv run python -m quant.crawler.cli fetch --stock 600519.SS    # dry fetch (no DB)
uv run python -m quant.ml.training --model baseline           # write data/models/baseline.json
uv run python -m quant.ml.serving --addr :50062               # prediction gRPC
uv run python -m quant.tools.pool_csv --csv data/000510cons.csv
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
| **Python** | `from __future__ import annotations` in all files; `argparse` CLI with subcommands; top-level package `quant/` |
| **Protobuf** | Package `quant.v1`; Go import path `github.com/gsxhnd/guanlan/internal/proto/quant/v1;v1`; HTTP annotations via Google API HTTP pattern |
| **Frontend** | `@/` → `web/src/`; `kebab-case` filenames; Zustand stores; shadcn/ui components in `src/components/ui/`; i18n keys as `dot.notation` |
| **DuckDB** | Snake_case table names; `CREATE TABLE IF NOT EXISTS` DDL only in Go (`internal/data/schema.go`) |
| **Docs** | Chinese; numbered prefixes; Mermaid diagrams |

## ⚠️ Pitfalls

1. **Single DuckDB writer** — Only `cmd/api` opens `data/guanlan.duckdb`. Crawler must stay stateless.
2. **No migration framework** — Schema evolves via `CREATE TABLE IF NOT EXISTS` (+ occasional `DROP` in `schema.go`). Be careful with ALTER/DROP.
3. **yfinance is unofficial** — No SLA, rate limits possible, breaking changes happen. See [open questions](./docs/dev/06-open-questions.md).
4. **Go tool chain** — `go.mod` uses `go 1.26.2` and `tool()` directive (requires Go ≥1.24). Includes `protoc-gen-go-http`.
5. **Frontend requires api + crawler for sync** — Vite proxies `/api` to `:8080`. Sync tasks need crawler on `:50061`.
6. **Roadmap** — 产品与平台阶段见 [docs/dev/05-roadmap.md](./docs/dev/05-roadmap.md)；架构见 [docs/dev/02-architecture.md](./docs/dev/02-architecture.md)。
