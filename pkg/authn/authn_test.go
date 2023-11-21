package authn

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
)

const defaultKey string = "some-key"
const defaultToken string = "some-token"

func TestApiKeyAuth(t *testing.T) {
	tests := []struct {
		name       string
		authHeader string
		want       int
	}{
		{
			"valid auth",
			defaultKey,
			http.StatusOK,
		},
		{
			"missing key",
			"",
			http.StatusUnauthorized,
		},
		{
			"invalid key",
			"foo",
			http.StatusUnauthorized,
		},
	}
	t.Parallel()
	for _, testToRun := range tests {
		test := testToRun
		t.Run(test.name, func(tt *testing.T) {
			tt.Parallel()

			r := chi.NewRouter()
			r.Use(ApiKeyAuth(defaultKey))
			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("X-API-KEY", test.authHeader)
			recorder := httptest.NewRecorder()
			r.ServeHTTP(recorder, req)
			res := recorder.Result()
			defer res.Body.Close()

			if res.StatusCode != test.want {
				tt.Errorf("response status code is incorrect, got %d, want %d", res.StatusCode, test.want)
			}
		})
	}
}

func TestBearerTokenAuth(t *testing.T) {
	tests := []struct {
		name       string
		authHeader string
		want       int
	}{
		{
			"valid auth",
			fmt.Sprintf("Bearer %s", defaultToken),
			http.StatusOK,
		},
		{
			"missing token",
			"",
			http.StatusUnauthorized,
		},
		{
			"invalid token",
			"Bearer foo",
			http.StatusUnauthorized,
		},
	}
	t.Parallel()
	for _, testToRun := range tests {
		test := testToRun
		t.Run(test.name, func(tt *testing.T) {
			tt.Parallel()

			r := chi.NewRouter()
			r.Use(BearerTokenAuth(defaultToken))
			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Authorization", test.authHeader)
			recorder := httptest.NewRecorder()
			r.ServeHTTP(recorder, req)
			res := recorder.Result()
			defer res.Body.Close()

			if res.StatusCode != test.want {
				tt.Errorf("response status code is incorrect, got %d, want %d", res.StatusCode, test.want)
			}
		})
	}
}
