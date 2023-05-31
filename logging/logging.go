package logging

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
)

func InitLogger(logLevel string, logFormatter string) {
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	logLevels := map[string]zerolog.Level{
		"panic": zerolog.PanicLevel,
		"fatal": zerolog.FatalLevel,
		"error": zerolog.ErrorLevel,
		"warn":  zerolog.WarnLevel,
		"info":  zerolog.InfoLevel,
		"debug": zerolog.DebugLevel,
	}
	zerolog.SetGlobalLevel(logLevels[logLevel])
	if logFormatter == "text" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
}
