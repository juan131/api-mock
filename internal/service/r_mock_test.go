package service

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/juan131/api-mock/pkg/api"
)

func Test_service_handleMock(t *testing.T) {
	svc := &service{
		cfg: &SvcConfig{
			failureCode:     http.StatusBadRequest,
			successRespBody: map[string]interface{}{"success": true},
			successCode:     http.StatusOK,
			successRatio:    0.5,
			methods:         []string{http.MethodGet},
			subRoutes:       []string{"/foo"},
		},
		logger: newStructuredLogger(slog.LevelDebug),
	}

	tests := []struct {
		name         string
		setupRequest func() *http.Request
		respHandler  func(tt *testing.T, resp *httptest.ResponseRecorder)
	}{
		{
			name: "success JSON response",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/v1/mock/foo", nil)
				return req
			},
			respHandler: func(tt *testing.T, resp *httptest.ResponseRecorder) {
				if resp.Code != http.StatusOK {
					tt.Errorf("expected status code %d, got %d", http.StatusOK, resp.Code)
				}
				if resp.Header().Get("Content-Type") != "application/json" {
					tt.Errorf("expected content type %s, got %s", "application/json", resp.Header().Get("Content-Type"))
				}
				buf := &bytes.Buffer{}
				enc := json.NewEncoder(buf)
				enc.SetEscapeHTML(true)
				err := enc.Encode(map[string]interface{}{"success": true})
				if err != nil {
					tt.Errorf("could not encode response body: %+v", err)
				}
				if resp.Body.String() != buf.String() {
					tt.Errorf("expected body %s, got %s", buf.String(), resp.Body.String())
				}
			},
		},
		{
			name: "second request must fail given the 0.5 success ratio",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/v1/mock/foo", nil)
				return req
			},
			respHandler: func(tt *testing.T, resp *httptest.ResponseRecorder) {
				if resp.Code != http.StatusBadRequest {
					tt.Errorf("expected status code %d, got %d", http.StatusBadRequest, resp.Code)
				}
				if resp.Header().Get("Content-Type") != "application/json" {
					tt.Errorf("expected content type %s, got %s", "application/json", resp.Header().Get("Content-Type"))
				}
				var errorResponse api.HTTPErrorResponse
				if err := json.Unmarshal(resp.Body.Bytes(), &errorResponse); err != nil {
					tt.Errorf("could not unmarshal response body: %+v", err)
				}
				if errorResponse.Error.Message != "failed request" || errorResponse.Error.Code != api.CodeFailedRequest {
					tt.Errorf("expected error message %s and code %d, got %s and %d", "failed request", api.CodeFailedRequest, errorResponse.Error.Message, errorResponse.Error.Code)
				}
			},
		},
		{
			name: "a third request must succeed again",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/v1/mock/foo", nil)
				return req
			},
			respHandler: func(tt *testing.T, resp *httptest.ResponseRecorder) {
				if resp.Code != http.StatusOK {
					tt.Errorf("expected status code %d, got %d", http.StatusOK, resp.Code)
				}
				if resp.Header().Get("Content-Type") != "application/json" {
					tt.Errorf("expected content type %s, got %s", "application/json", resp.Header().Get("Content-Type"))
				}
				buf := &bytes.Buffer{}
				enc := json.NewEncoder(buf)
				enc.SetEscapeHTML(true)
				err := enc.Encode(map[string]interface{}{"success": true})
				if err != nil {
					tt.Errorf("could not encode response body: %+v", err)
				}
				if resp.Body.String() != buf.String() {
					tt.Errorf("expected body %s, got %s", buf.String(), resp.Body.String())
				}
			},
		},
	}
	for _, testToRun := range tests {
		test := testToRun
		t.Run(test.name, func(tt *testing.T) {
			resp := httptest.NewRecorder()
			svc.reqCounter++
			svc.handleMock(resp, test.setupRequest())
			test.respHandler(tt, resp)
		})
	}
}

func Test_service_handleBatchMock(t *testing.T) {
	tests := []struct {
		name         string
		svc          *service
		setupRequest func() *http.Request
		respHandler  func(tt *testing.T, resp *httptest.ResponseRecorder)
	}{
		{
			name: "success batch request with custom response body",
			svc: &service{
				cfg: &SvcConfig{
					failureCode:     http.StatusBadRequest,
					successRespBody: map[string]interface{}{"success": true},
					successCode:     http.StatusOK,
					successRatio:    0.5,
					methods:         []string{http.MethodGet},
					subRoutes:       []string{"/foo"},
				},
				logger: newStructuredLogger(slog.LevelDebug),
			},
			setupRequest: func() *http.Request {
				reqBody := `[{
					"method": "GET",
					"relative_url": "/foo",
					"body": null
				}, {
					"method": "GET",
					"relative_url": "/bar",
					"body": null
				}]`
				encodedBody := url.Values{}
				encodedBody.Set("batch", reqBody)

				req := httptest.NewRequest(
					http.MethodPost,
					"/v1/mock/batch",
					strings.NewReader(encodedBody.Encode()),
				)
				return req
			},
			respHandler: func(tt *testing.T, resp *httptest.ResponseRecorder) {
				if resp.Code != http.StatusOK {
					tt.Errorf("expected status code %d, got %d", http.StatusOK, resp.Code)
				}
				if resp.Header().Get("Content-Type") != "application/json" {
					tt.Errorf("expected content type %s, got %s", "application/json", resp.Header().Get("Content-Type"))
				}
				var batchResponse []api.BatchResponse
				if err := json.Unmarshal(resp.Body.Bytes(), &batchResponse); err != nil {
					tt.Errorf("could not unmarshal response body: %+v", err)
				}

				if len(batchResponse) != 2 {
					tt.Errorf("expected batch response length %d, got %d", 2, len(batchResponse))
				}

				if batchResponse[0].Code != 200 || batchResponse[0].Body != "{\"success\":true}" {
					tt.Errorf("expected response code %d and body %s, got %d and %s", 200, "{\"success\":true}", batchResponse[0].Code, batchResponse[0].Body)
				}

				if batchResponse[1].Code != 400 {
					tt.Errorf("expected response code %d, got %d", 400, batchResponse[1].Code)
				}

				var errorResponse api.HTTPErrorResponse
				if err := json.Unmarshal([]byte(batchResponse[1].Body), &errorResponse); err != nil {
					tt.Errorf("could not unmarshal response body: %+v", err)
				}

				if errorResponse.Error.Message != "failed request" || errorResponse.Error.Code != api.CodeFailedRequest {
					tt.Errorf("expected error message %s and code %d, got %s and %d", "failed request", api.CodeFailedRequest, errorResponse.Error.Message, errorResponse.Error.Code)
				}
			},
		},
		{
			name: "success batch request with custom failure response body",
			svc: &service{
				cfg: &SvcConfig{
					failureRespBody: map[string]interface{}{"success": false},
					failureCode:     http.StatusBadRequest,
					successRespBody: map[string]interface{}{"success": true},
					successCode:     http.StatusOK,
					successRatio:    0.5,
					methods:         []string{http.MethodGet},
					subRoutes:       []string{"/foo"},
				},
				logger: newStructuredLogger(slog.LevelDebug),
			},
			setupRequest: func() *http.Request {
				reqBody := `[{
					"method": "GET",
					"relative_url": "/foo",
					"body": null
				}, {
					"method": "GET",
					"relative_url": "/bar",
					"body": null
				}]`
				encodedBody := url.Values{}
				encodedBody.Set("batch", reqBody)

				req := httptest.NewRequest(
					http.MethodPost,
					"/v1/mock/batch",
					strings.NewReader(encodedBody.Encode()),
				)
				return req
			},
			respHandler: func(tt *testing.T, resp *httptest.ResponseRecorder) {
				if resp.Code != http.StatusOK {
					tt.Errorf("expected status code %d, got %d", http.StatusOK, resp.Code)
				}
				if resp.Header().Get("Content-Type") != "application/json" {
					tt.Errorf("expected content type %s, got %s", "application/json", resp.Header().Get("Content-Type"))
				}
				var batchResponse []api.BatchResponse
				if err := json.Unmarshal(resp.Body.Bytes(), &batchResponse); err != nil {
					tt.Errorf("could not unmarshal response body: %+v", err)
				}

				if batchResponse[0].Code != 200 || batchResponse[0].Body != "{\"success\":true}" {
					tt.Errorf("expected response code %d and body %s, got %d and %s", 200, "{\"success\":true}", batchResponse[0].Code, batchResponse[0].Body)
				}

				if batchResponse[1].Code != 400 || batchResponse[1].Body != "{\"success\":false}" {
					tt.Errorf("expected response code %d and body %s, got %d and %s", 400, "{\"success\":false}", batchResponse[1].Code, batchResponse[1].Body)
				}
			},
		},
	}
	t.Parallel()
	for _, testToRun := range tests {
		test := testToRun
		t.Run(test.name, func(tt *testing.T) {
			tt.Parallel()
			resp := httptest.NewRecorder()
			test.svc.reqCounter++
			test.svc.handleBatchMock(resp, test.setupRequest())
			test.respHandler(tt, resp)
		})
	}
}

func Test_handleNotFound(t *testing.T) {
	svc := &service{
		logger: newStructuredLogger(slog.LevelDebug),
	}
	tests := []struct {
		name         string
		setupRequest func() *http.Request
		respHandler  func(tt *testing.T, resp *httptest.ResponseRecorder)
	}{
		{
			name: "request to unknown route",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/v1/mock/unknown", nil)
				return req
			},
			respHandler: func(tt *testing.T, resp *httptest.ResponseRecorder) {
				if resp.Code != http.StatusNotFound {
					tt.Errorf("expected status code %d, got %d", http.StatusNotFound, resp.Code)
				}
				if resp.Header().Get("Content-Type") != "application/json" {
					tt.Errorf("expected content type %s, got %s", "application/json", resp.Header().Get("Content-Type"))
				}
				var errorResponse api.HTTPErrorResponse
				if err := json.Unmarshal(resp.Body.Bytes(), &errorResponse); err != nil {
					tt.Errorf("could not unmarshal response body: %+v", err)
				}

				if errorResponse.Error.Message != "not found" || errorResponse.Error.Code != api.CodeNotFound {
					tt.Errorf("expected error message %s and code %d, got %s and %d", "not found", api.CodeNotFound, errorResponse.Error.Message, errorResponse.Error.Code)
				}
			},
		},
	}
	t.Parallel()
	for _, testToRun := range tests {
		test := testToRun
		t.Run(test.name, func(tt *testing.T) {
			tt.Parallel()
			resp := httptest.NewRecorder()
			svc.handleNotFound(resp, test.setupRequest())
			test.respHandler(tt, resp)
		})
	}
}

func Test_handleMethodNotAllowed(t *testing.T) {
	svc := &service{
		logger: newStructuredLogger(slog.LevelDebug),
	}
	tests := []struct {
		name         string
		setupRequest func() *http.Request
		respHandler  func(tt *testing.T, resp *httptest.ResponseRecorder)
	}{
		{
			name: "request to unknown route",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/v1/mock/unknown", nil)
				return req
			},
			respHandler: func(tt *testing.T, resp *httptest.ResponseRecorder) {
				if resp.Code != http.StatusNotFound {
					tt.Errorf("expected status code %d, got %d", http.StatusNotFound, resp.Code)
				}
				if resp.Header().Get("Content-Type") != "application/json" {
					tt.Errorf("expected content type %s, got %s", "application/json", resp.Header().Get("Content-Type"))
				}
				var errorResponse api.HTTPErrorResponse
				if err := json.Unmarshal(resp.Body.Bytes(), &errorResponse); err != nil {
					tt.Errorf("could not unmarshal response body: %+v", err)
				}

				if errorResponse.Error.Message != "method not allowed" || errorResponse.Error.Code != api.CodeMethodNotAllowed {
					tt.Errorf("expected error message %s and code %d, got %s and %d", "method not allowed", api.CodeMethodNotAllowed, errorResponse.Error.Message, errorResponse.Error.Code)
				}
			},
		},
	}
	t.Parallel()
	for _, testToRun := range tests {
		test := testToRun
		t.Run(test.name, func(tt *testing.T) {
			tt.Parallel()
			resp := httptest.NewRecorder()
			svc.handleMethodNotAllowed(resp, test.setupRequest())
			test.respHandler(tt, resp)
		})
	}
}

func Test_handleRateLimitExceeded(t *testing.T) {
	tests := []struct {
		name         string
		svc          *service
		setupRequest func() *http.Request
		respHandler  func(tt *testing.T, resp *httptest.ResponseRecorder)
	}{
		{
			name: "request limit exceeded with default response body",
			svc: &service{
				logger: newStructuredLogger(slog.LevelDebug),
				cfg:    &SvcConfig{},
			},
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/v1/mock/unknown", nil)
				return req
			},
			respHandler: func(tt *testing.T, resp *httptest.ResponseRecorder) {
				if resp.Code != http.StatusTooManyRequests {
					tt.Errorf("expected status code %d, got %d", http.StatusTooManyRequests, resp.Code)
				}
				if resp.Header().Get("Content-Type") != "application/json" {
					tt.Errorf("expected content type %s, got %s", "application/json", resp.Header().Get("Content-Type"))
				}
				var errorResponse api.HTTPErrorResponse
				if err := json.Unmarshal(resp.Body.Bytes(), &errorResponse); err != nil {
					tt.Errorf("could not unmarshal response body: %+v", err)
				}

				if errorResponse.Error.Message != "rate limit exceeded" || errorResponse.Error.Code != api.CodeRateLimitExceeded {
					tt.Errorf("expected error message %s and code %d, got %s and %d", "rate limit exceeded", api.CodeRateLimitExceeded, errorResponse.Error.Message, errorResponse.Error.Code)
				}
			},
		},
		{
			name: "request limit exceeded with custom response body",
			svc: &service{
				logger: newStructuredLogger(slog.LevelDebug),
				cfg: &SvcConfig{
					rateExceededRespBody: map[string]interface{}{"success": false},
				},
			},
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/v1/mock/unknown", nil)
				return req
			},
			respHandler: func(tt *testing.T, resp *httptest.ResponseRecorder) {
				if resp.Code != http.StatusTooManyRequests {
					tt.Errorf("expected status code %d, got %d", http.StatusTooManyRequests, resp.Code)
				}
				if resp.Header().Get("Content-Type") != "application/json" {
					tt.Errorf("expected content type %s, got %s", "application/json", resp.Header().Get("Content-Type"))
				}
				buf := &bytes.Buffer{}
				enc := json.NewEncoder(buf)
				enc.SetEscapeHTML(true)
				err := enc.Encode(map[string]interface{}{"success": false})
				if err != nil {
					tt.Errorf("could not encode response body: %+v", err)
				}
				if resp.Body.String() != buf.String() {
					tt.Errorf("expected body %s, got %s", buf.String(), resp.Body.String())
				}
			},
		},
	}
	t.Parallel()
	for _, testToRun := range tests {
		test := testToRun
		t.Run(test.name, func(tt *testing.T) {
			tt.Parallel()
			resp := httptest.NewRecorder()
			test.svc.handleRateLimitExceeded(resp, test.setupRequest())
			test.respHandler(tt, resp)
		})
	}
}

func Test_shouldFail(t *testing.T) {
	type args struct {
		successRatio    float64
		requestsCounter int
	}

	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "should not fail given the success ratio is 0.8 and the request counter is 1",
			args: args{
				successRatio:    0.8,
				requestsCounter: 1,
			},
			want: false,
		},
		{
			name: "should not fail given the success ratio is 0.8 and the request counter is 2",
			args: args{
				successRatio:    0.8,
				requestsCounter: 2,
			},
			want: false,
		},
		{
			name: "should not fail given the success ratio is 0.8 and the request counter is 3",
			args: args{
				successRatio:    0.8,
				requestsCounter: 3,
			},
			want: false,
		},
		{
			name: "should not fail given the success ratio is 0.8 and the request counter is 4",
			args: args{
				successRatio:    0.8,
				requestsCounter: 4,
			},
			want: false,
		},
		{
			name: "should fail given the success ratio is 0.8 and the request counter is 5",
			args: args{
				successRatio:    0.8,
				requestsCounter: 5,
			},
			want: true,
		},
		{
			name: "should not fail given the success ratio is 0.8 and the request counter is 6",
			args: args{
				successRatio:    0.8,
				requestsCounter: 6,
			},
			want: false,
		},
		{
			name: "should not fail given the success ratio is 0.8 and the request counter is 7",
			args: args{
				successRatio:    0.8,
				requestsCounter: 7,
			},
			want: false,
		},
		{
			name: "should not fail given the success ratio is 0.8 and the request counter is 8",
			args: args{
				successRatio:    0.8,
				requestsCounter: 8,
			},
			want: false,
		},
		{
			name: "should not fail given the success ratio is 0.8 and the request counter is 9",
			args: args{
				successRatio:    0.8,
				requestsCounter: 9,
			},
			want: false,
		},
		{
			name: "should fail given the success ratio is 0.8 and the request counter is 10",
			args: args{
				successRatio:    0.8,
				requestsCounter: 10,
			},
			want: true,
		},
	}
	t.Parallel()
	for _, testToRun := range tests {
		test := testToRun
		t.Run(test.name, func(tt *testing.T) {
			tt.Parallel()
			if got := shouldFail(test.args.successRatio, test.args.requestsCounter); got != test.want {
				tt.Errorf("shouldFail() = %v, want %v", got, test.want)
			}
		})
	}
}
