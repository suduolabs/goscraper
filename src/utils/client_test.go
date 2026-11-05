package utils

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/valyala/fasthttp"
)

func TestClientReadsLargeSessionExpiryHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for i := 0; i < 12; i++ {
			w.Header().Add("Set-Cookie", "expired="+strings.Repeat("x", 600)+"; Max-Age=0; Path=/")
		}
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("Please sign in to continue"))
	}))
	defer server.Close()

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)
	req.SetRequestURI(server.URL)
	if err := Client.Do(req, resp); err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode() != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode())
	}
	if string(resp.Body()) != "Please sign in to continue" {
		t.Fatalf("unexpected body: %q", resp.Body())
	}
}
