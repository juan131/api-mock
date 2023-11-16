package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/go-chi/render"

	"github.com/juan131/api-mock/pkg/api"
)

// incReqCounter implements a simple middleware handler for
// increasing a request counter on every request
func (svc *service) incReqCounter() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			svc.mu.Lock()
			defer svc.mu.Unlock()
			svc.reqCounter++
			next.ServeHTTP(w, r)
		})
	}
}

// handleMock mocks request handling
// Route: /v1/mock/*
func (svc *service) handleMock(w http.ResponseWriter, r *http.Request) {
	// Return failure based on success ratio and requests counter
	if shouldFail(svc.cfg.successRatio, svc.reqCounter) {
		render.Status(r, svc.cfg.failureCode)
		render.JSON(w, r, svc.cfg.failureRespBody)
		return
	}

	render.Status(r, svc.cfg.successCode)
	render.JSON(w, r, svc.cfg.successRespBody)
}

// handleBatchMock mocks batch request handling
// Route: /v1/mock/batch
func (svc *service) handleBatchMock(w http.ResponseWriter, r *http.Request) {
	var requests []api.SingleRequest

	encodedBody, err := io.ReadAll(r.Body)
	if err != nil {
		logID := svc.LogRequestFailure(r, fmt.Sprintf("[handleBatchMock] body reading error: %+v", err), err)
		renderJSON(w, r, http.StatusBadRequest, api.MakeHTTPErrorResponse("body parsing error", api.CodeInvalidBody, logID))
		return
	}

	decodedBody, err := url.ParseQuery(string(encodedBody))
	if err != nil {
		logID := svc.LogRequestFailure(r, fmt.Sprintf("[handleBatchMock] body reading error: %+v", err), err)
		renderJSON(w, r, http.StatusBadRequest, api.MakeHTTPErrorResponse("body parsing error", api.CodeInvalidBody, logID))
		return
	}

	if err := json.Unmarshal([]byte(decodedBody.Get("batch")), &requests); err != nil {
		logID := svc.LogRequestFailure(r, fmt.Sprintf("[handleBatchMock] body reading error: %+v", err), err)
		renderJSON(w, r, http.StatusBadRequest, api.MakeHTTPErrorResponse("body parsing error", api.CodeInvalidBody, logID))
		return
	}

	responses := make([]api.BatchResponse, 0, len(requests))
	for _, r := range requests {
		svc.logger.Info(fmt.Sprintf("Individual request: %v", r))
		if shouldFail(svc.cfg.successRatio, svc.reqCounter) {
			body, err := json.Marshal(svc.cfg.failureRespBody)
			if err != nil {
				continue
			}
			responses = append(responses, api.BatchResponse{
				Code: svc.cfg.failureCode,
				Body: string(body),
			})
		} else {
			body, err := json.Marshal(svc.cfg.successRespBody)
			if err != nil {
				continue
			}
			responses = append(responses, api.BatchResponse{
				Code: svc.cfg.successCode,
				Body: string(body),
			})
		}
		svc.reqCounter++
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, responses)
}

// handleNotFound handles not found requests
func (svc *service) handleNotFound(w http.ResponseWriter, r *http.Request) {
	logID := svc.LogRequestFailure(r, "[handleNotFound] request to "+r.URL.Path, nil)
	renderJSON(w, r, http.StatusNotFound, api.MakeHTTPErrorResponse("not found", api.CodeNotFound, logID))
}

// handleMethodNotAllowed handles not found requests
func (svc *service) handleMethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	logID := svc.LogRequestFailure(r, "[handleMethodNotAllowed] method "+r.Method, nil)
	renderJSON(w, r, http.StatusNotFound, api.MakeHTTPErrorResponse("method not allowed", api.CodeMethodNotAllowed, logID))
}

// handleRateLimitExceeded handles rate limit exceeded requests
func (svc *service) handleRateLimitExceeded(w http.ResponseWriter, r *http.Request) {
	render.Status(r, http.StatusTooManyRequests)
	render.JSON(w, r, svc.cfg.rateExceededRespBody)
}

// shouldFail returns true if the request should fail based
// on a given success ratio and a request counter
func shouldFail(successRatio float64, requestsCounter int) bool {
	failureRatio := 1 - successRatio

	return requestsCounter%int(1/failureRatio) == 0
}

// renderJSON sets HTTP response status code and marshals 'v' to JSON,
// automatically escaping HTML and setting the Content-Type as application/json.
func renderJSON(
	w http.ResponseWriter,
	r *http.Request,
	status int,
	v interface{},
) {
	render.Status(r, status)
	render.JSON(w, r, v)
}
