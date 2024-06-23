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

type CustomLogHandler struct{
	attrs []slog.Attr
}

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

	allAttrs := append([]slog.Attr{}, h.attrs...)
	record.Attrs(func(attr slog.Attr) bool {
		allAttrs = append(allAttrs, attr)
		return true
	})

	args := []interface{}{}
	for _, attr := range allAttrs {
		//fmt.Printf(">>>> attr.Key: %s\n", attr.Key)
		//fmt.Printf(">>>> attr.Value: %v\n", attr.Value)
		if attr.Key != "!BADKEY" {
			args = append(args, attr.Key)
		}
		args = append(args, attr.Value)
	}

	fileLine := fmt.Sprintf("%s:%d", filepath.Base(file), line)
	formattedMsg := fmt.Sprintf(msg, args...)

	_, err := fmt.Fprintf(os.Stdout, "%s [%-5s] [%-13s] [%-16s]: %s\n", time, level, fileLine, funcName, formattedMsg)
	if err != nil {
		return err
	}

	return nil
}

func (h CustomLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return CustomLogHandler{attrs: append(h.attrs, attrs...)}
}

func (h CustomLogHandler) WithGroup(name string) slog.Handler {
	return h
}

func SetupLogger() {
	logger := slog.New(CustomLogHandler{})
	Log = logger
}
