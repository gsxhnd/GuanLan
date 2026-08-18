package trace

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	otel_trace "go.opentelemetry.io/otel/trace"
)

// TraceConfig OpenTelemetry 配置
type TraceConfig struct {
	OtelEnable         bool              `yaml:"otel_enable"`          // 是否启用 OTLP 导出
	OtelEndpoint       string            `yaml:"otel_endpoint"`        // OTLP 端点
	OtelAuth           string            `yaml:"otel_auth"`            // 认证 token
	OtelServiceName    string            `yaml:"otel_service_name"`    // 服务名称
	OtelServiceVersion string            `yaml:"otel_service_version"` // 服务版本
	OtelEnvironment    string            `yaml:"otel_environment"`     // 部署环境
	OtelOrganization   string            `yaml:"otel_organization"`    // 组织名称
	OtelStreamName     string            `yaml:"otel_stream_name"`     // 数据流名称
	OtelSamplerType    string            `yaml:"otel_sampler_type"`    // 采样器类型: always_on, always_off, trace_id_ratio
	OtelSamplerRatio   float64           `yaml:"otel_sampler_ratio"`   // 采样比例 (0.0-1.0)
	OtelHeaders        map[string]string `yaml:"otel_headers"`         // 自定义 Headers
}

// Validate 验证配置有效性
func (c *TraceConfig) Validate() error {
	if c.OtelEnable {
		if c.OtelEndpoint == "" {
			return fmt.Errorf("otel_endpoint is required when otel_enable is true")
		}
	}
	if c.OtelServiceName == "" {
		c.OtelServiceName = "unknown-service"
	}
	if c.OtelSamplerType == "" {
		c.OtelSamplerType = "always_on"
	}
	if c.OtelSamplerRatio <= 0 || c.OtelSamplerRatio > 1 {
		c.OtelSamplerRatio = 1.0
	}
	return nil
}

// newSampler 根据配置创建采样器
func (c *TraceConfig) newSampler() sdktrace.Sampler {
	switch c.OtelSamplerType {
	case "always_off":
		return sdktrace.NeverSample()
	case "trace_id_ratio":
		return sdktrace.TraceIDRatioBased(c.OtelSamplerRatio)
	default: // always_on
		return sdktrace.AlwaysSample()
	}
}

// newResource 创建资源定义
func (c *TraceConfig) newResource(ctx context.Context) (*resource.Resource, error) {
	attrs := []attribute.KeyValue{
		semconv.ServiceName(c.OtelServiceName),
		semconv.ServiceVersion(c.OtelServiceVersion),
		semconv.DeploymentEnvironment(c.OtelEnvironment),
	}
	return resource.New(ctx,
		resource.WithAttributes(attrs...),
		resource.WithTelemetrySDK(),
		resource.WithProcess(),
	)
}

func NewTracerProvider(cfg *TraceConfig) (*sdktrace.TracerProvider, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("trace config validation failed: %w", err)
	}

	ctx := context.Background()
	r, err := cfg.newResource(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	opts := []sdktrace.TracerProviderOption{
		sdktrace.WithResource(r),
		sdktrace.WithSampler(cfg.newSampler()),
	}

	if cfg.OtelEnable {
		exp, err := cfg.newExporter(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
		}
		opts = append(opts, sdktrace.WithBatcher(exp,
			sdktrace.WithBatchTimeout(5*time.Second),
			sdktrace.WithMaxExportBatchSize(512),
		))
	}

	return sdktrace.NewTracerProvider(opts...), nil
}

// newExporter 创建 OTLP 导出器
func (cfg *TraceConfig) newExporter(ctx context.Context) (sdktrace.SpanExporter, error) {
	endpoint := cfg.OtelEndpoint

	headers := map[string]string{
		"Authorization": cfg.OtelAuth,
	}
	// 添加自定义 headers
	for k, v := range cfg.OtelHeaders {
		headers[k] = v
	}
	// 默认值可被自定义覆盖
	if _, ok := headers["organization"]; !ok && cfg.OtelOrganization != "" {
		headers["organization"] = cfg.OtelOrganization
	}
	if _, ok := headers["stream-name"]; !ok && cfg.OtelStreamName != "" {
		headers["stream-name"] = cfg.OtelStreamName
	}

	return otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithHeaders(headers),
		otlptracegrpc.WithCompressor("gzip"),
		otlptracegrpc.WithTimeout(5*time.Second),
	)
}

// NewTracer 创建 Tracer 实例
func NewTracer(tp *sdktrace.TracerProvider) otel_trace.Tracer {
	return tp.Tracer("")
}

// GetTraceID 从 context 中提取 trace_id
func GetTraceID(ctx context.Context) string {
	spanCtx := otel_trace.SpanContextFromContext(ctx)
	if spanCtx.IsValid() {
		return spanCtx.TraceID().String()
	}
	return ""
}

// GetSpanID 从 context 中提取 span_id
func GetSpanID(ctx context.Context) string {
	spanCtx := otel_trace.SpanContextFromContext(ctx)
	if spanCtx.IsValid() {
		return spanCtx.SpanID().String()
	}
	return ""
}

// SpanContextToMap 将 trace 信息提取为 map（用于日志）
func SpanContextToMap(ctx context.Context) map[string]string {
	return map[string]string{
		"trace_id": GetTraceID(ctx),
		"span_id":  GetSpanID(ctx),
	}
}

// NewInMemoryProvider 创建内存采样器（用于测试）
func NewInMemoryProvider() (*sdktrace.TracerProvider, *tracetest.SpanRecorder) {
	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))
	return tp, sr
}
