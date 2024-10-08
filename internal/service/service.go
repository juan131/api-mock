package service

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
)

const gracefulPeriod = time.Second * 30

type Service interface {
	// ListenAndServe listens and serves the http requests
	ListenAndServe()
	// MakeRouter initializes a http router
	MakeRouter()
	// LoadConfig loads the configuration from the environment
	LoadConfig() error
}

type service struct {
	cfg        *config      // service configuration
	router     *chi.Mux     // http router
	reqCounter int          // request counter
	mu         sync.Mutex   // Mutual exclusion lock
	logger     *slog.Logger // logger
}

// NewService creates a new service
func NewService() Service {
	switch os.Getenv("LOG_LEVEL") {
	case "debug":
		return &service{logger: newStructuredLogger(slog.LevelDebug)}
	case "warn":
		return &service{logger: newStructuredLogger(slog.LevelWarn)}
	case "error":
		return &service{logger: newStructuredLogger(slog.LevelError)}
	default:
		return &service{logger: newStructuredLogger(slog.LevelInfo)}
	}
}

// ListenAndServe listens and serves the http requests
func (svc *service) ListenAndServe() {
	// Listen and serve
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", svc.cfg.port),
		Handler:           svc.router,
		ReadHeaderTimeout: 5 * time.Second,
	}
	svc.logger.Info(fmt.Sprintf("service attempting to listen on port %d", svc.cfg.port))
	if err := svc.listenAndShutdown(srv); err != nil {
		svc.logger.Error("server error", "error", err)
		os.Exit(1)
	}
}

// LoadConfig loads the service configuration
func (svc *service) LoadConfig() error {
	var err error
	svc.cfg, err = loadConfigFromEnv()
	if err != nil {
		return err
	}

	svc.logger.Debug("Mock svc configuration:")
	svc.logger.Debug(fmt.Sprintf("API rate limit: %d requests per second", svc.cfg.rateLimit))
	svc.logger.Debug(fmt.Sprintf("Success ratio: %f", svc.cfg.successRatio))
	svc.logger.Debug(fmt.Sprintf("Supported sub routes: %s", svc.cfg.subRoutes))
	svc.logger.Debug(fmt.Sprintf("Supported methods: %s", svc.cfg.methods))

	return nil
}

// listenAndShutdown starts listening and serving the http requests asynchronously while at the same time listening for
// sigterm signals from the OS. In case that the server is in a shutdown cycle, it give a grace period for existing
// http requests to be served before fully closing down the server. This call is blocking.
func (svc *service) listenAndShutdown(server *http.Server) error {
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			svc.logger.Error("fail on listen", "error", err)
			os.Exit(1)
		}
	}()
	svc.logger.Info("server started")

	<-done
	svc.logger.Info("server stopped via sigterm")

	ctx, cancel := context.WithTimeout(context.Background(), gracefulPeriod)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		err = fmt.Errorf("server shutdown failed, %w", err)
		return err
	}

	svc.logger.Info("server exited properly")
	return nil
}
