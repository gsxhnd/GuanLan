// Package metrics 包含 HTTP 指标收集器实现
//
// HTTP 指标定义遵循 OpenTelemetry 语义约定：
//   - http.server.requests: HTTP 请求总数
//   - http.server.duration: HTTP 请求耗时分布（直方图）
//   - http.server.errors: HTTP 错误请求数
//   - http.server.active_requests: 当前活跃请求数
//
// 指标属性：
//   - http.request.method: HTTP 方法 (GET, POST, etc.)
//   - http.route: 路由模板路径 (使用模板而非原始路径，避免高基数)
//   - http.response.status_code: HTTP 状态码
//   - network.protocol.name: 网络协议 (http/1.1, http/2)
//
// 使用示例：
//
//	mp, err := metrics.NewMeterProvider(cfg)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer mp.Shutdown(nil)
//
//	recorder, err := metrics.NewHTTPRecorder(mp)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// 记录请求
//	recorder.RecordRequest(ctx, metrics.HTTPRequestInfo{
//	    Method:     "GET",
//	    Route:      "/api/users",
//	    StatusCode: 200,
//	    DurationMs: 45.5,
//	    HasError:   false,
//	    Protocol:   "http/1.1",
//	})
package metrics

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// HTTPRecorder HTTP metrics 记录器
type HTTPRecorder struct {
	requests       metric.Int64Counter
	duration       metric.Float64Histogram
	errors         metric.Int64Counter
	activeRequests metric.Int64UpDownCounter
}

// NewHTTPRecorder 创建 HTTP metrics 记录器
func NewHTTPRecorder(mp *MeterProvider) (*HTTPRecorder, error) {
	meter := Meter("go_sample_code")

	requests, err := meter.Int64Counter(
		"http.server.requests",
		metric.WithDescription("Total number of HTTP requests"),
		metric.WithUnit("{request}"),
	)
	if err != nil {
		return nil, err
	}

	duration, err := meter.Float64Histogram(
		"http.server.duration",
		metric.WithDescription("HTTP request duration in milliseconds"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		return nil, err
	}

	errors, err := meter.Int64Counter(
		"http.server.errors",
		metric.WithDescription("Total number of HTTP requests with 5xx responses"),
		metric.WithUnit("{error}"),
	)
	if err != nil {
		return nil, err
	}

	activeRequests, err := meter.Int64UpDownCounter(
		"http.server.active_requests",
		metric.WithDescription("Number of active HTTP requests"),
		metric.WithUnit("{request}"),
	)
	if err != nil {
		return nil, err
	}

	return &HTTPRecorder{
		requests:       requests,
		duration:       duration,
		errors:         errors,
		activeRequests: activeRequests,
	}, nil
}

// HTTPRequestInfo HTTP 请求信息
type HTTPRequestInfo struct {
	Method     string
	Route      string
	StatusCode int
	DurationMs float64
	HasError   bool
	Protocol   string
}

// RecordRequest 记录 HTTP 请求指标
func (r *HTTPRecorder) RecordRequest(ctx context.Context, info HTTPRequestInfo) {
	attrs := []attribute.KeyValue{
		attribute.String("http.request.method", info.Method),
		attribute.String("http.route", info.Route),
		attribute.Int("http.response.status_code", info.StatusCode),
		attribute.String("network.protocol.name", info.Protocol),
	}

	// 记录请求总数
	r.requests.Add(ctx, 1, metric.WithAttributes(attrs...))

	// 记录请求耗时
	durationAttrs := append(attrs,
		attribute.String("http.response.status_code_category", statusCategory(info.StatusCode)),
	)
	r.duration.Record(ctx, info.DurationMs, metric.WithAttributes(durationAttrs...))

	// 记录错误
	if info.HasError || info.StatusCode >= 500 {
		errorAttrs := append(attrs,
			attribute.String("error.type", errorType(info.StatusCode)),
		)
		r.errors.Add(ctx, 1, metric.WithAttributes(errorAttrs...))
	}
}

// ActiveRequestAdd 增加/减少活跃请求数
func (r *HTTPRecorder) ActiveRequestAdd(ctx context.Context, delta int64) {
	r.activeRequests.Add(ctx, delta)
}

// statusCategory 返回状态码类别
func statusCategory(statusCode int) string {
	switch {
	case statusCode >= 500:
		return "5xx"
	case statusCode >= 400:
		return "4xx"
	case statusCode >= 300:
		return "3xx"
	case statusCode >= 200:
		return "2xx"
	default:
		return "1xx"
	}
}

// errorType 返回错误类型
func errorType(statusCode int) string {
	switch {
	case statusCode >= 500:
		return "server_error"
	case statusCode >= 400:
		return "client_error"
	default:
		return "unknown"
	}
}
