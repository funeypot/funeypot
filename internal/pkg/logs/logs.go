package logs

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger = *zap.SugaredLogger

var def Logger

func init() {
	config := zap.NewDevelopmentConfig()
	logger, err := config.Build(zap.AddStacktrace(zapcore.DPanicLevel))
	if err != nil {
		panic(fmt.Errorf("init logger: %w", err))
	}
	def = logger.Sugar()
}

func Default() Logger {
	return def
}

type loggerInContext struct{}

// From returns a Logger which is stored in context as a value,
// if the Logger doesn't exist, it returns Default().WithCtx(ctx)
func From(ctx context.Context) Logger {
	if got, ok := TryFrom(ctx); ok {
		return got
	}
	return Default()
}

// TryFrom returns a Logger which is stored in context as a value,
// if the Logger doesn't exist, it returns false.
func TryFrom(ctx context.Context) (Logger, bool) {
	if got := ctx.Value(loggerInContext{}); got != nil {
		if logger, ok := got.(Logger); ok {
			return logger, true
		}
	}
	return nil, false
}

// With returns a new context with the logger stored in it as a value
func With(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, loggerInContext{}, logger)
}
