package api

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/henvic/httpretty"
)

const LOG_FILE = "http-logger.log"

// Headers that carry credentials and must never reach the trace log.
var sensitiveHeaders = []string{
	"Authorization",
	"Proxy-Authorization",
	"Cookie",
	"Set-Cookie",
	"X-Api-Key",
	"X-Auth-Token",
}

// skipCredentialBodies suppresses request/response bodies that carry tokens.
// The OAuth token endpoints post application/x-www-form-urlencoded bodies
// containing subject_token / refresh_token, and answer with JSON containing
// access_token and refresh_token.
func skipCredentialBodies(h http.Header) (bool, error) {
	return strings.Contains(h.Get("Content-Type"), "application/x-www-form-urlencoded"), nil
}

func newLogger(logBodies bool) *httpretty.Logger {
	logger := &httpretty.Logger{
		Time:            true,
		TLS:             false,
		Colors:          true,
		RequestHeader:   true,
		RequestBody:     logBodies,
		ResponseHeader:  true,
		ResponseBody:    logBodies,
		Formatters:      []httpretty.Formatter{&httpretty.JSONFormatter{}},
		MaxResponseBody: 10000,
	}

	// Redact credential-bearing headers. Without this, every traced request
	// prints its full "Authorization: Bearer <token>".
	logger.SkipHeader(sensitiveHeaders)
	logger.SetBodyFilter(skipCredentialBodies)

	if os.Getenv("LOG_MODE") == "file" {
		// 0600: the trace log can still contain sensitive payloads, so it must
		// not be world-readable.
		file, err := os.OpenFile(LOG_FILE, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
		if err != nil {
			// Leave the logger writing to its default (stderr) rather than
			// calling SetOutput with a nil file, which would panic on write.
			fmt.Fprintf(os.Stderr, "Error while opening log file %q: %s\n", LOG_FILE, err)
		} else {
			logger.SetOutput(file)
		}
	}

	return logger
}

// CreateLogger returns a tracing RoundTripper for general API traffic.
// Credential headers are redacted; bodies are logged.
func CreateLogger() func(http.RoundTripper) http.RoundTripper {
	return newLogger(true).RoundTripper
}

// CreateAuthLogger returns a tracing RoundTripper for endpoints that exchange
// credentials (OAuth token, token refresh, PAT introspection). Bodies are
// suppressed entirely because both the request and the response carry tokens.
func CreateAuthLogger() func(http.RoundTripper) http.RoundTripper {
	return newLogger(false).RoundTripper
}
