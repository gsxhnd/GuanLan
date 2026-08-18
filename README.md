# GuanLan

观澜：语出《孟子》“观水有术，必观其澜”。本地优先的量化投资工作台（A 股 + 美股）。

详细说明：[docs/README.md](./docs/README.md)、[AGENTS.md](./AGENTS.md)。

## 开发启动

```bash
uv sync
go mod tidy
go install tool
buf generate

# 可选：写出 baseline 模型，prediction 才不会返回 UNIMPLEMENTED
uv run python -m quant.ml.training --model baseline

goreman start   # api :8080 + crawler :50061 + prediction :50062
# 前端
cd web && pnpm dev   # :1420，代理 /api → :8080
```

## 进程

| 进程 | 地址 | 说明 |
|------|------|------|
| `cmd/api` | `:8080` | HTTP + DuckDB + 编排 + cron |
| `quant.crawler` | `:50061` | 无状态行情 gRPC |
| `quant.ml.serving` | `:50062` | 懒加载推理 gRPC |
