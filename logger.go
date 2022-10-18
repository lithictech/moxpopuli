package moxpopuli

import "context"

type LogLevel string

const (
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
)

type Logger interface {
	Log(level LogLevel, msg ...interface{})
}

func LoggerInContext(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

func LoggerFromContext(ctx context.Context) Logger {
	return ctx.Value(loggerKey).(Logger)
}

const loggerKey = "_moxpopuliLogger"
