package logger

import (
	"context"

	"go.uber.org/zap"
)

var (
	baseLogger    *zap.Logger
	loggerFields  = "logger.fields"
	RequestFields = "request.fields"

	FunctionName = "function.name"
)

func init() {
	var err error
	baseLogger, err = zap.NewProduction()
	if err != nil {
		panic(err)
	}
}

func With(ctx context.Context, fields ...zap.Field) context.Context {
	data := ctx.Value(loggerFields)
	storeFields := make([]zap.Field, 0)
	if data != nil {
		storeFields = data.([]zap.Field)
	}
	storeFields = append(storeFields, fields...)
	return context.WithValue(ctx, loggerFields, storeFields)
}

func FromContext(ctx context.Context) []zap.Field {
	data := ctx.Value(loggerFields)
	fields := make([]zap.Field, 0)
	if data != nil {
		fields = data.([]zap.Field)
	}
	return fields
}

func Info(ctx context.Context, msg string, fields ...zap.Field) {
	storeFields := FromContext(ctx)
	storeFields = append(storeFields, fields...)
	baseLogger.Info(msg, storeFields...)
}

func Error(ctx context.Context, msg string, fields ...zap.Field) {
	storeFields := FromContext(ctx)
	storeFields = append(storeFields, fields...)
	baseLogger.Error(msg, storeFields...)
}

func Debug(ctx context.Context, msg string, fields ...zap.Field) {
	storeFields := FromContext(ctx)
	storeFields = append(storeFields, fields...)
	baseLogger.Debug(msg, storeFields...)
}

func Warn(ctx context.Context, msg string, fields ...zap.Field) {
	storeFields := FromContext(ctx)
	storeFields = append(storeFields, fields...)
	baseLogger.Warn(msg, storeFields...)
}

func Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	storeFields := FromContext(ctx)
	storeFields = append(storeFields, fields...)
	baseLogger.Fatal(msg, storeFields...)
}

func Base() *zap.Logger {
	return baseLogger
}
