package prox

import (
	"log"

	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger

// InitLog initialises the logger with options specified.
func InitLog(l *logrus.Logger) {
	logger = l
	log.SetOutput(logger.Writer())
}

// init will initialise the logger when the package is imported and
// open the maxmind database to find country info.
func init() {
	l := logrus.New()
	l.SetLevel(logrus.ErrorLevel)

	InitLog(l)
}
