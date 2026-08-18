# 技术栈选型与依赖

本文档维护 GuanLan 的技术选型、版本要求、开发工具和硬件约束。架构职责见 [02-architecture.md](./02-architecture.md)。

## 1. 技术栈总览

| 类别 | 选型 | 版本要求 |
|------|------|----------|
| 前端 | React + TypeScript | 18+ |
| 构建工具 | Vite | 5+ |
| UI 组件库 | shadcn/ui | latest |
| HTTP | Kratos `transport/http` + `protoc-gen-go-http` | kratos v2 |
| OpenAPI | `protoc-gen-openapiv2` | 与 HTTP 注解共用 |
| 接口服务 | Go（唯一打开 DuckDB 的进程） | 见 `go.mod` |
| 日频爬虫 | Python crawler gRPC（无状态） | 3.12+ |
| 推理 | Python prediction gRPC（懒加载） | 3.12+ |
| 训练 | Python CLI，内存 DuckDB 读 Parquet | 3.12+ |
| 数据库 | DuckDB 单文件 | latest |
| 数据源 | yfinance + 可扩展补充源 | latest |
| DI / 定时 | uber-go/fx、robfig/cron | — |
| 包管理 | uv / Go modules / pnpm | stable |

## 2. 选型理由

### 2.1 Go + Python

- Go 负责 HTTP、校验、任务编排和 DuckDB 读写。
- Python 负责 yfinance 拉取、特征定义、训练与推理；**不打开** `data/guanlan.duckdb`。
- 训练与推理分进程，避免训练拖垮 serving；特征列在 `quant/ml/features` 共用以防 train/serve skew。
- 不用 `kratos.App` / Temporal；HTTP 只用 go-http + transport/http。

### 2.2 DuckDB

- 嵌入式、零配置；**整库单写者**，因此只由 `cmd/api` 打开。
- 列式存储适合日频与快照 `COPY TO PARQUET`。

### 2.3 React + shadcn/ui

- 单用户工作台；Vite 将 `/api` 代理到 `:8080`。

## 3. 外部依赖

| 依赖 | 用途 | 备注 |
|------|------|------|
| yfinance | 行情 | 非官方 API，无 SLA |
| grpcio / protobuf | Python gRPC | crawler + prediction |
| 可扩展补充源 | 数据冗余 | 见开放问题 |

## 4. 开发环境

| 工具 | 用途 |
|------|------|
| uv | Python 依赖 |
| Go modules + `go install tool` | 含 `protoc-gen-go-http`、air、goreman |
| buf | proto lint / generate |
| pnpm | 前端 |
| goreman | 本地三进程：api + crawler + prediction |

## 5. 硬件约束

优先适配 Jetson Orin Nano Super：

| 资源 | 规格 |
|------|------|
| CPU | 6 核 ARM Cortex-A78AE |
| 内存 | 8GB LPDDR5 |
| GPU | 512-core NVIDIA Ampere GPU |

实现影响：

1. Go 控制常驻内存；重计算放 Python。
2. crawler 约 150–200MB；prediction + torch **必须懒加载**（当前 baseline 为 JSON，不导入 torch）。
3. DuckDB 单进程写入。
