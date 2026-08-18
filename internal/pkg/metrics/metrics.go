// Package metrics 提供 OpenTelemetry Metrics 功能
//
// 该包实现了基于 OpenTelemetry 的 metrics 收集和导出能力，
// 支持将指标数据通过 OTLP 协议发送到 OpenTelemetry Collector。
//
// 主要功能：
//   - 创建和管理 MeterProvider
//   - HTTP 请求指标收集（requests, duration, errors, active_requests）
//   - 统一配置管理
//   - 与 Grafana/Prometheus 生态集成
//
// 架构说明：
//
//	Go App -> OpenTelemetry SDK -> OTLP Exporter -> OTel Collector -> Prometheus -> Grafana
//
// 配置示例：
//
//	cfg := &MetricsConfig{
//	    OtelEnable:         true,
//	    OtelEndpoint:       "otel-collector:4317",
//	    OtelServiceName:    "demo-go-framework",
//	    OtelServiceVersion: "1.0.0",
//	    OtelEnvironment:    "development",
//	    ExportInterval:     10 * time.Second,
//	}
//
//	mp, err := metrics.NewMeterProvider(cfg)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer mp.Shutdown(context.Background())
//
// 相关文档：
//   - Grafana OTel 接入: ../../grafana.md
//   - OTel Collector 配置: ../../config/otel-collector-config.yaml
package metrics

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	otelpkg "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

// MeterProviderOption MeterProvider 配置选项
type MeterProviderOption = metric.Option

// MeterProvider OTLP MeterProvider 配置
type MeterProvider struct {
	provider *metric.MeterProvider
	config   *MetricsConfig
}

// NewMeterProvider 创建 MeterProvider
func NewMeterProvider(cfg *MetricsConfig) (*MeterProvider, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("metrics config validation failed: %w", err)
	}

	mp := &MeterProvider{
		config: cfg,
	}

	ctx := context.Background()
	r, err := cfg.newResource(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	opts := []metric.Option{
		metric.WithResource(r),
	}

	if cfg.OtelEnable {
		exp, err := cfg.newExporter(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
		}

		reader := metric.NewPeriodicReader(exp,
			metric.WithInterval(cfg.ExportInterval),
			metric.WithTimeout(cfg.ExportTimeout),
		)
		opts = append(opts, metric.WithReader(reader))
	}

	provider := metric.NewMeterProvider(opts...)
	mp.provider = provider
	otel.SetMeterProvider(provider)

	return mp, nil
}

// newResource 创建资源定义
func (c *MetricsConfig) newResource(ctx context.Context) (*resource.Resource, error) {
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

// newExporter 创建 OTLP metrics 导出器
func (c *MetricsConfig) newExporter(ctx context.Context) (metric.Exporter, error) {
	endpoint := c.OtelEndpoint

	headers := map[string]string{
		"Authorization": c.OtelAuth,
	}
	// 添加自定义 headers
	for k, v := range c.OtelHeaders {
		headers[k] = v
	}
	// 默认值可被自定义覆盖
	if _, ok := headers["organization"]; !ok && c.OtelOrganization != "" {
		headers["organization"] = c.OtelOrganization
	}
	if _, ok := headers["stream-name"]; !ok && c.OtelStreamName != "" {
		headers["stream-name"] = c.OtelStreamName
	}

	return otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithEndpoint(endpoint),
		otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithHeaders(headers),
		otlpmetricgrpc.WithCompressor("gzip"),
		otlpmetricgrpc.WithTimeout(c.ExportTimeout),
	)
}

// MeterProvider 返回内部的 MeterProvider
func (mp *MeterProvider) MeterProvider() *metric.MeterProvider {
	return mp.provider
}

// Shutdown 关闭 MeterProvider
func (mp *MeterProvider) Shutdown(ctx context.Context) error {
	if mp.provider != nil {
		return mp.provider.Shutdown(ctx)
	}
	return nil
}

// Meter 返回全局 Meter 实例
func Meter(name string) otelpkg.Meter {
	return otel.Meter(name)
}
