package service

import (
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_loadConfigFromEnv(t *testing.T) {
	type env struct {
		port, apiKey, apiToken, respDelay, failureRespCode, failureRespBody, successRespCode, successRespBody, successRatio, rateLimit, rateExceededRespBody, methods, subRoutes string
	}
	tests := []struct {
		name    string
		env     env
		want    *config
		wantErr bool
	}{
		{
			name: "Valid configuration (defaults)",
			want: &config{
				port:                 8080,
				failureCode:          http.StatusBadRequest,
				failureRespBody:      nil,
				successCode:          http.StatusOK,
				successRespBody:      map[string]interface{}{"success": true},
				successRatio:         1.0,
				rateLimit:            1000,
				rateExceededRespBody: nil,
			},
			wantErr: false,
		},
		{
			name: "Valid configuration",
			env: env{
				port:                 "8080",
				apiToken:             "some-token",
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
			want: &config{
				port:                 8080,
				apiToken:             "some-token",
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
			name: "Both API key and token set",
			env: env{
				apiKey:   "some-key",
				apiToken: "some-token",
			},
			want:    nil,
			wantErr: true,
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
			name: "Invalid response delay",
			env: env{
				respDelay: "not-a-number",
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
			tt.Setenv("API_KEY", test.env.apiKey)
			tt.Setenv("API_TOKEN", test.env.apiToken)
			tt.Setenv("RESP_DELAY", test.env.respDelay)
			tt.Setenv("FAILURE_RESP_CODE", test.env.failureRespCode)
			tt.Setenv("FAILURE_RESP_BODY", test.env.failureRespBody)
			tt.Setenv("SUCCESS_RESP_CODE", test.env.successRespCode)
			tt.Setenv("SUCCESS_RESP_BODY", test.env.successRespBody)
			tt.Setenv("SUCCESS_RATIO", test.env.successRatio)
			tt.Setenv("RATE_LIMIT", test.env.rateLimit)
			tt.Setenv("RATE_EXCEEDED_RESP_BODY", test.env.rateExceededRespBody)
			tt.Setenv("METHODS", test.env.methods)
			tt.Setenv("SUB_ROUTES", test.env.subRoutes)

			got, err := loadConfigFromEnv()
			if (err != nil) != test.wantErr {
				t.Errorf("LoadConfigFromEnv() error = %v, wantErr %v", err, test.wantErr)
			}
			if !cmp.Equal(got, test.want, cmp.AllowUnexported(config{})) {
				t.Errorf("LoadConfigFromEnv() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestStringSliceContains(t *testing.T) {
	type args struct {
		s []string
		e string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Successful search",
			args: args{
				s: []string{"one", "two", "three"},
				e: "two",
			},
			want: true,
		},
		{
			name: "Unsuccessful search",
			args: args{
				s: []string{"one", "two", "three"},
				e: "four",
			},
			want: false,
		},
	}
	t.Parallel()
	for _, testToRun := range tests {
		test := testToRun
		t.Run(test.name, func(tt *testing.T) {
			tt.Parallel()
			if got := stringSliceContains(test.args.s, test.args.e); got != test.want {
				tt.Errorf("StringSliceContains() = %v, want %v", got, test.want)
			}
		})
	}
}

type testStruct struct {
	One   string
	Two   int
	Three bool
	Four  []testSubStruct
	Five  string
}

type testSubStruct struct {
	Six string
}

type jsonStruct struct {
	One   string        `json:"uno"`
	Two   int           `json:"dos"`
	Three bool          `json:"tres"`
	Four  jsonSubStruct `json:"cuatro"`
	Five  string        `json:"cinco"`
}

type jsonSubStruct struct {
	Six string `json:"seis"`
}

func TestStructToMapStringInterface(t *testing.T) {
	type args struct {
		s interface{}
	}

	tests := []struct {
		name    string
		args    args
		want    map[string]interface{}
		wantErr bool
	}{
		{
			name: "Successful transformation",
			args: args{
				s: testStruct{
					One:   "one",
					Two:   2,
					Three: true,
					Four:  nil,
					Five:  "five",
				},
			},
			want: map[string]interface{}{
				"One":   "one",
				"Two":   float64(2),
				"Three": true,
				"Four":  nil,
				"Five":  "five",
			},
			wantErr: false,
		},
		{
			name: "Successful transformation (JSON)",
			args: args{
				s: jsonStruct{
					One:   "one",
					Two:   2,
					Three: true,
					Four: jsonSubStruct{
						Six: "six",
					},
					Five: "five",
				},
			},
			want: map[string]interface{}{
				"uno":  "one",
				"dos":  float64(2),
				"tres": true,
				"cuatro": map[string]interface{}{
					"seis": "six",
				},
				"cinco": "five",
			},
			wantErr: false,
		},
	}
	t.Parallel()
	for _, testToRun := range tests {
		test := testToRun
		t.Run(test.name, func(tt *testing.T) {
			tt.Parallel()
			got, err := structToMapStringInterface(test.args.s)
			if (err != nil) != test.wantErr {
				tt.Errorf("StructToMapStringInterface() error = %v, wantErr %v", err, test.wantErr)
				return
			}
			if !cmp.Equal(got, test.want) {
				tt.Errorf("StructToMapStringInterface() = %v, want %v", got, test.want)
			}
		})
	}
}
