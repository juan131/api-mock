package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const (
	defaultPort      int     = 8080
	defaultRateLimit int     = 1000
	defaultRatio     float64 = 1.0
)

// config is the service configuration
type config struct {
	port                 int                    // server listening port
	apiKey               string                 // api key
	apiToken             string                 // api token
	methods, subRoutes   []string               // supported sub-routes
	respDelay            time.Duration          // response delay in milliseconds
	failureCode          int                    // response code for failed requests
	failureRespBody      map[string]interface{} // response body for failed requests
	successCode          int                    // response code for successful requests
	successRespBody      map[string]interface{} // response body for successful requests
	successRatio         float64                // ratio of successful requests
	rateLimit            int                    // rate limit (requests per second)
	rateExceededRespBody map[string]interface{} // response body for rate exceeded requests
}

// loadConfigFromEnv loads the configuration from the environment.
//
//nolint:cyclop // many env variables to parse
func loadConfigFromEnv() (*config, error) {
	var err error
	cfg := config{
		apiKey:   os.Getenv("API_KEY"),
		apiToken: os.Getenv("API_TOKEN"),
	}

	if cfg.apiKey != "" && cfg.apiToken != "" {
		return nil, errors.New("only one of API_KEY or API_TOKEN can be set")
	}

	portENV := os.Getenv("PORT")
	if portENV != "" {
		cfg.port, err = strconv.Atoi(portENV)
		if err != nil {
			return nil, fmt.Errorf("invalid int format for PORT: %w", err)
		}
	} else {
		cfg.port = defaultPort
	}

	respDelayENV := os.Getenv("RESP_DELAY")
	if respDelayENV != "" {
		respDelayINT, err := strconv.Atoi(respDelayENV)
		if err != nil {
			return nil, fmt.Errorf("invalid int format for RESP_DELAY: %w", err)
		}

		cfg.respDelay = time.Duration(respDelayINT) * time.Millisecond
		if cfg.respDelay > 30*time.Second {
			return nil, fmt.Errorf("RESP_DELAY cannot be greater than 30 seconds")
		}
	}

	failureCodeEnv := os.Getenv("FAILURE_RESP_CODE")
	if failureCodeEnv != "" {
		cfg.failureCode, err = strconv.Atoi(failureCodeEnv)
		if err != nil {
			return nil, fmt.Errorf("invalid int format for FAILURE_RESP_CODE: %w", err)
		}
	} else {
		cfg.failureCode = http.StatusBadRequest // default response code
	}

	failureRespBodyEnv := os.Getenv("FAILURE_RESP_BODY")
	if failureRespBodyEnv != "" {
		if err = json.Unmarshal([]byte(failureRespBodyEnv), &cfg.failureRespBody); err != nil {
			return nil, fmt.Errorf("invalid json format for FAILURE_RESP_BODY: %w", err)
		}
	}

	successCodeEnv := os.Getenv("SUCCESS_RESP_CODE")
	if successCodeEnv != "" {
		cfg.successCode, err = strconv.Atoi(successCodeEnv)
		if err != nil {
			return nil, fmt.Errorf("invalid int format for SUCCESS_RESP_CODE: %w", err)
		}
	} else {
		cfg.successCode = http.StatusOK // default response code
	}

	successRepBodyEnv := os.Getenv("SUCCESS_RESP_BODY")
	if successRepBodyEnv != "" {
		if err = json.Unmarshal([]byte(successRepBodyEnv), &cfg.successRespBody); err != nil {
			return nil, fmt.Errorf("invalid json format for SUCCESS_RESP_BODY: %w", err)
		}
	} else {
		cfg.successRespBody = map[string]interface{}{"success": true} // default response body
	}

	successRatioEnv := os.Getenv("SUCCESS_RATIO")
	if successRatioEnv != "" {
		cfg.successRatio, err = strconv.ParseFloat(successRatioEnv, 64)
		if err != nil || cfg.successRatio <= 0 || cfg.successRatio > 1 {
			return nil, fmt.Errorf("invalid value for SUCCESS_RATIO")
		}
	} else {
		cfg.successRatio = defaultRatio
	}

	rateLimitEnv := os.Getenv("RATE_LIMIT")
	if rateLimitEnv != "" {
		cfg.rateLimit, err = strconv.Atoi(rateLimitEnv)
		if err != nil {
			return nil, fmt.Errorf("invalid int format for RATE_LIMIT")
		}
	} else {
		cfg.rateLimit = defaultRateLimit
	}

	rateExceededRespBodyEnv := os.Getenv("RATE_EXCEEDED_RESP_BODY")
	if rateExceededRespBodyEnv != "" {
		if err = json.Unmarshal([]byte(rateExceededRespBodyEnv), &cfg.rateExceededRespBody); err != nil {
			return nil, fmt.Errorf("invalid json format for RATE_EXCEEDED_RESP_BODY: %w", err)
		}
	}

	methodsEnv := os.Getenv("METHODS")
	if methodsEnv != "" {
		allowedMethods := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}
		cfg.methods = strings.Split(methodsEnv, ",")
		for _, method := range cfg.methods {
			if !stringSliceContains(allowedMethods, method) {
				return nil, fmt.Errorf("method %s is not allowed", method)
			}
		}
	}

	subRoutesEnv := os.Getenv("SUB_ROUTES")
	if subRoutesEnv != "" {
		cfg.subRoutes = strings.Split(subRoutesEnv, ",")
	}

	return &cfg, nil
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

// structToMapStringInterface transforms a struct of the given type into a map[string]interface{}
func structToMapStringInterface(s interface{}) (map[string]interface{}, error) {
	if reflect.TypeOf(s).Kind() != reflect.Struct {
		return nil, fmt.Errorf("s must be a struct")
	}

	buf, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}

	var mapVal map[string]interface{}
	if err = json.Unmarshal(buf, &mapVal); err != nil {
		return nil, err
	}

	return mapVal, nil
}
