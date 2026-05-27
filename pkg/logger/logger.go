package logger

import (
	"io"
	"log/slog"
	"os"

	"gopkg.in/natefinch/lumberjack.v2"
)

var Log *slog.Logger

func Init() {
	writer := io.MultiWriter(os.Stdout,
		&lumberjack.Logger{
			Filename:   "./app.log",
			MaxSize:    20,
			MaxBackups: 5,
		})

	handler := slog.NewTextHandler(writer, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	})

	Log = slog.New(handler)
}
