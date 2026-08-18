package metrics

import (
	"context"
	"testing"
	"time"
)

func TestNewMeterProvider_Disabled(t *testing.T) {
	cfg := &MetricsConfig{
		OtelEnable: false,
	}

	mp, err := NewMeterProvider(cfg)
	if err != nil {
		t.Fatalf("NewMeterProvider() error = %v", err)
	}
	if mp == nil {
		t.Fatal("expected MeterProvider to be non-nil")
	}
	if mp.provider == nil {
		t.Fatal("expected internal provider to be non-nil")
	}
}

func TestNewMeterProvider_InvalidConfig(t *testing.T) {
	cfg := &MetricsConfig{
		OtelEnable:   true,
		OtelEndpoint: "", // Missing endpoint
	}

	_, err := NewMeterProvider(cfg)
	if err == nil {
		t.Error("expected error for missing endpoint")
	}
}

func TestNewMeterProvider_ValidEnabledConfig(t *testing.T) {
	cfg := &MetricsConfig{
		OtelEnable:   true,
		OtelEndpoint: "localhost:4317",
	}

	mp, err := NewMeterProvider(cfg)
	if err != nil {
		t.Fatalf("NewMeterProvider() error = %v", err)
	}
	if mp == nil {
		t.Fatal("expected MeterProvider to be non-nil")
	}
	if mp.provider == nil {
		t.Fatal("expected internal provider to be non-nil")
	}
	if mp.config != cfg {
		t.Error("expected config to be stored")
	}
}

func TestMeterProvider_Shutdown(t *testing.T) {
	cfg := &MetricsConfig{
		OtelEnable: false,
	}

	mp, err := NewMeterProvider(cfg)
	if err != nil {
		t.Fatalf("NewMeterProvider() error = %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = mp.Shutdown(ctx)
	if err != nil {
		t.Errorf("Shutdown() error = %v", err)
	}
}

func TestMeterProvider_MeterProvider(t *testing.T) {
	cfg := &MetricsConfig{
		OtelEnable: false,
	}

	mp, err := NewMeterProvider(cfg)
	if err != nil {
		t.Fatalf("NewMeterProvider() error = %v", err)
	}

	provider := mp.MeterProvider()
	if provider == nil {
		t.Error("expected MeterProvider() to return non-nil")
	}
}
