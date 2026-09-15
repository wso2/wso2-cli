package api

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const (
	fakeBearer  = "eyJhbGciOiJSUzI1NiJ9.FAKE-ACCESS-TOKEN-SHOULD-NEVER-BE-LOGGED"
	fakeRefresh = "FAKE-REFRESH-TOKEN-SHOULD-NEVER-BE-LOGGED"
)

// traceRequest drives one request through the general-purpose trace logger
// (the one CreateLogger returns) and returns everything the logger wrote.
func traceRequest(t *testing.T, contentType, reqBody, respBody string) string {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.Copy(io.Discard, r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(respBody))
	}))
	defer srv.Close()

	var buf bytes.Buffer
	logger := newLogger(true)
	logger.SetOutput(&buf)

	req, err := http.NewRequest("POST", srv.URL, strings.NewReader(reqBody))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+fakeBearer)
	req.Header.Set("Content-Type", contentType)

	client := &http.Client{Transport: logger.RoundTripper(nil)}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	return buf.String()
}

// The Authorization header must never appear in trace output. Before this was
// fixed, TRACE_ENABLED=true printed the full bearer token on every request.
func TestTraceLoggerRedactsAuthorizationHeader(t *testing.T) {
	out := traceRequest(t, "application/json", `{"hello":"world"}`, `{"ok":true}`)

	if out == "" {
		t.Fatal("logger produced no output; test would vacuously pass")
	}
	if strings.Contains(out, fakeBearer) {
		t.Error("bearer token leaked into trace log")
	}
	if strings.Contains(out, "Bearer ") {
		t.Error("Authorization header value leaked into trace log")
	}
}

// OAuth token exchanges post form-urlencoded bodies containing subject_token /
// refresh_token. Those bodies must be suppressed.
func TestTraceLoggerSkipsFormEncodedCredentialBodies(t *testing.T) {
	body := "grant_type=refresh_token&refresh_token=" + fakeRefresh
	out := traceRequest(t, "application/x-www-form-urlencoded", body, `{"ok":true}`)

	if strings.Contains(out, fakeRefresh) {
		t.Error("refresh token from form body leaked into trace log")
	}
}

// The auth logger additionally suppresses bodies in both directions, because
// the token endpoint's JSON *response* carries access_token / refresh_token.
func TestAuthLoggerSuppressesTokenResponseBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"access_token":"` + fakeBearer + `","refresh_token":"` + fakeRefresh + `"}`))
	}))
	defer srv.Close()

	var buf bytes.Buffer
	logger := newLogger(false) // what CreateAuthLogger uses
	logger.SetOutput(&buf)

	req, _ := http.NewRequest("POST", srv.URL, strings.NewReader("grant_type=refresh_token"))
	req.Header.Set("Authorization", "Bearer "+fakeBearer)

	client := &http.Client{Transport: logger.RoundTripper(nil)}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	out := buf.String()
	if strings.Contains(out, fakeBearer) {
		t.Error("access token leaked into auth trace log")
	}
	if strings.Contains(out, fakeRefresh) {
		t.Error("refresh token leaked into auth trace log")
	}
}

func TestSkipCredentialBodies(t *testing.T) {
	cases := []struct {
		contentType string
		wantSkip    bool
	}{
		{"application/x-www-form-urlencoded", true},
		{"application/x-www-form-urlencoded; charset=utf8", true},
		{"application/json", false},
		{"", false},
	}
	for _, tc := range cases {
		h := http.Header{}
		if tc.contentType != "" {
			h.Set("Content-Type", tc.contentType)
		}
		got, err := skipCredentialBodies(h)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != tc.wantSkip {
			t.Errorf("content-type %q: skip=%v, want %v", tc.contentType, got, tc.wantSkip)
		}
	}
}
