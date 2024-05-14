package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthcheckHandler(t *testing.T) {
	app := &application{}

	req := httptest.NewRequest(http.MethodGet, "/v1/healthcheck", nil)
	w := httptest.NewRecorder()

	app.healthcheckHandler(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Errorf("expected status OK; got %v", res.Status)
	}

	expectedBody := strings.TrimSpace(`
{
	"status": "available",
	"system_info": {
		"environment": "",
		"version": "(devel)"
	}
}`)
	actualBody := strings.TrimSpace(w.Body.String())

	if actualBody != expectedBody {
		t.Errorf("expected body to be %v; got %v", expectedBody, actualBody)
	}
}
