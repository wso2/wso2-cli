package api

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func TestSummarizeBody(t *testing.T) {
	nginx502 := `<html>
<head><title>502 Bad Gateway</title></head>
<body>
<center><h1>502 Bad Gateway</h1></center>
<hr><center>nginx</center>
</body>
</html>`

	cases := []struct {
		name string
		body string
		want string
	}{
		{"nginx html page uses title", nginx502, "502 Bad Gateway (HTML error page)"},
		{"doctype html without title", "<!DOCTYPE html><body>boom</body>", "(HTML error page)"},
		{"empty body", "   ", "(empty response body)"},
		{"plain text passes through", "upstream timed out", "upstream timed out"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := summarizeBody([]byte(tc.body)); got != tc.want {
				t.Errorf("summarizeBody() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestSummarizeBodyTruncatesLongBodies(t *testing.T) {
	long := strings.Repeat("x", maxErrorBodyChars*2)
	got := summarizeBody([]byte(long))

	if !strings.HasSuffix(got, "… (truncated)") {
		t.Errorf("expected truncation marker, got suffix %q", got[max(0, len(got)-20):])
	}
	if len(got) > maxErrorBodyChars+len("… (truncated)") {
		t.Errorf("summary too long: %d chars", len(got))
	}
}

// Retried requests must re-send the body. Previously the body was read inside
// the retry closure, which drained it on attempt 1, so attempts 2..N arrived
// with an empty body — silently turning a retried write into a no-op payload.
func TestDoWithTokenResendsBodyOnRetry(t *testing.T) {
	var (
		mu       sync.Mutex
		received []string
	)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		mu.Lock()
		received = append(received, string(b))
		n := len(received)
		mu.Unlock()

		// Fail the first attempt at the transport level by hijacking and
		// dropping the connection, forcing retry-go to try again.
		if n == 1 {
			hj, ok := w.(http.Hijacker)
			if !ok {
				t.Error("server does not support hijacking")
				return
			}
			conn, _, err := hj.Hijack()
			if err != nil {
				t.Errorf("hijack: %v", err)
				return
			}
			conn.Close()
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := &IPHTTPClient{client: &http.Client{}}

	const payload = `{"name":"retry-me"}`
	req, err := http.NewRequest("POST", srv.URL, bytes.NewReader([]byte(payload)))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}

	resp, err := c.doWithToken(req, "test-token")
	if err != nil {
		t.Fatalf("doWithToken: %v", err)
	}
	defer resp.Body.Close()

	mu.Lock()
	defer mu.Unlock()
	if len(received) < 2 {
		t.Fatalf("expected at least 2 attempts, got %d", len(received))
	}
	for i, body := range received {
		if body != payload {
			t.Errorf("attempt %d body = %q, want %q", i+1, body, payload)
		}
	}
}
