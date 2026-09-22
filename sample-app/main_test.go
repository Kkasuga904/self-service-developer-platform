package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthEndpoints(t *testing.T) {
	for _, path := range []string{"/healthz", "/readyz", "/metrics"} {
		request := httptest.NewRequest(http.MethodGet, path, nil)
		response := httptest.NewRecorder()
		handler().ServeHTTP(response, request)
		if response.Code != http.StatusOK {
			t.Errorf("GET %s status = %d, want 200", path, response.Code)
		}
	}
}

func TestMetricsExposition(t *testing.T) {
	server := handler()
	for range 3 {
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		server.ServeHTTP(httptest.NewRecorder(), request)
	}
	request := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	body := response.Body.String()
	for _, expected := range []string{
		`http_requests_total{method="GET",path="/",code="200"} 3`,
		"http_request_duration_seconds_bucket{",
		`le="+Inf"}`,
		"http_request_duration_seconds_count{",
		"http_request_duration_seconds_sum{",
	} {
		if !strings.Contains(body, expected) {
			t.Errorf("/metrics exposition missing %q in:\n%s", expected, body)
		}
	}
	if strings.Contains(body, "http_request_errors_total{") {
		t.Errorf("/metrics should not report errors before any 5xx response")
	}
}
