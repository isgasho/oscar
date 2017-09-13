package main

import (
	"encoding/json"
	"net/http"
)

// currLogLevel holds the current log detail desired
var currLogLevel logLevel = logLevelError

type logLevel int

// Log level values
const (
	logLevelDebug logLevel = 1
	logLevelInfo           = 2
	logLevelWarn           = 3
	logLevelError          = 4
)

func validLogLevel(lvl int) bool {
	switch logLevel(lvl) {
	case logLevelDebug:
	case logLevelInfo:
	case logLevelWarn:
	case logLevelError:
	default:
		return false
	}

	return true
}

func shouldLogDebug() bool {
	return currLogLevel <= logLevelDebug
}

func shouldLogInfo() bool {
	return currLogLevel <= logLevelInfo
}

func shouldLogWarn() bool {
	return currLogLevel <= logLevelWarn
}

func shouldLogError() bool {
	return currLogLevel <= logLevelError
}

// logLevelHandler handles GET /log-level
func logLevelHandler(w http.ResponseWriter, r *http.Request) {
	sendSuccess(w, map[string]logLevel{
		"log_level": currLogLevel,
	})
}

// setLogLevelHandler handles PUT /log-level
func setLogLevelHandler(w http.ResponseWriter, r *http.Request) {
	body := struct {
		LogLevel int `json:"log_level"`
	}{}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		sendBadReq(w, "invalid PUT body")
		return
	}

	if !validLogLevel(body.LogLevel) {
		sendBadReq(w, "invalid log level")
		return
	}

	currLogLevel = logLevel(body.LogLevel)
	sendSuccess(w, map[string]logLevel{
		"log_level": currLogLevel,
	})
}
