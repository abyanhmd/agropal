package utils

import (
	"agropal/config"
	"agropal/internal/domain/shared"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

type Logger struct {
	logger zerolog.Logger
}

func NewLogger(c config.Config) *Logger {
	var output io.Writer

	if strings.ToUpper(c.Stage) == "PRODUCTION" {
		logPath := path.Join(c.Path.Log, "app.log")
		file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			panic(err)
		}

		output = zerolog.MultiLevelWriter(os.Stdout, file)
		zerolog.TimeFieldFormat = time.RFC3339
	} else {
		output = zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC822Z}
	}

	logger := zerolog.New(output).With().Timestamp().Logger()

	return &Logger{
		logger: logger,
	}
}

func getCallerInfo(skip int) (funcName, file string) {
	pc, filePath, line, ok := runtime.Caller(skip)
	if !ok {
		return "unknown", "unknown"
	}
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "unknown", filePath
	}
	funcFull := fn.Name()

	wd, _ := os.Getwd()
	relPath, err := filepath.Rel(wd, filePath)
	if err != nil {
		relPath = filePath // fallback to absolute if error
	}

	// normalize slashes for consistent logging
	relPath = filepath.ToSlash(relPath)

	return funcFull, fmt.Sprintf("%s:%d", relPath, line)
}

func (l *Logger) withFields(ctx context.Context, level zerolog.Level, asset shared.AssetType) *zerolog.Event {
	var event *zerolog.Event
	switch level {
	case zerolog.InfoLevel:
		event = l.logger.Info()
	case zerolog.WarnLevel:
		event = l.logger.Warn()
	case zerolog.ErrorLevel:
		event = l.logger.Error()
	case zerolog.DebugLevel:
		event = l.logger.Debug()
	case zerolog.FatalLevel:
		event = l.logger.Fatal()
	}

	fileExclusion := []string{
		"logger",
		"handler",
		"repository",
		"application",
		"router",
	}

	funcName, filePath := getCallerInfo(3)
	shortFuncName := simplifyFuncName(funcName)
	if strings.Contains(shortFuncName, ".func") {
		shortFuncName = strings.Split(strings.Join(strings.Split(filePath, "/")[2:], "/"), ":")[0]
	} else {
		parts := strings.Split(funcName, ".")
		shortFuncName = parts[len(parts)-1]
	}

	var operation string
	if len(funcName) > 3 {
		if strings.Contains(shortFuncName, "middleware/logger.go") {
			operation = "HTTP Request"
		} else if strings.Contains(shortFuncName, ".go") {
			operation = shortFuncName
			shortFuncName += "()"
		} else {
			operation = strings.Split(filePath, ":")[0]
			shortFuncName += "()"
		}
	}
	event.Str("caller", operation)

	for _, exclude := range fileExclusion {
		if strings.Contains(funcName, exclude) && level == zerolog.InfoLevel {
			event.Str("file_path", filePath)
			break
		}
	}

	fields := []struct {
		Key   string
		Value string
	}{
		{"operation", funcName},
		{string(ContextMethod), GetMetadata(ctx, ContextMethod)},
		{string(ContextURLPath), GetMetadata(ctx, ContextURLPath)},
		{string(ContextUserAgent), GetMetadata(ctx, ContextUserAgent)},
		{string(ContextIPAddress), GetMetadata(ctx, ContextIPAddress)},
		{string(ContextUserID), GetMetadata(ctx, ContextUserID)},
		{string(ContextEmail), GetMetadata(ctx, ContextEmail)},
		{string(ContextRole), GetMetadata(ctx, ContextRole)},
	}

	for _, f := range fields {
		event.Interface(f.Key, f.Value)
	}

	return event
}

func simplifyFuncName(full string) string {
	parts := strings.Split(full, "/")
	last := parts[len(parts)-1]

	last = strings.ReplaceAll(last, "(*", "")
	last = strings.ReplaceAll(last, ")", "")

	return strings.Replace(last, ".", ".", 1)
}

func (l *Logger) Info(ctx context.Context, asset shared.AssetType, msg string) {
	l.withFields(ctx, zerolog.InfoLevel, asset).Msg(msg)
}

func (l *Logger) Warn(ctx context.Context, asset shared.AssetType, msg string) {
	l.withFields(ctx, zerolog.WarnLevel, asset).Msg(msg)
}

func (l *Logger) Error(ctx context.Context, asset shared.AssetType, msg string) {
	l.withFields(ctx, zerolog.ErrorLevel, asset).Msg(msg)
}

func (l *Logger) Debug(ctx context.Context, asset shared.AssetType, msg string) {
	l.withFields(ctx, zerolog.DebugLevel, asset).Msg(msg)
}

func (l *Logger) Fatal(ctx context.Context, asset shared.AssetType, msg string) {
	l.withFields(ctx, zerolog.FatalLevel, asset).Msg(msg)
}
