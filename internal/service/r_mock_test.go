package service

import (
	"bytes"
	"encoding/json"
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
			failureRespBody: map[string]interface{}{"success": false},
			failureCode:     http.StatusBadRequest,
			successRespBody: map[string]interface{}{"success": true},
			successCode:     http.StatusOK,
			successRatio:    0.5,
			methods:         []string{http.MethodGet},
			subRoutes:       []string{"/foo"},
		},
		logger: newStructuredLogger(),
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
	svc := &service{
		cfg: &SvcConfig{
			failureRespBody: map[string]interface{}{"success": false},
			failureCode:     http.StatusBadRequest,
			successRespBody: map[string]interface{}{"success": true},
			successCode:     http.StatusOK,
			successRatio:    0.5,
			methods:         []string{http.MethodGet},
			subRoutes:       []string{"/foo"},
		},
		logger: newStructuredLogger(),
	}

	tests := []struct {
		name         string
		setupRequest func() *http.Request
		respHandler  func(tt *testing.T, resp *httptest.ResponseRecorder)
	}{
		{
			name: "success batch request",
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
				buf := &bytes.Buffer{}
				enc := json.NewEncoder(buf)
				enc.SetEscapeHTML(true)
				err := enc.Encode([]api.BatchResponse{{Code: 200, Body: "{\"success\":true}"}, {Code: 400, Body: "{\"success\":false}"}})
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
			svc.reqCounter++
			svc.handleBatchMock(resp, test.setupRequest())
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
