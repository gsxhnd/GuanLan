package logger

import (
	"context"
	"os"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelzap"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LoggerConfig struct {
	Output          string `yaml:"output"`
	Filepath        string `yaml:"filepath"`
	Filename        string `yaml:"filename"`
	Level           string `yaml:"level"`
	MaxBackups      int    `yaml:"max_backups"`
	MaxAge          int    `yaml:"max_age"`
	OtelEnable      bool   `yaml:"otel_enable"`
	OtelEndpoint    string `yaml:"otel_endpoint"`
	OtelAuth        string `yaml:"otel_auth"`
	OtelServiceName string `yaml:"otel_service_name"`
}

type Logger interface {
	Debug(string, ...zap.Field)
	DebugCtx(context.Context, string, ...zap.Field)
	Info(string, ...zap.Field)
	InfoCtx(context.Context, string, ...zap.Field)
	Warn(string, ...zap.Field)
	WarnCtx(context.Context, string, ...zap.Field)
	Error(string, ...zap.Field)
	ErrorCtx(context.Context, string, ...zap.Field)
	Panic(string, ...zap.Field)
	PanicCtx(context.Context, string, ...zap.Field)
	GetLogger() *zap.Logger
	Shutdown(ctx context.Context)
}

type logger struct {
	l            *zap.Logger
	otelProvider *log.LoggerProvider
}

func NewLogger(cfg *LoggerConfig) (Logger, error) {
	var (
		core       zapcore.Core
		level      = zap.InfoLevel
		devEncoder = zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
			MessageKey:  "msg",
			LevelKey:    "level",
			TimeKey:     "time",
			CallerKey:   "caller",
			EncodeLevel: zapcore.CapitalColorLevelEncoder,
			EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
				enc.AppendString(t.UTC().Format("2006-01-02T15:04:05.000000-07:00"))
			},
			EncodeCaller: zapcore.ShortCallerEncoder,
			EncodeDuration: func(d time.Duration, enc zapcore.PrimitiveArrayEncoder) {
				enc.AppendInt64(int64(d) / 1000000)
			},
		})
		prodEncoder = zapcore.NewJSONEncoder(zapcore.EncoderConfig{
			MessageKey:  "msg",
			LevelKey:    "level",
			TimeKey:     "time",
			CallerKey:   "caller",
			EncodeLevel: zapcore.CapitalLevelEncoder,
			EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
				enc.AppendString(t.UTC().Format("2006-01-02T15:04:05.000000-07:00"))
			},
			EncodeCaller: zapcore.ShortCallerEncoder,
			EncodeDuration: func(d time.Duration, enc zapcore.PrimitiveArrayEncoder) {
				enc.AppendInt64(int64(d) / 1000000)
			},
		})
	)

	switch cfg.Level {
	case "debug":
		level = zap.DebugLevel
	case "info":
		level = zap.InfoLevel
	case "warn":
		level = zap.WarnLevel
	default:
		level = zap.InfoLevel
	}

	switch cfg.Output {
	case "console":
		core = zapcore.NewCore(devEncoder, os.Stdout, level)
	case "file":
		core = zapcore.NewCore(prodEncoder, os.Stdout, level)
	case "none":
		core = zapcore.NewNopCore()
	default:
		core = zapcore.NewCore(devEncoder, os.Stdout, level)
	}

	if cfg.OtelEnable {
		var (
			endpoint = cfg.OtelEndpoint
			auth     = cfg.OtelAuth
			ctx      = context.Background()
		)
		exp, err := otlploggrpc.New(ctx,
			otlploggrpc.WithEndpoint(endpoint),
			otlploggrpc.WithInsecure(),
			otlploggrpc.WithHeaders(map[string]string{
				"Authorization": auth,
				"organization":  "default",
				"stream-name":   "default",
			}),
			otlploggrpc.WithCompressor("gzip"),
			otlploggrpc.WithTimeout(5*time.Second),
		)
		if err != nil {
			return nil, err
		}

		processor := log.NewBatchProcessor(exp)
		r, err := resource.New(ctx, resource.WithAttributes(attribute.String("service.name", cfg.OtelServiceName)))
		if err != nil {
			return nil, err
		}

		provider := log.NewLoggerProvider(log.WithProcessor(processor), log.WithResource(r))

		core = zapcore.NewTee(
			core,
			otelzap.NewCore(cfg.OtelServiceName, otelzap.WithLoggerProvider(provider)),
		)
		return &logger{
			l:            zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1)),
			otelProvider: provider,
		}, nil
	} else {
		return &logger{
			l: zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1)),
		}, nil
	}
}

func (l *logger) GetLogger() *zap.Logger {
	return l.l
}

func (l *logger) Debug(msg string, fields ...zap.Field) {
	l.l.Debug(msg, fields...)
}

func (l *logger) DebugCtx(ctx context.Context, msg string, fields ...zap.Field) {
	f := l.traceExtract(ctx, fields...)
	l.l.Debug(msg, append(f, zapcore.Field{Type: zapcore.SkipType, Interface: ctx})...)
}

func (l *logger) Info(msg string, fields ...zap.Field) {
	l.l.Info(msg, fields...)
}

func (l *logger) InfoCtx(ctx context.Context, msg string, fields ...zap.Field) {
	f := l.traceExtract(ctx, fields...)
	l.l.Info(msg, append(f, zapcore.Field{Type: zapcore.SkipType, Interface: ctx})...)
}

func (l *logger) Warn(msg string, fields ...zap.Field) {
	l.l.Warn(msg, fields...)
}
func (l *logger) WarnCtx(ctx context.Context, msg string, fields ...zap.Field) {
	f := l.traceExtract(ctx, fields...)
	l.l.Warn(msg, append(f, zapcore.Field{Type: zapcore.SkipType, Interface: ctx})...)
}

func (l *logger) Error(msg string, fields ...zap.Field) {
	l.l.Error(msg, fields...)
}
func (l *logger) ErrorCtx(ctx context.Context, msg string, fields ...zap.Field) {
	f := l.traceExtract(ctx, fields...)
	l.l.Error(msg, append(f, zapcore.Field{Type: zapcore.SkipType, Interface: ctx})...)
}

func (l *logger) Panic(msg string, fields ...zap.Field) {
	l.l.Panic(msg, fields...)
}
func (l *logger) PanicCtx(ctx context.Context, msg string, fields ...zap.Field) {
	f := l.traceExtract(ctx, fields...)
	l.l.Panic(msg, append(f, zapcore.Field{Type: zapcore.SkipType, Interface: ctx})...)
}

func (l *logger) traceExtract(ctx context.Context, fields ...zap.Field) []zap.Field {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		fields = append(fields,
			zap.String("trace_id", span.SpanContext().TraceID().String()),
			zap.String("span_id", span.SpanContext().SpanID().String()),
			zap.String("trace_flags", span.SpanContext().TraceFlags().String()),
		)
	}
	return fields
}

func (l *logger) Shutdown(ctx context.Context) {
	l.l.Core().Sync()
	l.l.Sync()
	if l.otelProvider != nil {
		l.otelProvider.Shutdown(ctx)
	}
}
