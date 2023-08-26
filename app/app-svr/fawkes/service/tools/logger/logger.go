package logger

import (
	"context"
	"fmt"
	"runtime"

	"go-common/library/log"
)

const (
	_log        = "log"
	_callerFile = "caller_file"
	_callerFunc = "caller_func"
)

func Info(format string, args ...interface{}) {
	logFormat(log.Infov, context.Background(), format, args...)
}

func Warn(format string, args ...interface{}) {
	logFormat(log.Warnv, context.Background(), format, args...)
}

func Error(format string, args ...interface{}) {
	logFormat(log.Errorv, context.Background(), format, args...)
}

func Infoc(ctx context.Context, format string, args ...interface{}) {
	logFormat(log.Infov, ctx, format, args...)
}
func Warnc(ctx context.Context, format string, args ...interface{}) {
	logFormat(log.Warnv, ctx, format, args...)
}
func Errorc(ctx context.Context, format string, args ...interface{}) {
	logFormat(log.Errorv, ctx, format, args...)
}

// ----------

func logFormat(handler func(context.Context, ...log.D), ctx context.Context, format string, args ...interface{}) {
	callerFile, callerFunc := callerLine()
	handler(ctx,
		log.KVString(_log, fmt.Sprintf(format, args...)),
		log.KVString(_callerFile, callerFile),
		log.KVString(_callerFunc, callerFunc))
}

func callerLine() (callerFile, callerFunc string) {
	pc, file, line, ok := runtime.Caller(3)
	if !ok {
		return
	}
	f := runtime.FuncForPC(pc)
	callerFile = fmt.Sprintf("%s:%d", file, line)
	callerFunc = f.Name()
	return
}
