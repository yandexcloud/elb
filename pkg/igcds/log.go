package main

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
)

var loglevels = []string{
	zerolog.DebugLevel.String(),
	zerolog.InfoLevel.String(),
	zerolog.WarnLevel.String(),
	zerolog.ErrorLevel.String(),
}

func loglevel(v string) zerolog.Level {
	switch v {
	case zerolog.DebugLevel.String():
		return zerolog.DebugLevel
	case zerolog.InfoLevel.String():
		return zerolog.InfoLevel
	case zerolog.WarnLevel.String():
		return zerolog.WarnLevel
	case zerolog.ErrorLevel.String():
		return zerolog.ErrorLevel
	}

	fmt.Printf("[WARN] unexpected log level %s", v)
	return zerolog.InfoLevel
}

func getlog(level string) zerolog.Logger {
	return zerolog.New(os.Stdout).
		Level(zerolog.Level(loglevel(level))).
		With().
		Timestamp().
		Logger()
}
