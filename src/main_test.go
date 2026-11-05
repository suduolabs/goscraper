package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
)

func TestProductionMiddlewarePreservesSessionErrors(t *testing.T) {
	t.Setenv("VALIDATION_KEY", "test-validation-key")
	for _, tc := range []struct {
		name    string
		err     error
		status  int
		invalid bool
	}{
		{"expired session", errors.New("invalid response format"), 401, true},
		{"invalid token", errors.New("invalid token format"), 401, true},
		{"transport failure", errors.New("connection refused"), 500, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			app := newApp()
			app.Get("/test-error", func(*fiber.Ctx) error { return tc.err })
			req := httptest.NewRequest("GET", "/test-error", nil)
			req.Header.Set("X-CSRF-Token", "test-session")
			req.Header.Set("Authorization", "Bearer "+base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d.test-validation-key", time.Now().Unix()))))
			resp, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != tc.status {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tc.status)
			}
			var payload struct {
				TokenInvalid bool `json:"tokenInvalid"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
				t.Fatal(err)
			}
			if payload.TokenInvalid != tc.invalid {
				t.Fatalf("tokenInvalid = %v, want %v", payload.TokenInvalid, tc.invalid)
			}
		})
	}
}
