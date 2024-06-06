package log

import (
	"context"
	"fmt"
	"golang.org/x/exp/slog"
	"os"
	"path/filepath"
	"runtime"
)

// Log Global logger
var Log *slog.Logger

type CustomLogHandler struct{}

func (h CustomLogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}

func (h CustomLogHandler) Handle(ctx context.Context, record slog.Record) error {
	pc, file, line, ok := runtime.Caller(3)
	if !ok {
		file = "???"
		line = 0
	}
	funcName := runtime.FuncForPC(pc).Name()
	funcName = filepath.Base(funcName)

	level := record.Level.String()
	time := record.Time.Format("2006/01/02 15:04:05")
	msg := record.Message

	_, err := fmt.Fprintf(os.Stdout, "%s [%-5s] [%s:%d, %s]: %s\n", time, level, filepath.Base(file), line, funcName, msg)
	if err != nil {
		return err
	}

	return nil
}

func (h CustomLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h CustomLogHandler) WithGroup(name string) slog.Handler {
	return h
}

func SetupLogger() {
	logger := slog.New(CustomLogHandler{})
	Log = logger
}
