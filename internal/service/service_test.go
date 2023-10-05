package service

import (
	"net/http"
	"reflect"
	"testing"
)

func Test_LoadConfigFromEnv(t *testing.T) {
	type env struct {
		port, failureRespCode, failureRespBody, successRespCode, successRespBody, successRatio, rateLimit, rateExceededRespBody, methods, subRoutes string
	}
	tests := []struct {
		name    string
		env     env
		want    *SvcConfig
		wantErr bool
	}{
		{
			name: "Valid configuration (defaults)",
			want: &SvcConfig{
				port:                 8080,
				failureCode:          http.StatusBadRequest,
				failureRespBody:      map[string]interface{}{"success": false},
				successCode:          http.StatusOK,
				successRespBody:      map[string]interface{}{"success": true},
				successRatio:         1.0,
				rateLimit:            1000,
				rateExceededRespBody: map[string]interface{}{"success": false, "error": "rate limit exceeded"},
			},
			wantErr: false,
		},
		{
			name: "Valid configuration",
			env: env{
				port:                 "8080",
				failureRespCode:      "400",
				failureRespBody:      `{"success": false}`,
				successRespCode:      "200",
				successRespBody:      `{"success": true}`,
				successRatio:         "0.5",
				rateLimit:            "10",
				rateExceededRespBody: `{"success": false, "error": "rate limit exceeded"}`,
				methods:              "GET,POST,PUT",
				subRoutes:            "/foo,/bar",
			},
			want: &SvcConfig{
				port:                 8080,
				failureCode:          http.StatusBadRequest,
				failureRespBody:      map[string]interface{}{"success": false},
				successCode:          http.StatusOK,
				successRespBody:      map[string]interface{}{"success": true},
				successRatio:         0.5,
				rateLimit:            10,
				rateExceededRespBody: map[string]interface{}{"success": false, "error": "rate limit exceeded"},
				methods:              []string{"GET", "POST", "PUT"},
				subRoutes:            []string{"/foo", "/bar"},
			},
			wantErr: false,
		},
		{
			name: "Invalid port",
			env: env{
				port: "not-a-number",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Invalid failure response code",
			env: env{
				failureRespCode: "not-a-number",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Invalid failure response body",
			env: env{
				failureRespBody: "not-json",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Invalid success response code",
			env: env{
				successRespCode: "not-a-number",
			},
			wantErr: true,
		},
		{
			name: "Invalid success response body",
			env: env{
				successRespBody: "not-json",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Invalid rate limit",
			env: env{
				rateLimit: "not-a-number",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Invalid rate exceeded response body",
			env: env{
				rateExceededRespBody: "not-json",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Invalid success response ratio",
			env: env{
				successRatio: "2.0",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Invalid methods",
			env: env{
				methods: "JUMP,RUN,FLY",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, testToRun := range tests {
		test := testToRun
		t.Run(test.name, func(tt *testing.T) {
			tt.Setenv("PORT", test.env.port)
			tt.Setenv("FAILURE_RESP_CODE", test.env.failureRespCode)
			tt.Setenv("FAILURE_RESP_BODY", test.env.failureRespBody)
			tt.Setenv("SUCCESS_RESP_CODE", test.env.successRespCode)
			tt.Setenv("SUCCESS_RESP_BODY", test.env.successRespBody)
			tt.Setenv("SUCCESS_RATIO", test.env.successRatio)
			tt.Setenv("RATE_LIMIT", test.env.rateLimit)
			tt.Setenv("RATE_EXCEEDED_RESP_BODY", test.env.rateExceededRespBody)
			tt.Setenv("METHODS", test.env.methods)
			tt.Setenv("SUB_ROUTES", test.env.subRoutes)

			got, err := LoadConfigFromEnv()
			if (err != nil) != test.wantErr {
				t.Errorf("LoadConfigFromEnv() error = %v, wantErr %v", err, test.wantErr)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("LoadConfigFromEnv() = %v, want %v", got, test.want)
			}
		})
	}
}
