# Logger - 通用日志工具包

基于 zap 的高性能日志库，支持 OpenTelemetry 集成和链路追踪。

## 特性

- 🚀 高性能日志输出（基于 uber/zap）
- 📝 支持控制台和文件输出
- 🔄 自动日志轮转
- 📊 OpenTelemetry 集成
- 🔍 链路追踪支持（Trace ID/Span ID）
- ⚙️ 灵活的配置选项

## 快速开始

### 基础使用

```go
package main

import (
    "go_sample_code/logger"
)

func main() {
    // 使用默认配置
    cfg := logger.DefaultConfig()
    log, err := logger.NewLogger(cfg)
    if err != nil {
        panic(err)
    }
    defer log.Shutdown(context.Background())
    
    log.Info("Hello World")
}
```

### 自定义配置

```go
cfg := &logger.LoggerConfig{
    Output:          "file",           // console 或 file
    Filepath:        "./logs",         // 日志文件路径
    Filename:        "app.log",        // 日志文件名
    Level:           "debug",          // debug, info, warn
    MaxBackups:      7,                // 保留日志文件数
    MaxAge:          30,               // 日志保留天数
    OtelEnable:      false,            // 是否启用 OpenTelemetry
    OtelServiceName: "my-service",     // 服务名称
}

log, _ := logger.NewLogger(cfg)
```

### 带上下文的日志

```go
import "context"

ctx := context.Background()
log.InfoCtx(ctx, "处理请求", zap.String("user", "alice"))
```

## 配置说明

| 字段 | 类型 | 说明 | 默认值 |
|------|------|------|--------|
| Output | string | 输出方式：console/file | console |
| Filepath | string | 文件输出路径 | - |
| Filename | string | 日志文件名 | - |
| Level | string | 日志级别：debug/info/warn | info |
| MaxBackups | int | 保留文件数 | 7 |
| MaxAge | int | 保留天数 | 30 |
| OtelEnable | bool | 启用 OpenTelemetry | false |
| OtelEndpoint | string | OTLP 端点 | - |
| OtelAuth | string | 认证信息 | - |
| OtelServiceName | string | 服务名称 | app |

## API 方法

```go
type Logger interface {
    Debug(msg string, fields ...zap.Field)
    DebugCtx(ctx context.Context, msg string, fields ...zap.Field)
    Info(msg string, fields ...zap.Field)
    InfoCtx(ctx context.Context, msg string, fields ...zap.Field)
    Warn(msg string, fields ...zap.Field)
    WarnCtx(ctx context.Context, msg string, fields ...zap.Field)
    Error(msg string, fields ...zap.Field)
    ErrorCtx(ctx context.Context, msg string, fields ...zap.Field)
    Panic(msg string, fields ...zap.Field)
    PanicCtx(ctx context.Context, msg string, fields ...zap.Field)
    GetLogger() *zap.Logger
    Shutdown(ctx context.Context)
}
```

## OpenTelemetry 集成

启用 OpenTelemetry 后，日志会自动包含 Trace ID 和 Span ID：

```go
cfg := &logger.LoggerConfig{
    OtelEnable:      true,
    OtelEndpoint:    "localhost:4317",
    OtelServiceName: "my-service",
}

log, _ := logger.NewLogger(cfg)
log.InfoCtx(ctx, "操作完成")
// 输出包含: trace_id, span_id, trace_flags
```

## 依赖

- `go.uber.org/zap` - 日志核心
- `go.opentelemetry.io/otel` - 可观测性
- `github.com/DeRuina/timberjack` - 日志轮转
