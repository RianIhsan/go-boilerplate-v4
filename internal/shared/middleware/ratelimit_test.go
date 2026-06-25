package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/RianIhsan/go-boilerplate-v4/internal/shared/middleware"
)

func TestRateLimit(t *testing.T) {
	const limit = 2

	mw := middleware.RateLimit(limit, time.Minute)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := range limit {
		req := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("request %d: status = %d, want %d", i+1, rr.Code, http.StatusOK)
		}
	}

	req := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTooManyRequests {
		t.Fatalf("request over limit: status = %d, want %d", rr.Code, http.StatusTooManyRequests)
	}

	var body struct {
		Success bool `json:"success"`
		Error   struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode rate-limited response body: %v", err)
	}
	if body.Success {
		t.Error("rate-limited response success = true, want false")
	}
	if body.Error.Code != "TOO_MANY_REQUESTS" {
		t.Errorf("rate-limited response error.code = %q, want %q", body.Error.Code, "TOO_MANY_REQUESTS")
	}
}
