// Package log is an wrapper package that exposes underlying logger instance.
package log

import (
	"log"
	"os"
	"sort"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Log levels as constants.
const (
	LevelInfo  = "info"
	LevelWarn  = "warn"
	LevelDebug = "debug"
	LevelError = "error"
	LevelFatal = "fatal"
)

// declare default logger
var logger Logger

// GetZap returns an instance of the underlying Zap logger.
func GetZap() *zap.Logger {
	return logger.(*Log).zap
}

// initialize logger during package initialization
func init() {
	lgr, err := New()
	if err != nil {
		log.Panic(err)
	}

	logger = lgr
}

// Debug logs a message with the "debug" severity level.
// It takes a message and zero or more fields.
func Debug(msg any, fds ...Field) {
	logger.Debug(msg, fds...)
}

// Info logs a message with the "info" severity level.
// It takes a message  and zero or more fields.
func Info(msg any, fds ...Field) {
	logger.Info(msg, fds...)
}

// Warn logs a message with the "warn" severity level.
// It takes a message and zero or more fields.
func Warn(msg any, fds ...Field) {
	logger.Warn(msg, fds...)
}

// Error logs a message with the "error" severity level.
// It takes an error and zero or more fields.
func Error(msg any, fds ...Field) {
	logger.Error(msg, fds...)
}

// Fatal logs a message with the "fatal" severity level.
// It takes a message and zero or more fields.
// Also it will exit the program with `1“ exit code.
func Fatal(msg any, fds ...Field) {
	logger.Fatal(msg, fds...)
}

// Panic logs a message with the "panic" severity level.
// It takes a message and zero or more fields.
// Also it will panic.
func Panic(msg any, fds ...Field) {
	logger.Panic(msg, fds...)
}

// Sync calls the underlying Core's Sync method, flushing any buffered log
// entries. Applications should take care to call Sync before exiting.
func Sync() error {
	return logger.Sync()
}

// InfoLogger is an interface for logging messages with the "info" severity level.
type InfoLogger interface {
	Info(msg any, fds ...Field)
}

// WarnLogger is an interface for logging messages with the "warn" severity level.
type WarnLogger interface {
	Warn(msg any, fds ...Field)
}

// DebugLogger is an interface for logging messages with the "debug" severity level.
type DebugLogger interface {
	Debug(msg any, fds ...Field)
}

// ErrorLogger is an interface for logging messages with the "error" severity level.
type ErrorLogger interface {
	Error(msg any, fds ...Field)
}

// FatalLogger is an interface for logging messages with the "fatal" severity level.
type FatalLogger interface {
	Fatal(msg any, fds ...Field)
}

// PanicLogger is an interface for logging messages with the "panic" severity level.
type PanicLogger interface {
	Panic(msg any, fds ...Field)
}

// Synchronizer is an interface that exposes Sync() method.
type Synchronizer interface {
	Sync() error
}

// Logger is an interface that encompasses all the different severity levels.
type Logger interface {
	InfoLogger
	WarnLogger
	DebugLogger
	ErrorLogger
	FatalLogger
	PanicLogger
	Synchronizer
}

// Field is an type to pass additional values to the log function.
type Field = zap.Field

// Any creates a new `Field` that associates a key with an arbitrary value.
// It takes a string key and an interface{} value, and returns a `Field`.
func Any(key string, val any) Field {
	return zap.Any(key, val)
}

// Tip returns a `Field` with the given value, using the key "tip".
// This function is a shorthand for creating a field with a message value.
func Tip(val any) Field {
	return Any("tip", val)
}

// New creates a new instance of a logger.
func New() (Logger, error) {
	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.TimeKey = "time"
	cfg.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	cfg.DisableStacktrace = true

	// Set the log level based on the `LOG_LEVEL` environment variable, or use the default of `info`.
	switch os.Getenv("LOG_LEVEL") {
	case LevelInfo:
		cfg.Level.SetLevel(zap.InfoLevel)
	case LevelWarn:
		cfg.Level.SetLevel(zap.WarnLevel)
	case LevelDebug:
		cfg.Level.SetLevel(zap.DebugLevel)
	case LevelError:
		cfg.Level.SetLevel(zap.ErrorLevel)
	case LevelFatal:
		cfg.Level.SetLevel(zap.FatalLevel)
	default:
		cfg.Level.SetLevel(zap.InfoLevel)
	}

	var lgr *zap.Logger
	var err error

	// Split logging config is experimental, this can be used to disable it.
	if os.Getenv("DISABLE_SPLIT_LOGGING") != "" {
		lgr, err = BuildDefaultLogger(cfg)
	} else {
		lgr, err = BuildSplitLogger(cfg)
	}

	if err != nil {
		return nil, err
	}

	return &Log{zap: lgr}, nil
}

// BuildSplitLogger returns a logger built using the default zap.Config.Build method. Sends all logs to stderr.
func BuildDefaultLogger(cfg zap.Config) (*zap.Logger, error) {
	// Build the logger with a caller skip of 2, which causes the logger to report the line number of the calling function.
	return cfg.Build(zap.AddCallerSkip(2))
}

// BuildSplitLogger returns a logger that sends some logs to stdout and some to stderr.
func BuildSplitLogger(cfg zap.Config) (*zap.Logger, error) {
	warnAndUp := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.WarnLevel && cfg.Level.Enabled(lvl)
	})
	lessThanWarn := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl < zapcore.WarnLevel && cfg.Level.Enabled(lvl)
	})
	encoder := zapcore.NewJSONEncoder(cfg.EncoderConfig)
	core := zapcore.NewTee(
		zapcore.NewCore(encoder, zapcore.Lock(os.Stderr), warnAndUp),
		zapcore.NewCore(encoder, zapcore.Lock(os.Stdout), lessThanWarn),
	)

	opts, err := buildCfgOptions(cfg)
	if err != nil {
		return nil, err
	}
	logger := zap.New(core, opts...)

	// Build the logger with a caller skip of 2, which causes the logger to report the line number of the calling function.
	logger = logger.WithOptions(zap.AddCallerSkip(2))
	return logger, nil
}

// buildCfgOptions replicates the internal zap.config.buildOptions(), which we can't use if we want to split
// logs into stdout and stderr.
func buildCfgOptions(cfg zap.Config) ([]zap.Option, error) {
	// Sink for internal logger errors
	errSink, _, err := zap.Open("stderr")
	if err != nil {
		return nil, err
	}
	opts := []zap.Option{zap.ErrorOutput(errSink)}

	opts = append(opts, zap.AddCaller())

	if scfg := cfg.Sampling; scfg != nil {
		opts = append(opts, zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			var samplerOpts []zapcore.SamplerOption
			if scfg.Hook != nil {
				samplerOpts = append(samplerOpts, zapcore.SamplerHook(scfg.Hook))
			}
			return zapcore.NewSamplerWithOptions(
				core,
				time.Second,
				cfg.Sampling.Initial,
				cfg.Sampling.Thereafter,
				samplerOpts...,
			)
		}))
	}

	if len(cfg.InitialFields) > 0 {
		fs := make([]Field, 0, len(cfg.InitialFields))
		keys := make([]string, 0, len(cfg.InitialFields))
		for k := range cfg.InitialFields {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fs = append(fs, Any(k, cfg.InitialFields[k]))
		}
		opts = append(opts, zap.Fields(fs...))
	}

	return opts, nil
}

// Log is a struct that holds a logger implementation.
type Log struct {
	zap *zap.Logger
}

// Debug logs a message with the "debug" severity level.
// It takes a message string and zero or more fields.
func (l *Log) Debug(msg any, fds ...Field) {
	l.zap.Debug(getMessage(msg), fds...)
}

// Info logs a message with the "info" severity level.
// It takes a message string and zero or more fields.
func (l *Log) Info(msg any, fds ...Field) {
	l.zap.Info(getMessage(msg), fds...)
}

// Warn logs a message with the "warn" severity level.
// It takes a message string and zero or more fields.
func (l *Log) Warn(msg any, fds ...Field) {
	l.zap.Warn(getMessage(msg), fds...)
}

// Error logs a message with the "error" severity level.
// It takes a message and zero or more fields.
func (l *Log) Error(msg any, fds ...Field) {
	l.zap.Error(getMessage(msg), fds...)
}

// Fatal logs a message with the "fatal" severity level.
// It takes a message and zero or more fields.
// Also it will exit the program with `1“ exit code.
func (l *Log) Fatal(msg any, fds ...Field) {
	l.zap.Fatal(getMessage(msg), fds...)
}

// Fatal logs a message with the "panic" severity level.
// It takes a message and zero or more fields.
// Also it will panic.
func (l *Log) Panic(msg any, fds ...Field) {
	l.zap.Panic(getMessage(msg), fds...)
}

// Sync calls the underlying Core's Sync method, flushing any buffered log
// entries. Applications should take care to call Sync before exiting.
func (l *Log) Sync() error {
	return l.zap.Sync()
}
