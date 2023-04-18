package logging

import (
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/onrik/logrus/filename"
)

func InitLogger(logLevel string, logFormatter string) {
	logLevels := map[string]log.Level{
		"panic": log.PanicLevel,
		"fatal": log.FatalLevel,
		"error": log.ErrorLevel,
		"warn":  log.WarnLevel,
		"info":  log.InfoLevel,
		"debug": log.DebugLevel,
	}
	log.SetLevel(logLevels[logLevel])
	log.AddHook(filename.NewHook())
	if logFormatter == "text" {
		customFormatter := new(log.TextFormatter)
		customFormatter.FullTimestamp = true
		customFormatter.TimestampFormat = "2006-01-02 15:04:05"
		log.SetFormatter(customFormatter)
	} else {
		log.SetFormatter(new(log.JSONFormatter))
	}
	log.SetOutput(os.Stdout)
}
