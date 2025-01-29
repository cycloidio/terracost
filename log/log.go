package log

import (
	"os"

	"log/slog"
)

var (
	Logger *slog.Logger

	Level = new(slog.LevelVar)
)

func init() {
	Level.Set(slog.LevelInfo)

	Logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: Level,
	}))
}
