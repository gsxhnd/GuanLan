package metrics

import (
	"testing"
	"time"
)

func TestMetricsConfig_ApplyDefaults(t *testing.T) {
	cfg := &MetricsConfig{}

	cfg.ApplyDefaults()

	if cfg.OtelEnable != false {
		t.Errorf("expected OtelEnable to be false, got %v", cfg.OtelEnable)
	}
	if cfg.OtelServiceName != "demo-go-framework" {
		t.Errorf("expected OtelServiceName to be 'demo-go-framework', got %s", cfg.OtelServiceName)
	}
	if cfg.OtelServiceVersion != "1.0.0" {
		t.Errorf("expected OtelServiceVersion to be '1.0.0', got %s", cfg.OtelServiceVersion)
	}
	if cfg.OtelEnvironment != "development" {
		t.Errorf("expected OtelEnvironment to be 'development', got %s", cfg.OtelEnvironment)
	}
	if cfg.ExportInterval != 10*time.Second {
		t.Errorf("expected ExportInterval to be 10s, got %v", cfg.ExportInterval)
	}
	if cfg.ExportTimeout != 5*time.Second {
		t.Errorf("expected ExportTimeout to be 5s, got %v", cfg.ExportTimeout)
	}
}

func TestMetricsConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     MetricsConfig
		wantErr bool
	}{
		{
			name: "valid disabled config",
			cfg: MetricsConfig{
				OtelEnable: false,
			},
			wantErr: false,
		},
		{
			name: "valid enabled config",
			cfg: MetricsConfig{
				OtelEnable:   true,
				OtelEndpoint: "localhost:4317",
			},
			wantErr: false,
		},
		{
			name: "enabled without endpoint",
			cfg: MetricsConfig{
				OtelEnable:   true,
				OtelEndpoint: "",
			},
			wantErr: true,
		},
		{
			name: "empty service name gets default",
			cfg: MetricsConfig{
				OtelEnable:      false,
				OtelServiceName: "",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMetricsConfig_Validate_DefaultServiceName(t *testing.T) {
	cfg := &MetricsConfig{
		OtelServiceName: "",
	}

	err := cfg.Validate()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if cfg.OtelServiceName != "unknown-service" {
		t.Errorf("expected OtelServiceName to be 'unknown-service', got %s", cfg.OtelServiceName)
	}
}

func TestDefaultMetricsConfig(t *testing.T) {
	cfg := DefaultMetricsConfig()

	if cfg.OtelEnable != false {
		t.Errorf("expected OtelEnable to be false")
	}
	if cfg.OtelServiceName != "demo-go-framework" {
		t.Errorf("expected OtelServiceName to be 'demo-go-framework'")
	}
	if cfg.ExportInterval != 10*time.Second {
		t.Errorf("expected ExportInterval to be 10s")
	}
	if cfg.ExportTimeout != 5*time.Second {
		t.Errorf("expected ExportTimeout to be 5s")
	}
}

func TestMetricsConfig_ExportIntervalDefaults(t *testing.T) {
	tests := []struct {
		name     string
		interval time.Duration
		expected time.Duration
	}{
		{
			name:     "zero interval",
			interval: 0,
			expected: 10 * time.Second,
		},
		{
			name:     "negative interval",
			interval: -1 * time.Second,
			expected: 10 * time.Second,
		},
		{
			name:     "positive interval",
			interval: 30 * time.Second,
			expected: 30 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &MetricsConfig{
				ExportInterval: tt.interval,
			}
			cfg.ApplyDefaults()
			if cfg.ExportInterval != tt.expected {
				t.Errorf("expected ExportInterval to be %v, got %v", tt.expected, cfg.ExportInterval)
			}
		})
	}
}

func TestMetricsConfig_ExportTimeoutDefaults(t *testing.T) {
	tests := []struct {
		name     string
		timeout  time.Duration
		expected time.Duration
	}{
		{
			name:     "zero timeout",
			timeout:  0,
			expected: 5 * time.Second,
		},
		{
			name:     "negative timeout",
			timeout:  -1 * time.Second,
			expected: 5 * time.Second,
		},
		{
			name:     "positive timeout",
			timeout:  10 * time.Second,
			expected: 10 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &MetricsConfig{
				ExportTimeout: tt.timeout,
			}
			cfg.ApplyDefaults()
			if cfg.ExportTimeout != tt.expected {
				t.Errorf("expected ExportTimeout to be %v, got %v", tt.expected, cfg.ExportTimeout)
			}
		})
	}
}
