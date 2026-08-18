// Package metrics 配置定义
//
// MetricsConfig 结构体定义了 OpenTelemetry Metrics 的所有配置选项，
// 支持通过 YAML 配置文件进行配置。
//
// YAML 配置示例：
//
//	metrics:
//	  otel_enable: true
//	  otel_endpoint: "otel-collector:4317"
//	  otel_service_name: "demo-go-framework"
//	  otel_service_version: "1.0.0"
//	  otel_environment: "development"
//	  export_interval: "10s"
//	  export_timeout: "5s"
//	  otel_headers:
//	    custom-header: "value"
//
// 统一配置建议：
//
//	为保持 trace、log、metrics 配置一致性，推荐使用统一的 OTel 配置结构：
//	- service.name: 服务名称（所有组件保持一致）
//	- service.version: 服务版本
//	- deployment.environment: 部署环境
//
// 详见 grafana.md 文档第 5 步配置建议。
package metrics

import (
	"fmt"
	"time"
)

// MetricsConfig OpenTelemetry Metrics 配置
type MetricsConfig struct {
	OtelEnable         bool              `yaml:"otel_enable"`          // 是否启用 OTLP metrics 导出
	OtelEndpoint       string            `yaml:"otel_endpoint"`        // OTLP 端点
	OtelAuth           string            `yaml:"otel_auth"`            // 认证 token
	OtelServiceName    string            `yaml:"otel_service_name"`    // 服务名称
	OtelServiceVersion string            `yaml:"otel_service_version"` // 服务版本
	OtelEnvironment    string            `yaml:"otel_environment"`     // 部署环境
	OtelOrganization   string            `yaml:"otel_organization"`    // 组织名称
	OtelStreamName     string            `yaml:"otel_stream_name"`     // 数据流名称
	OtelHeaders        map[string]string `yaml:"otel_headers"`         // 自定义 Headers
	ExportInterval     time.Duration     `yaml:"export_interval"`      // metrics 导出周期
	ExportTimeout      time.Duration     `yaml:"export_timeout"`       // 单次导出超时时间
}

// DefaultMetricsConfig 返回默认 metrics 配置
func DefaultMetricsConfig() MetricsConfig {
	return MetricsConfig{
		OtelEnable:         false,
		OtelServiceName:    "demo-go-framework",
		OtelServiceVersion: "1.0.0",
		OtelEnvironment:    "development",
		ExportInterval:     10 * time.Second,
		ExportTimeout:      5 * time.Second,
	}
}

// ApplyDefaults 应用默认值
func (c *MetricsConfig) ApplyDefaults() {
	if c.OtelServiceName == "" {
		c.OtelServiceName = "demo-go-framework"
	}
	if c.OtelServiceVersion == "" {
		c.OtelServiceVersion = "1.0.0"
	}
	if c.OtelEnvironment == "" {
		c.OtelEnvironment = "development"
	}
	if c.ExportInterval <= 0 {
		c.ExportInterval = 10 * time.Second
	}
	if c.ExportTimeout <= 0 {
		c.ExportTimeout = 5 * time.Second
	}
}

// Validate 验证配置有效性
func (c *MetricsConfig) Validate() error {
	if c.OtelEnable {
		if c.OtelEndpoint == "" {
			return fmt.Errorf("otel_endpoint is required when otel_enable is true")
		}
	}
	if c.OtelServiceName == "" {
		c.OtelServiceName = "unknown-service"
	}
	return nil
}
