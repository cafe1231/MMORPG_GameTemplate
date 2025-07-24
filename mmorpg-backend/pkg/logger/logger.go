package logger

import (
	"os"
	"time"

	"github.com/gin-gonic/gin"
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
	WithError(err error) Logger
}

type logger struct {
	*logrus.Entry
}

func NewWithService(service string) Logger {
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

// New creates a new logger with default service name
func New() Logger {
	return NewWithService("mmorpg")
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

func (l *logger) WithError(err error) Logger {
	return &logger{Entry: l.Entry.WithError(err)}
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

// GinLogger returns a gin middleware for logging requests
func GinLogger() gin.HandlerFunc {
	log := New()
	
	return func(c *gin.Context) {
		startTime := time.Now()
		
		// Process request
		c.Next()
		
		// Log request details
		latency := time.Since(startTime)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		path := c.Request.URL.Path
		
		fields := map[string]interface{}{
			"status":     statusCode,
			"latency_ms": latency.Milliseconds(),
			"client_ip":  clientIP,
			"method":     method,
			"path":       path,
			"user_agent": c.Request.UserAgent(),
		}
		
		if len(c.Errors) > 0 {
			fields["errors"] = c.Errors.String()
		}
		
		entry := log.WithFields(fields)
		
		if statusCode >= 500 {
			entry.Error("Server error")
		} else if statusCode >= 400 {
			entry.Warn("Client error")
		} else {
			entry.Info("Request completed")
		}
	}
}