package gearbox

import (
	"log/slog"
	"os"
)

type LogConfig struct {
	JSONLogging bool   `yaml:"json_logging"`
	Level       string `yaml:"level"`
	AddSource   bool   `yaml:"add_source"`
}

func newLogger(config LogConfig) *slog.Logger {
	var handler slog.Handler
	level, err := parseLevel(config.Level)
	if err != nil {
		level = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: config.AddSource,
	}

	if config.JSONLogging {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}

func parseLevel(level string) (slog.Level, error) {
	var l slog.Level
	err := l.UnmarshalText([]byte(level))
	return l, err
}
