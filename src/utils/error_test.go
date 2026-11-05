package utils

import (
	"encoding/json"
	"errors"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

// The helpers signal an expired Academia session by returning an error whose
// text is "invalid response format". The frontend only redirects to /invalid
// when it sees tokenInvalid, so this mapping is the contract between them.
func TestHandleErrorMapsExpiredSession(t *testing.T) {
	cases := []struct {
		err         error
		wantStatus  int
		wantInvalid bool
	}{
		{errors.New("invalid response format"), fiber.StatusUnauthorized, true},
		{errors.New("invalid token format"), fiber.StatusUnauthorized, true},
		{errors.New("dial tcp: connection refused"), fiber.StatusInternalServerError, false},
	}

	for _, c := range cases {
		app := fiber.New(fiber.Config{ErrorHandler: HandleError})
		app.Get("/x", func(*fiber.Ctx) error { return c.err })

		resp, err := app.Test(httptest.NewRequest("GET", "/x", nil))
		if err != nil {
			t.Fatalf("app.Test: %v", err)
		}
		if resp.StatusCode != c.wantStatus {
			t.Errorf("%q: status = %d, want %d", c.err, resp.StatusCode, c.wantStatus)
		}

		body, _ := io.ReadAll(resp.Body)
		var payload map[string]any
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("%q: body %s is not JSON: %v", c.err, body, err)
		}
		if got, _ := payload["tokenInvalid"].(bool); got != c.wantInvalid {
			t.Errorf("%q: tokenInvalid = %v, want %v (body %s)", c.err, got, c.wantInvalid, body)
		}
	}
}
