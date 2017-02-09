package logging

import (
	"encoding/json"
	"fmt"
	"github.com/Sirupsen/logrus"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	"net/http"
)

const (
	LOG_TYPE_APP         = "application_log"
	LOG_TYPE_ACCESS      = "webapp_access"
	LOG_SEVERITY_INFO    = "info"
	LOG_SEVERITY_DEBUG   = "debug"
	LOG_SEVERITY_FATAL   = "fatal"
	LOG_SEVERITY_ERROR   = "error"
	LOG_SEVERITY_WARNING = "warning"

)

var log *logrus.Logger

func init() {
	log = logrus.New()
	if os.Getenv("DEBUG") == "1" {
		log.Level = logrus.DebugLevel
	}
	if os.Getenv("LOG_FORMAT") != "plain" {
		log.Formatter = &CustomJsonFormatter{}
	}
}

type Logger struct {
}

func New() *Logger {
	return &Logger{}
}

type CustomJsonFormatter struct {
}

func (f *CustomJsonFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	data := make(logrus.Fields, len(entry.Data))
	for k, v := range entry.Data {
		switch v := v.(type) {
		case error:
			// Otherwise errors are ignored by `encoding/json`
			// https://github.com/Sirupsen/logrus/issues/137
			data[k] = v.Error()
		default:
			data[k] = v
		}
	}

	serialized, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal fields to JSON, %v", err)
	}
	return append(serialized, '\n'), nil
}

func concatArgs(args ...interface{}) string {
	return strings.TrimRight(fmt.Sprintln(args...), "\n")
}
func (logger *Logger) Fatal(args ...interface{}) {
	logger.format(LOG_SEVERITY_FATAL, concatArgs(args...)).Fatal()
}

func (logger *Logger) Fatalf(format string, args ...interface{}) {
	logger.format(LOG_SEVERITY_FATAL, format, args...).Fatal()
}

func (logger *Logger) Error(args ...interface{}) {
	logger.format(LOG_SEVERITY_ERROR, concatArgs(args...)).Error()
}

func (logger *Logger) Errorf(format string, args ...interface{}) {
	logger.format(LOG_SEVERITY_ERROR, format, args...).Error()
}

func (logger *Logger) Panic(args ...interface{}) {
	logger.format(LOG_SEVERITY_FATAL, concatArgs(args...)).Error()
}

func (logger *Logger) Panicf(format string, args ...interface{}) {
	logger.format(LOG_SEVERITY_FATAL, format, args...).Error()
}

func (logger *Logger) Warning(args ...interface{}) {
	logger.format(LOG_SEVERITY_WARNING, concatArgs(args...)).Warning()
}

func (logger *Logger) Warningf(format string, args ...interface{}) {
	logger.format(LOG_SEVERITY_WARNING, format, args...).Warning()
}

func (logger *Logger) Info(args ...interface{}) {
	logger.format(LOG_SEVERITY_INFO, concatArgs(args...)).Info()
}

func (logger *Logger) Infof(format string, args ...interface{}) {
	logger.format(LOG_SEVERITY_INFO, format, args...).Info()
}

func (logger *Logger) Debug(args ...interface{}) {
	logger.format(LOG_SEVERITY_DEBUG, concatArgs(args...)).Debug()
}

func (logger *Logger) Debugf(format string, args ...interface{}) {
	logger.format(LOG_SEVERITY_DEBUG, format, args...).Debug()
}

func (logger *Logger) format(severity, format string, args ...interface{}) logrus.FieldLogger {
	hostname, _ := os.Hostname()
	_, file, line, _ := runtime.Caller(2) // skip 2 levels inside logger.go

	return log.WithFields(logrus.Fields{
		"log_type":    LOG_TYPE_APP,
		"@timestamp":  time.Now().Format("2006-01-02T15:04:05.999-07:00"),
		"severity":    severity,
		"pid":         os.Getpid(),
		"description": fmt.Sprintf(format, args...),
		"server_name": hostname,
		"class":       fmt.Sprintf("%s:%d", filepath.Base(file), line),
	})
}

func (logger *Logger) Access(req *http.Request, res http.ResponseWriter, res_time_ms time.Duration, res_status int, res_size int, extra map[string]interface{}) {
	fields := logrus.Fields{
		"@timestamp":       time.Now().Format("2006-01-02T15:04:05.999-07:00"),
		"log_type":         LOG_TYPE_ACCESS,
		"remote_host":      strings.Split(req.RemoteAddr, ":")[0], // don't care about the port
		"server_name":      strings.Split(req.Host, ":")[0],       // don't care about the port
		"request_command":  req.Method,
		"request_uri":      req.RequestURI,
		"request_protocol": req.Proto,
		"status_code":      res_status,
		"response_time":    res_time_ms,
		"bytes_sent":       res_size,
		"content_type":     strings.Split(res.Header().Get("content-type"), ";")[0],
	}

	if extra != nil {
		for k, v := range extra {
			fields[k] = v
		}
	}
	log.WithFields(fields).Info("")
}
