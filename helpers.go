package sdbot

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

// Sanitize returns a new string with non-alphanumeric characters removed.
// This is particularly useful when it comes to identifying unique usernames.
// Keep in mind that Pokemon Showdown usernames are _not_ case sensitive, so a
// downcased and sanitized username is a unique identifier.
func Sanitize(s string) string {
	reg := regexp.MustCompile("[^A-Za-z0-9]")
	return strings.ToLower(reg.ReplaceAllString(s, ""))
}

// Public API for access to the LoggerList loggers.

// CheckErr checks if an error is nil, and if it is not, logs the error to
// default os.Stderr logger. TODO Perhaps a nice stack trace here as well.
func CheckErr(err error) {
	if err != nil {
		Error(err)
	}
}

// CheckErrAll checks if an error is nil, and if it is not, logs the error to
// all the loggers in the LoggerList. TODO Perhaps a nice stack trace here as
// well.
func CheckErrAll(err error) {
	if err != nil {
		ErrorAll(err)
	}
}

// Debug logs debug messages to the default os.Stderr logger.
func Debug(s string) {
	logDebug(&loggers.Loggers[0], s)
}

// Info logs informatic messages to the default os.Stderr logger.
func Info(s string) {
	logInfo(&loggers.Loggers[0], s)
}

// Warn logs warning messages to the default os.Stderr logger.
func Warn(s string) {
	logWarn(&loggers.Loggers[0], s)
}

// Error logs errors to the default os.Stderr logger.
func Error(err error) {
	logError(&loggers.Loggers[0], err)
}

// Fatal logs fatal messages to the default os.Stderr logger.
func Fatal(s string) {
	logFatal(&loggers.Loggers[0], s)
}

// Debugf logs debug messages with formatting to the default os.Stderr logger.
func Debugf(format string, a ...interface{}) {
	logDebugf(&loggers.Loggers[0], format, a...)
}

// Infof logs informatic messages with formatting to the default os.Stderr
// logger.
func Infof(format string, a ...interface{}) {
	logInfof(&loggers.Loggers[0], format, a...)
}

// Warnf logs warning messages with formatting to the default os.Stderr logger.
func Warnf(format string, a ...interface{}) {
	logWarnf(&loggers.Loggers[0], format, a...)
}

// Errorf logs errors with formatting to the default os.Stderr logger.
func Errorf(format string, a ...interface{}) {
	logErrorf(&loggers.Loggers[0], format, a...)
}

// Fatalf logs fatal messages with formatting to the default os.Stderr logger.
func Fatalf(format string, a ...interface{}) {
	logFatalf(&loggers.Loggers[0], format, a...)
}

// ErrorAll logs errors to all the loggers in the LoggerList.
func ErrorAll(err error) {
	logErrorAll(loggers, err)
}

// Inspect logs a debug message to the default os.Stderr logger, taking
// any interface and inpecting its contents.
func Inspect(i interface{}) {
	Debugf("%+v", i)
}

// HasteKey holds the JSON unmarshalling of a Hastebin POST request.
type HasteKey struct {
	Key string
}

// Haste uploads a file to Hastebin and returns the response URL.
// TODO Cute, but perhaps this should be its own package.
func Haste(buf io.Reader, bodyType string) (string, error) {
	data, err := ioutil.ReadAll(buf)
	if err != nil {
		Error(err)
		return "", err
	}

	trimmedData := strings.TrimRight(string(data), "\r\n")

	res, err := http.Post("http://hastebin.com/documents", bodyType, bytes.NewBufferString(trimmedData))
	if err != nil {
		Error(err)
		return "", err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		Error(err)
		return "", err
	}

	var hk HasteKey
	err = json.Unmarshal(body, &hk)
	if err != nil {
		Error(err)
		return "", err
	}

	return "http://hastebin.com/" + hk.Key, nil
}
