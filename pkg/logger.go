package pkg

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
	event := l.logger.WithLevel(level)

	if level >= zerolog.WarnLevel {
		funcName, filePath := getCallerInfo(3)
		event.Str("operation", funcName)
		event.Str("file_path", filePath)
	} else {
		event.Str("asset", string(asset))
	}

	l.addContextFields(ctx, event)

	return event
}

func (l *Logger) addContextFields(ctx context.Context, event *zerolog.Event) {
	keys := []shared.ContextKey{
		shared.KeyMethod,
		shared.KeyURLPath,
		shared.KeyUserAgent,
		shared.KeyIPAddress,
		shared.KeyUserID,
		shared.KeyEmail,
		shared.KeyRole,
	}

	for _, k := range keys {
		if val, ok := ctx.Value(k).(string); ok && val != "" {
			event.Str(string(k), val)
		}
	}
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
