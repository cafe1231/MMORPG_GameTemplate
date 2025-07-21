package logger

import (
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

type Logger interface {
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	WithField(key string, value interface{}) Logger
	WithFields(fields map[string]interface{}) Logger
}

type logger struct {
	*logrus.Entry
}

func New(service string) Logger {
	log := logrus.New()

	log.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339Nano,
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "@timestamp",
			logrus.FieldKeyLevel: "level",
			logrus.FieldKeyMsg:   "message",
		},
	})

	level := os.Getenv("LOG_LEVEL")
	if level == "" {
		level = "info"
	}

	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		log.Warnf("Invalid log level %s, defaulting to info", level)
		logLevel = logrus.InfoLevel
	}
	log.SetLevel(logLevel)

	log.SetOutput(os.Stdout)

	return &logger{
		Entry: log.WithFields(logrus.Fields{
			"service": service,
			"version": getVersion(),
		}),
	}
}

func (l *logger) WithField(key string, value interface{}) Logger {
	return &logger{Entry: l.Entry.WithField(key, value)}
}

func (l *logger) WithFields(fields map[string]interface{}) Logger {
	logFields := logrus.Fields{}
	for k, v := range fields {
		logFields[k] = v
	}
	return &logger{Entry: l.Entry.WithFields(logFields)}
}

func getVersion() string {
	version := os.Getenv("SERVICE_VERSION")
	if version == "" {
		version = "0.1.0"
	}
	return version
}

func NewNoop() Logger {
	log := logrus.New()
	log.SetOutput(os.NewFile(0, os.DevNull))
	return &logger{Entry: logrus.NewEntry(log)}
}