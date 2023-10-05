package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/go-chi/chi"

	"github.com/juan131/api-mock/internal/logger"
)

const (
	defaultPort    = 8080
	gracefulPeriod = time.Second * 30
)

type Service interface {
	// ListenAndServe listens and serves the http requests
	ListenAndServe()
	// MakeRouter initializes a http router
	MakeRouter()
	// LogConfiguration logs the service configuration
	LogConfiguration()
}

type service struct {
	cfg        *SvcConfig // service configuration
	router     *chi.Mux   // http router
	reqCounter int        // request counter
	mu         sync.Mutex // Mutual exclusion lock
}

// SvcConfig is the service configuration
type SvcConfig struct {
	port                 int                    // server listening port
	methods, subRoutes   []string               // supported sub-routes
	failureCode          int                    // response code for failed requests
	failureRespBody      map[string]interface{} // response body for failed requests
	successCode          int                    // response code for successful requests
	successRespBody      map[string]interface{} // response body for successful requests
	successRatio         float64                // ratio of successful requests
	rateLimit            int                    // rate limit (requests per second)
	rateExceededRespBody map[string]interface{} // response body for rate exceeded requests
}

// ListenAndServe listens and serves the http requests
func (svc *service) ListenAndServe() {
	// Listen and serve
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", svc.cfg.port),
		Handler:           svc.router,
		ReadHeaderTimeout: 5 * time.Second,
	}
	logger.Info("service attempting to listen on port %d", svc.cfg.port)
	err := listenAndShutdown(srv)
	if err != nil {
		logger.Fatal(err, "Server error: %+v", err)
	}
}

// Make makes a Service struct which wraps all callable methods encompassing the mock service
func Make(cfg *SvcConfig) Service {
	return &service{
		cfg: cfg,
	}
}

// LoadConfigFromEnv loads the configuration from the environment.
//
//nolint:cyclop // many env variables to parse
func LoadConfigFromEnv() (*SvcConfig, error) {
	var err error
	svc := SvcConfig{}

	portENV := os.Getenv("PORT")
	if portENV != "" {
		svc.port, err = strconv.Atoi(portENV)
		if err != nil {
			return nil, fmt.Errorf("invalid port format for PORT: %w", err)
		}
	} else {
		svc.port = defaultPort
	}

	failureCodeEnv := os.Getenv("FAILURE_RESP_CODE")
	if failureCodeEnv != "" {
		svc.failureCode, err = strconv.Atoi(failureCodeEnv)
		if err != nil {
			return nil, fmt.Errorf("invalid int format for FAILURE_RESP_CODE: %w", err)
		}
	} else {
		svc.failureCode = http.StatusBadRequest // default response code
	}

	failureRespBodyEnv := os.Getenv("FAILURE_RESP_BODY")
	if failureRespBodyEnv != "" {
		if err = json.Unmarshal([]byte(failureRespBodyEnv), &svc.failureRespBody); err != nil {
			return nil, fmt.Errorf("invalid json format for FAILURE_RESP_BODY: %w", err)
		}
	} else {
		svc.failureRespBody = map[string]interface{}{"success": false} // default response body
	}

	successCodeEnv := os.Getenv("SUCCESS_RESP_CODE")
	if successCodeEnv != "" {
		svc.successCode, err = strconv.Atoi(successCodeEnv)
		if err != nil {
			return nil, fmt.Errorf("invalid int format for SUCCESS_RESP_CODE: %w", err)
		}
	} else {
		svc.successCode = http.StatusOK // default response code
	}

	successRepBodyEnv := os.Getenv("SUCCESS_RESP_BODY")
	if successRepBodyEnv != "" {
		if err = json.Unmarshal([]byte(successRepBodyEnv), &svc.successRespBody); err != nil {
			return nil, fmt.Errorf("invalid json format for SUCCESS_RESP_BODY: %w", err)
		}
	} else {
		svc.successRespBody = map[string]interface{}{"success": true} // default response body
	}

	successRatioEnv := os.Getenv("SUCCESS_RATIO")
	if successRatioEnv != "" {
		svc.successRatio, err = strconv.ParseFloat(successRatioEnv, 64)
		if err != nil || svc.successRatio <= 0 || svc.successRatio > 1 {
			return nil, fmt.Errorf("invalid value for SUCCESS_RATIO")
		}
	} else {
		svc.successRatio = 1.0 // default success ratio
	}

	rateLimitEnv := os.Getenv("RATE_LIMIT")
	if rateLimitEnv != "" {
		svc.rateLimit, err = strconv.Atoi(rateLimitEnv)
		if err != nil {
			return nil, fmt.Errorf("invalid int format for RATE_LIMIT")
		}
	} else {
		svc.rateLimit = 1000 // default rate limit: 1000 requests per second
	}

	rateExceededRespBodyEnv := os.Getenv("RATE_EXCEEDED_RESP_BODY")
	if rateExceededRespBodyEnv != "" {
		if err = json.Unmarshal([]byte(rateExceededRespBodyEnv), &svc.rateExceededRespBody); err != nil {
			return nil, fmt.Errorf("invalid json format for RATE_EXCEEDED_RESP_BODY: %w", err)
		}
	} else {
		svc.rateExceededRespBody = map[string]interface{}{"success": false, "error": "rate limit exceeded"} // default response body
	}

	methodsEnv := os.Getenv("METHODS")
	if methodsEnv != "" {
		allowedMethods := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}
		svc.methods = strings.Split(methodsEnv, ",")
		for _, method := range svc.methods {
			if !stringSliceContains(allowedMethods, method) {
				return nil, fmt.Errorf("method %s is not allowed", method)
			}
		}
	}

	subRoutesEnv := os.Getenv("SUB_ROUTES")
	if subRoutesEnv != "" {
		svc.subRoutes = strings.Split(subRoutesEnv, ",")
	}

	return &svc, nil
}

// LogConfiguration logs the configuration of the mock service
func (svc *service) LogConfiguration() {
	logger.Debug("Mock svc configuration:")
	logger.Debug("API rate limit: %d requests per second", svc.cfg.rateLimit)
	logger.Debug("Success ratio: %f", svc.cfg.successRatio)
	logger.Debug("Supported sub routes: %s", svc.cfg.subRoutes)
	logger.Debug("Supported methods: %s", svc.cfg.methods)
}

// listenAndShutdown starts listening and serving the http requests asynchronously while at the same time listening for
// sigterm signals from the OS. In case that the server is in a shutdown cycle, it give a grace period for existing
// http requests to be served before fully closing down the server. This call is blocking.
func listenAndShutdown(server *http.Server) error {
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			logger.Fatal(err, "fail on listen")
		}
	}()
	logger.Info("server started")

	<-done
	logger.Info("server stopped via sigterm")

	ctx, cancel := context.WithTimeout(context.Background(), gracefulPeriod)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		err = fmt.Errorf("server shutdown failed, %w", err)
		return err
	}

	logger.Info("server exited properly")
	return nil
}

// stringSliceContains is a helper function to detect whether a string slice contains a string or not
func stringSliceContains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
