package logger

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// prefix is the prefix string for each logging output
//
//nolint:gochecknoglobals // easy way to set global prefix
var prefix string

// InitGCPFormat initializes the logger to comply with GCP logging syntax
func InitGCPFormat() {
	zerolog.LevelFieldName = "severity"
	zerolog.TimestampFieldName = "timestamp"
	zerolog.TimeFieldFormat = time.RFC3339Nano

	// Pretty logger configuration which has a performance/resource impact
	prettyEnv := os.Getenv("PRETTYLOG")
	if prettyEnv != "" {
		pretty, err := strconv.ParseBool(prettyEnv)
		if err != nil {
			Error(err, "could not parse env variable")
		}
		if pretty {
			log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
		}
	}
}

// SetLogPrefix sets global prefix for logging output
func SetLogPrefix(logPrefix string) {
	prefix = fmt.Sprintf("[%s] ", logPrefix)
}

// DisableLogger disables any logging output from zerolog
func DisableLogger() {
	zerolog.SetGlobalLevel(zerolog.Disabled)
}

// Debug logs at DEBUG severity level with zerolog
func Debug(format string, v ...interface{}) {
	if os.Getenv("LOG_LEVEL") == "debug" {
		log.Debug().Msgf(fmt.Sprintf("%s%s", prefix, format), v...)
	}
}

// Info logs at INFO severity level with zerolog
func Info(format string, v ...interface{}) {
	log.Info().Msgf(fmt.Sprintf("%s%s", prefix, format), v...)
}

// Warn logs at WARN severity level with zerolog
func Warn(format string, v ...interface{}) {
	log.Warn().Msgf(fmt.Sprintf("%s%s", prefix, format), v...)
}

// Error logs at ERROR severity level with zerolog
func Error(err error, format string, v ...interface{}) {
	log.Error().Err(err).Msgf(fmt.Sprintf("%s%s", prefix, format), v...)
}

// Fatal logs error at FATAL severity level with zerolog and exits the process
func Fatal(err error, format string, v ...interface{}) {
	log.Fatal().Err(err).Msgf(fmt.Sprintf("%s%s", prefix, format), v...)
}

// RequestLogger implements a simple middleware handler for loggings request with
// some useful information such as Request URI
func RequestLogger() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			request(r, "")
			next.ServeHTTP(w, r)
		})
	}
}

// RequestFailure logs the request data
// and also error or warning messages with attached log tracking id
func RequestFailure(r *http.Request, msg string, err error) string {
	id := strconv.FormatUint(rand.Uint64(), 16)
	request(r, id)
	message := fmt.Sprintf("[%s] %s", id, msg)
	if err != nil {
		Error(err, message)
	} else {
		Warn(message)
	}
	return id
}

// request logs a request with some useful information such as Request URI.
// E.g., "POST https://backendendpoint/somepath HTTP/1.1" from https://www.someorigin.com
func request(r *http.Request, trackID string) {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	origin := r.Header.Get("Origin")
	if origin == "" {
		origin = "unknown"
	}

	reqInfo := fmt.Sprintf(
		"%s %s://%s%s %s\" from %s", r.Method,
		scheme,
		r.Host,
		r.RequestURI,
		r.Proto,
		origin,
	)
	if trackID != "" {
		reqInfo = fmt.Sprintf("[%s] %s", trackID, reqInfo)
	}

	log.Debug().Msgf(reqInfo)
}
