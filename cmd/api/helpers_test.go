package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/spobly/greenlight/internal/validator"
)

func TestWriteJSON(t *testing.T) {
	// Initialize HTTP recorder (acts like a response writer)
	w := httptest.NewRecorder()

	// Dummy data
	data := envelope{"message": "hello"}

	// Initialize application structure
	app := &application{}

	// Execute function
	err := app.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	// Check response body
	expectedBody := `{
	"message": "hello"
}
`
	if w.Body.String() != expectedBody {
		t.Errorf("Expected body to be %s, got %s", expectedBody, w.Body.String())
	}

	// Check status code
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", w.Code)
	}
}

func TestReadInt(t *testing.T) {
	qs := url.Values{}
	qs.Set("page", "1")
	validator := validator.New()

	app := &application{}
	page := app.readInt(qs, "page", 0, validator)

	if page != 1 {
		t.Errorf("Expected 1, got %d", page)
	}

	qs.Set("page", "not_an_int")
	page = app.readInt(qs, "page", 0, validator)
	if !validator.HasErrors() {
		t.Errorf("Expected an error for 'page' key")
	}
}
