package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

var log zerolog.Logger

// Init configures the global zerolog logger.
// level: "debug", "info", "warn", "error"
// format: "json" or "console"
func Init(level, format string) {
	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		lvl = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(lvl)

	if format == "console" {
		log = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}).
			With().Timestamp().Caller().Logger()
	} else {
		log = zerolog.New(os.Stdout).
			With().Timestamp().Caller().Logger()
	}
}

// Get returns the configured logger instance.
func Get() *zerolog.Logger {
	return &log
}
