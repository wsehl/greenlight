package main

import (
	"encoding/json"
	"log/slog"

	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"sync"
	"testing"

	"github.com/spobly/greenlight/internal/config"
)

func TestRateLimitConcurrent(t *testing.T) {
	cfg := &config.Config{
		Limiter: struct {
			RPS     float64
			Burst   int
			Enabled bool
		}{
			RPS:     1.0,
			Burst:   5,
			Enabled: true,
		},
	}

	app := &application{
		config: *cfg,
	}

	testHandler := app.rateLimit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	ts := httptest.NewServer(testHandler)
	defer ts.Close()

	var wg sync.WaitGroup
	var lock sync.Mutex
	successCount := 0
	failCount := 0

	// Simulate 10 concurrent requests
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp, err := http.Get(ts.URL)
			if err != nil {
				t.Fatalf("could not make GET request: %v", err)
			}
			defer resp.Body.Close()

			lock.Lock()
			if resp.StatusCode == http.StatusOK {
				successCount++
			} else if resp.StatusCode == http.StatusTooManyRequests {
				failCount++
			}
			lock.Unlock()
		}()
	}

	wg.Wait()

	if successCount > 5 {
		t.Errorf("expected at most 5 successful requests due to burst rate; got %d", successCount)
	}
	if failCount < 5 {
		t.Errorf("expected at least 5 failures due to rate limiting; got %d", failCount)
	}
}

func TestRecoverPanic(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	app := &application{
		logger: logger,
	}

	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("something went wrong")
	})

	testHandler := app.recoverPanic(panicHandler)

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	testHandler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusInternalServerError)
	}

	expected := map[string]interface{}{
		"error": "the server encountered a problem and could not process your request",
	}

	var actual map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &actual); err != nil {
		t.Fatal("could not unmarshal response:", err)
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("handler returned unexpected body: got %v want %v",
			actual, expected)
	}

	if contentType := rr.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("expected content type application/json; got %s", contentType)
	}
}
