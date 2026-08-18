package metrics

import (
	"context"
	"testing"
)

func TestNewHTTPRecorder(t *testing.T) {
	cfg := DefaultMetricsConfig()
	cfg.OtelEnable = false

	mp, err := NewMeterProvider(&cfg)
	if err != nil {
		t.Fatalf("failed to create meter provider: %v", err)
	}
	defer mp.Shutdown(context.Background())

	recorder, err := NewHTTPRecorder(mp)
	if err != nil {
		t.Fatalf("failed to create HTTPRecorder: %v", err)
	}

	if recorder == nil {
		t.Fatal("expected HTTPRecorder to be non-nil")
	}
	if recorder.requests == nil {
		t.Error("expected requests counter to be non-nil")
	}
	if recorder.duration == nil {
		t.Error("expected duration histogram to be non-nil")
	}
	if recorder.errors == nil {
		t.Error("expected errors counter to be non-nil")
	}
	if recorder.activeRequests == nil {
		t.Error("expected activeRequests counter to be non-nil")
	}
}

func TestHTTPRecorder_RecordRequest(t *testing.T) {
	cfg := DefaultMetricsConfig()
	cfg.OtelEnable = false

	mp, err := NewMeterProvider(&cfg)
	if err != nil {
		t.Fatalf("failed to create meter provider: %v", err)
	}
	defer mp.Shutdown(context.Background())

	recorder, err := NewHTTPRecorder(mp)
	if err != nil {
		t.Fatalf("failed to create HTTPRecorder: %v", err)
	}

	ctx := context.Background()

	// 测试正常请求
	recorder.RecordRequest(ctx, HTTPRequestInfo{
		Method:     "GET",
		Route:      "/api/users",
		StatusCode: 200,
		DurationMs: 50.5,
		HasError:   false,
		Protocol:   "http/1.1",
	})

	// 测试错误请求
	recorder.RecordRequest(ctx, HTTPRequestInfo{
		Method:     "POST",
		Route:      "/api/users",
		StatusCode: 500,
		DurationMs: 100.0,
		HasError:   true,
		Protocol:   "http/1.1",
	})

	// 测试 4xx 请求
	recorder.RecordRequest(ctx, HTTPRequestInfo{
		Method:     "GET",
		Route:      "/api/users/:id",
		StatusCode: 404,
		DurationMs: 10.0,
		HasError:   false,
		Protocol:   "http/2",
	})
}

func TestHTTPRecorder_ActiveRequestAdd(t *testing.T) {
	cfg := DefaultMetricsConfig()
	cfg.OtelEnable = false

	mp, err := NewMeterProvider(&cfg)
	if err != nil {
		t.Fatalf("failed to create meter provider: %v", err)
	}
	defer mp.Shutdown(context.Background())

	recorder, err := NewHTTPRecorder(mp)
	if err != nil {
		t.Fatalf("failed to create HTTPRecorder: %v", err)
	}

	ctx := context.Background()

	// 增加活跃请求
	recorder.ActiveRequestAdd(ctx, 1)
	recorder.ActiveRequestAdd(ctx, 1)
	recorder.ActiveRequestAdd(ctx, -1)
}

func TestHTTPRequestInfo(t *testing.T) {
	info := HTTPRequestInfo{
		Method:     "GET",
		Route:      "/api/health",
		StatusCode: 200,
		DurationMs: 5.5,
		HasError:   false,
		Protocol:   "http/1.1",
	}

	if info.Method != "GET" {
		t.Errorf("expected method 'GET', got '%s'", info.Method)
	}
	if info.Route != "/api/health" {
		t.Errorf("expected route '/api/health', got '%s'", info.Route)
	}
	if info.StatusCode != 200 {
		t.Errorf("expected status code 200, got %d", info.StatusCode)
	}
	if info.DurationMs != 5.5 {
		t.Errorf("expected duration 5.5, got %f", info.DurationMs)
	}
	if info.HasError {
		t.Error("expected HasError to be false")
	}
	if info.Protocol != "http/1.1" {
		t.Errorf("expected protocol 'http/1.1', got '%s'", info.Protocol)
	}
}

func TestStatusCategory(t *testing.T) {
	tests := []struct {
		statusCode int
		expected   string
	}{
		{100, "1xx"},
		{200, "2xx"},
		{201, "2xx"},
		{301, "3xx"},
		{404, "4xx"},
		{500, "5xx"},
		{503, "5xx"},
	}

	for _, tt := range tests {
		result := statusCategory(tt.statusCode)
		if result != tt.expected {
			t.Errorf("statusCategory(%d) = %s, want %s", tt.statusCode, result, tt.expected)
		}
	}
}

func TestErrorType(t *testing.T) {
	tests := []struct {
		statusCode int
		expected   string
	}{
		{200, "unknown"},
		{400, "client_error"},
		{404, "client_error"},
		{500, "server_error"},
		{503, "server_error"},
	}

	for _, tt := range tests {
		result := errorType(tt.statusCode)
		if result != tt.expected {
			t.Errorf("errorType(%d) = %s, want %s", tt.statusCode, result, tt.expected)
		}
	}
}

func TestHTTPRecorder_MultipleRecordings(t *testing.T) {
	cfg := DefaultMetricsConfig()
	cfg.OtelEnable = false

	mp, err := NewMeterProvider(&cfg)
	if err != nil {
		t.Fatalf("failed to create meter provider: %v", err)
	}
	defer mp.Shutdown(context.Background())

	recorder, err := NewHTTPRecorder(mp)
	if err != nil {
		t.Fatalf("failed to create HTTPRecorder: %v", err)
	}

	ctx := context.Background()

	// 记录多个请求
	routes := []string{"/api/users", "/api/users/:id", "/api/health"}
	methods := []string{"GET", "POST", "PUT", "DELETE"}
	statusCodes := []int{200, 201, 400, 404, 500}

	for i := 0; i < 100; i++ {
		route := routes[i%len(routes)]
		method := methods[i%len(methods)]
		status := statusCodes[i%len(statusCodes)]

		recorder.RecordRequest(ctx, HTTPRequestInfo{
			Method:     method,
			Route:      route,
			StatusCode: status,
			DurationMs: float64(i % 1000),
			HasError:   status >= 500,
			Protocol:   "http/1.1",
		})
	}
}
