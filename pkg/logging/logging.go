package logging

import (
	"os"

	"github.com/sirupsen/logrus"
)

var DefaultLogger = InitializeDefaultLogger()

func InitializeDefaultLogger() *logrus.Logger {
	logger := logrus.New()
	debugMode, set := os.LookupEnv("HAWKV6_GENERIC_PROCESSOR_DEBUG")
	if set && debugMode == "true" || set && debugMode == "TRUE" {
		logger.SetLevel(logrus.DebugLevel)
	} else {
		logger.SetLevel(logrus.InfoLevel)
	}
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	return logger
}
