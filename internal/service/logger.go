package service

import (
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"strconv"
)

// newStructuredLogger creates a new structured logger compatible with GCP logging syntax
func newStructuredLogger(level slog.Level) *slog.Logger {
	return slog.New(
		slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			Level: level,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				switch a.Key {
				case slog.LevelKey:
					return slog.Attr{
						Key:   "severity",
						Value: a.Value,
					}
				case slog.TimeKey:
					return slog.Attr{
						Key:   "timestamp",
						Value: a.Value,
					}
				default:
					return a
				}
			},
		}),
	)
}

// RequestLogger implements a simple middleware handler for loggings request with
// some useful information such as Request URI
func (svc *service) RequestLogger() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			svc.request(r, "")
			next.ServeHTTP(w, r)
		})
	}
}

// LogRequestFailure logs the request data
// and also error or warning messages with attached log tracking id
func (svc *service) LogRequestFailure(r *http.Request, msg string, err error) string {
	id := strconv.FormatUint(rand.Uint64(), 16)
	svc.request(r, id)
	message := fmt.Sprintf("[%s] %s", id, msg)
	if err != nil {
		svc.logger.Error(message, "error", err)
	} else {
		svc.logger.Warn(message)
	}
	return id
}

// request logs a request with some useful information such as Request URI.
// E.g., "POST https://backendendpoint/somepath HTTP/1.1" from https://www.someorigin.com
func (svc *service) request(r *http.Request, trackID string) {
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

	svc.logger.Debug(reqInfo)
}
