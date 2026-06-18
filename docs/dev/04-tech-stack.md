# 技术栈选型与依赖

本文档维护 GuanLan 的技术选型、版本要求、开发工具和硬件约束。架构职责见 [02-architecture.md](./02-architecture.md)。

## 1. 技术栈总览

| 类别 | 选型 | 版本要求 |
|------|------|----------|
| 前端 | React + TypeScript | 18+ |
| 构建工具 | Vite | 5+ |
| UI 组件库 | shadcn/ui | latest |
| 外部网关 | gRPC Gateway | latest |
| 接口服务 | Go + gRPC | 1.22+ |
| 日频数据与推理 | Python + Qlib + PyTorch | 3.12+ / 2.x |
| 训练与回测 | Python + Qlib + PyTorch | 3.12+ / 2.x |
| 数据库 | DuckDB | latest |
| 数据源 | yfinance + Qlib + 可扩展补充源 | latest |
| 包管理 | uv / Go modules / pnpm 或 npm | stable |
| 容器化 | Docker | 可选 |

## 2. 选型理由

### 2.1 Go + Python

- Go 负责 gRPC 接口服务、请求校验、任务编排和状态聚合，适合实现清晰稳定的本地服务边界。
- gRPC Gateway 作为外部网关，将前端 HTTP/JSON 请求转换为内部 gRPC 调用。
- Python 负责日频数据获取、特征工程、模型训练、回测和推理，便于复用 Qlib 与 PyTorch 生态。
- 模型优先在 Python 推理服务中运行；是否导出 ONNX 作为独立部署格式视后续性能和部署需求决定。

### 2.2 DuckDB

- 嵌入式、零配置，不需要独立数据库服务。
- 列式存储适合日频行情、特征和回测查询。
- 支持本地文件化部署，便于备份和迁移。

### 2.3 React + shadcn/ui

- React 和 TypeScript 适合构建可维护的单用户工作台。
- shadcn/ui 基于 Base UI 和 Tailwind CSS，组件可定制。
- Vite 构建快，适合分阶段开发。

### 2.4 Qlib

- 提供数据处理、模型训练和回测能力。
- 适合作为训练流程和量化工程设计参考。
- 与 PyTorch 结合后便于导出可部署模型。

## 3. 外部依赖

| 依赖 | 用途 | 备注 |
|------|------|------|
| yfinance | 美股数据获取 | Yahoo Finance 非官方 API |
| Qlib | A 股数据、训练和回测参考 | Microsoft 开源项目 |
| 可扩展补充源 | 数据冗余 | 具体源待决策 |

## 4. 开发环境

| 工具 | 用途 |
|------|------|
| uv | Python 依赖管理与虚拟环境 |
| Go modules | Go 接口服务依赖管理 |
| protoc / grpc-gateway | gRPC 与 HTTP 网关代码生成 |
| pnpm/npm | 前端依赖管理 |
| Docker | 可选容器化部署 |

## 5. 硬件约束

优先适配 Jetson Orin Nano Super 这类资源受限本地环境：

| 资源 | 规格 |
|------|------|
| CPU | 6 核 ARM Cortex-A78AE |
| 内存 | 8GB LPDDR5 |
| GPU | 512-core NVIDIA Ampere GPU |

实现影响：

1. Go 接口服务需要控制常驻内存占用，并避免承担重计算任务。
2. Python 日频数据、训练和推理任务宜独立进程运行，避免与 API 服务抢占资源。
3. DuckDB 需要配置合理内存上限。
4. Python 推理可评估 Jetson GPU 加速收益，必要时再评估 ONNX 导出。
