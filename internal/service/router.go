package service

import (
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/go-chi/render"

	"github.com/juan131/api-mock/pkg/authn"
)

const uriPrefix string = "/v1/mock"

// MakeRouter initiates the service's http router with a chi Mux object.
// This also include routes initializations.
func (svc *service) MakeRouter() {
	router := chi.NewRouter()

	// Middlewares
	router.Use(
		cors.Handler(cors.Options{
			AllowedOrigins:   []string{"*"},
			AllowedMethods:   svc.cfg.methods,
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
			ExposedHeaders:   []string{"Link"},
			AllowCredentials: true,
			MaxAge:           300, // Maximum value not ignored by any of major browsers
		}),
		middleware.Recoverer,
		middleware.Timeout(30*time.Second),
		middleware.Heartbeat("/live"),
		middleware.Heartbeat("/ready"),
		render.SetContentType(render.ContentTypeJSON),
	)

	// Endpoints handled by the service
	router.Route(uriPrefix, func(r chi.Router) {
		r.NotFound(svc.handleNotFound)
		r.MethodNotAllowed(svc.handleMethodNotAllowed)

		r.Use(svc.RequestLogger())
		if svc.cfg.apiToken != "" {
			r.Use(authn.BearerTokenAuth(svc.cfg.apiToken))
		}
		r.Use(httprate.Limit(
			svc.cfg.rateLimit, // requests
			time.Second,       // per duration
			httprate.WithLimitHandler(svc.handleRateLimitExceeded),
		))
		r.Use(svc.incReqCounter())

		for _, subRoute := range svc.cfg.subRoutes {
			for _, method := range svc.cfg.methods {
				switch method {
				case http.MethodGet:
					r.Get(subRoute, svc.handleMock)
				case http.MethodPost:
					r.Post(subRoute, svc.handleMock)
				case http.MethodPut:
					r.Put(subRoute, svc.handleMock)
				case http.MethodDelete:
					r.Delete(subRoute, svc.handleMock)
				case http.MethodPatch:
					r.Patch(subRoute, svc.handleMock)
				}
			}
		}

		r.Post("/batch", svc.handleBatchMock)
	})

	svc.router = router
}
