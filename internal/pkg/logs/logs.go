package logs

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var def *zap.SugaredLogger

func init() {
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	logger, err := config.Build(zap.AddStacktrace(zapcore.DPanicLevel))
	if err != nil {
		panic(fmt.Errorf("init logger: %w", err))
	}
	def = logger.Sugar()
}

func Default() *zap.SugaredLogger {
	return def
}

type loggerInContext struct{}

// From returns a Logger which is stored in context as a value,
// if the Logger doesn't exist, it returns Default().WithCtx(ctx)
func From(ctx context.Context) *zap.SugaredLogger {
	if got, ok := TryFrom(ctx); ok {
		return got
	}
	return Default()
}

// TryFrom returns a Logger which is stored in context as a value,
// if the Logger doesn't exist, it returns false.
func TryFrom(ctx context.Context) (*zap.SugaredLogger, bool) {
	if got := ctx.Value(loggerInContext{}); got != nil {
		if logger, ok := got.(*zap.SugaredLogger); ok {
			return logger, true
		}
	}
	return nil, false
}

// With returns a new context with the logger stored in it as a value
func With(ctx context.Context, logger *zap.SugaredLogger) context.Context {
	return context.WithValue(ctx, loggerInContext{}, logger)
}
