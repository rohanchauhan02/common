package logs

import (
	"fmt"
	"io"
	"path"
	"runtime"
	"strings"
	"sync"

	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo"
	gommonLog "github.com/labstack/gommon/log"
	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

// CommonLogger return object of interface in Common log package
type CommonLogger struct {
	logger    *logrus.Logger
	prefix    string
	requestID string
}

var (
	logger   *logrus.Logger
	once     sync.Once
	instance *CommonLogger
)

// NewCommonLog is a factory that return  interface of log pakcage
func NewCommonLog(prefix ...string) *CommonLogger {
	once.Do(func() {
		logger = logrus.New()
		logger.Formatter = &prefixed.TextFormatter{
			FullTimestamp: true,
		}
		instance = &CommonLogger{
			logger: logger,
		}
		if len(prefix) > 0 {
			instance.prefix = prefix[0]
		}
	})
	// This will reset the prefix if the changes occured
	// As it will not possible to set for the test suite
	if len(prefix) > 0 && !strings.EqualFold(instance.prefix, prefix[0]) {
		instance.prefix = prefix[0]
	}

	return instance
}

func (q *CommonLogger) decorateLog() *logrus.Entry {
	var source string
	if pc, file, line, ok := runtime.Caller(2); ok {
		var funcName string
		if fn := runtime.FuncForPC(pc); fn != nil {
			funcName = fn.Name()
			if i := strings.LastIndex(funcName, "."); i != -1 {
				funcName = funcName[i+1:]
			}
		}

		source = fmt.Sprintf("%s:%v:%s()", path.Base(file), line, path.Base(funcName))
	}
	e := q.logger.WithFields(logrus.Fields{
		"source": source,
	})
	if q.prefix != "" {
		e = e.WithFields(logrus.Fields{
			"prefix": q.prefix,
		})
	}
	if q.requestID != "" {
		e = e.WithFields(logrus.Fields{
			"requestID": q.requestID,
		})
	}
	return e
}

func (q *CommonLogger) Output() io.Writer {
	return q.logger.Out
}

// SetOutput logger io.Writer
func (q *CommonLogger) SetOutput(w io.Writer) {
	// do nothing
}

func (q *CommonLogger) Prefix() string {
	return q.prefix
}

func (q *CommonLogger) SetPrefix(p string) {
	q.prefix = p
}

func (q *CommonLogger) Level() gommonLog.Lvl {
	return toEchoLevel(q.logger.Level)
}

func (q *CommonLogger) SetLevel(v gommonLog.Lvl) {
	q.logger.Level = toLogrusLevel(v)
}

func (q *CommonLogger) SetHeader(h string) {
	// do nothing
}

func (q *CommonLogger) Print(i ...interface{}) {
	q.decorateLog().Print(i...)
}

func (q *CommonLogger) Printf(format string, args ...interface{}) {
	q.decorateLog().Printf(format, args...)
}

func (q *CommonLogger) Printj(j gommonLog.JSON) {
	q.decorateLog().Printf("%+v", j)
}

func (q *CommonLogger) Debug(i ...interface{}) {
	q.decorateLog().Debug(i...)
}

func (q *CommonLogger) Debugf(format string, args ...interface{}) {
	q.decorateLog().Debugf(format, args...)
}

func (q *CommonLogger) Debugj(j gommonLog.JSON) {
	q.decorateLog().Debugf("%+v", j)
}

// Info is a logrus log message at level info on the standard logger
func (q *CommonLogger) Info(i ...interface{}) {
	q.decorateLog().Info(i...)
}

// Infof is a logrus log message at level infof on the standard logger
func (q *CommonLogger) Infof(format string, args ...interface{}) {
	q.decorateLog().Infof(format, args...)
}

func (q *CommonLogger) Infoj(j gommonLog.JSON) {
	q.decorateLog().Infof("%+v", j)
}

func (q *CommonLogger) Warn(i ...interface{}) {
	q.decorateLog().Warn(i...)
}

func (q *CommonLogger) Warnf(format string, args ...interface{}) {
	q.decorateLog().Warnf(format, args...)
}

func (q *CommonLogger) Warnj(j gommonLog.JSON) {
	q.decorateLog().Warnf("%+v", j)
}

// Error is a logrus log message at level error on the standard logger
func (q *CommonLogger) Error(i ...interface{}) {
	q.decorateLog().Error(i...)
	q.sentry(fmt.Sprintf("%+v", i...))
}

// Errorf is a logrus log message at level errorf on the standard logger
func (q *CommonLogger) Errorf(format string, args ...interface{}) {
	q.decorateLog().Errorf(format, args...)
	q.sentry(fmt.Sprintf(format, args...))
}

func (q *CommonLogger) Errorj(j gommonLog.JSON) {
	q.decorateLog().Errorf("%+v", j)
	q.sentry(fmt.Sprintf("%+v", j))
}

func (q *CommonLogger) Fatal(i ...interface{}) {
	q.decorateLog().Fatal(i...)
	q.sentry(fmt.Sprintf("%+v", i...))
}

func (q *CommonLogger) Fatalj(j gommonLog.JSON) {
	q.decorateLog().Fatalf("%+v", j)
	q.sentry(fmt.Sprintf("%+v", j))
}

func (q *CommonLogger) Fatalf(format string, args ...interface{}) {
	q.decorateLog().Fatalf(format, args...)
	q.sentry(fmt.Sprintf(format, args...))
}

func (q *CommonLogger) Panic(i ...interface{}) {
	q.decorateLog().Panic(i...)
	q.sentry(fmt.Sprintf("%+v", i...))
}

func (q *CommonLogger) Panicj(j gommonLog.JSON) {
	q.decorateLog().Panicf("%+v", j)
	q.sentry(fmt.Sprintf("%+v", j))
}

func (q *CommonLogger) Panicf(format string, args ...interface{}) {
	q.decorateLog().Panicf(format, args...)
	q.sentry(fmt.Sprintf(format, args...))
}

func (q *CommonLogger) sentry(message string) {
	sentry.CaptureMessage(message)
}

func (q *CommonLogger) MiddlewareLoggerRequestID() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			requestId := c.Request().Header.Get(echo.HeaderXRequestID)
			q.requestID = requestId
			sentry.ConfigureScope(func(scope *sentry.Scope) {
				scope.SetTag("x-request-id", requestId)
			})
			return next(c)
		}
	}
}

// To logrus.Level
func toLogrusLevel(level gommonLog.Lvl) logrus.Level {
	switch level {
	case gommonLog.DEBUG:
		return logrus.DebugLevel
	case gommonLog.INFO:
		return logrus.InfoLevel
	case gommonLog.WARN:
		return logrus.WarnLevel
	case gommonLog.ERROR:
		return logrus.ErrorLevel
	}
	return logrus.InfoLevel
}

// To Echo.gommonLog.lvl
func toEchoLevel(level logrus.Level) gommonLog.Lvl {
	switch level {
	case logrus.DebugLevel:
		return gommonLog.DEBUG
	case logrus.InfoLevel:
		return gommonLog.INFO
	case logrus.WarnLevel:
		return gommonLog.WARN
	case logrus.ErrorLevel:
		return gommonLog.ERROR
	}
	return gommonLog.OFF
}
