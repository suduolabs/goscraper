package helpers

import (
	"crypto/tls"
	"crypto/x509"
	"goscraper/src/utils"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/valyala/fasthttp"
)

func TestCalendarHTTPClassifiesUpstreamFailures(t *testing.T) {
	for _, tc := range []struct {
		name        string
		status      int
		body        string
		wantInvalid bool
		wantError   bool
	}{
		{"unauthorized", 401, "expired", true, true},
		{"forbidden", 403, "blocked", false, false},
		{"sign-in page", 200, "Please sign in to continue", true, true},
		{"maintenance page", 200, "Scheduled maintenance", false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				for i := 0; i < 12; i++ {
					w.Header().Add("Set-Cookie", "expired="+strings.Repeat("x", 600)+"; Max-Age=0")
				}
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer server.Close()
			roots := x509.NewCertPool()
			roots.AddCert(server.Certificate())
			client := &fasthttp.Client{
				ReadBufferSize: utils.Client.ReadBufferSize,
				TLSConfig:      &tls.Config{RootCAs: roots, ServerName: server.Certificate().DNSNames[0]},
				Dial:           func(string) (net.Conn, error) { return net.Dial("tcp", server.Listener.Addr().String()) },
			}
			original := utils.Client
			utils.Client = client
			defer func() { utils.Client = original; client.CloseIdleConnections() }()
			resp, err := NewCalendarFetcher(time.Now(), "test-session").GetCalendar()
			if (err != nil) != tc.wantError {
				t.Fatalf("error = %v, want error %v", err, tc.wantError)
			}
			invalid := err != nil && strings.Contains(err.Error(), "invalid response format")
			if invalid != tc.wantInvalid {
				t.Fatalf("session invalid = %v, want %v (error %v)", invalid, tc.wantInvalid, err)
			}
			if !tc.wantError && (resp == nil || !resp.Error || resp.Status != tc.status) {
				t.Fatalf("response = %+v, want upstream status %d", resp, tc.status)
			}
		})
	}
}
