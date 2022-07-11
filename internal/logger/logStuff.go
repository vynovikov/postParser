package logger

import (
	log "github.com/sirupsen/logrus"
)

var L *log.Logger

func init() {
	L = log.New()
	L.SetFormatter(&log.TextFormatter{
		ForceColors:            true,
		FullTimestamp:          true,
		TimestampFormat:        "01.02.2006 15:04:05",
		DisableLevelTruncation: true,
	})
}
