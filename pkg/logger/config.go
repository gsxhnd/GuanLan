package logger

// Config holds logger configuration
type Config = LoggerConfig

// DefaultConfig returns a default logger configuration
func DefaultConfig() *Config {
	return &Config{
		Output:          "console",
		Level:           "info",
		MaxBackups:      7,
		MaxAge:          30,
		OtelEnable:      false,
		OtelServiceName: "app",
	}
}
