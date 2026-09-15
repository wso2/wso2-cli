package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/avast/retry-go"
)

const (
	NO_OF_RETRIES = 3
	RETRY_DELAY   = 3000 * time.Millisecond

	// Cap on how much of an unrecognized error body is echoed to the user.
	maxErrorBodyChars = 300
)

// summarizeBody makes a non-JSON error body fit for a CLI error message.
// Gateways return full HTML error pages (e.g. an nginx 502), and dumping one
// verbatim buries the actual status code in markup.
func summarizeBody(body []byte) string {
	s := strings.TrimSpace(string(body))
	if s == "" {
		return "(empty response body)"
	}

	lower := strings.ToLower(s)
	if strings.HasPrefix(lower, "<!doctype html") || strings.HasPrefix(lower, "<html") {
		// Prefer the <title>, which is where gateways put the useful summary.
		if i := strings.Index(lower, "<title>"); i != -1 {
			if j := strings.Index(lower[i:], "</title>"); j != -1 {
				if title := strings.TrimSpace(s[i+len("<title>") : i+j]); title != "" {
					return fmt.Sprintf("%s (HTML error page)", title)
				}
			}
		}
		return "(HTML error page)"
	}

	if len(s) > maxErrorBodyChars {
		return s[:maxErrorBodyChars] + "… (truncated)"
	}
	return s
}

type IPHTTPClient struct {
	client       *http.Client
	tokenStore   ReadOnlyTokenStore
	traceEnabled bool
}

func NewIPHTTPClient(tokenStore ReadOnlyTokenStore) *IPHTTPClient {
	c := &IPHTTPClient{
		client:       &http.Client{},
		tokenStore:   tokenStore,
		traceEnabled: os.Getenv("TRACE_ENABLED") == "true",
	}

	if c.traceEnabled {
		c.client.Transport = CreateLogger()(c.client.Transport)
	}

	return c
}

// DoForActiveOrg makes an authenticated request using the active session token.
// Prefer this over Do() when the org ID for token lookup is unknown or is a UUID
// rather than the integer org ID expected by the token store.
func (c *IPHTTPClient) DoForActiveOrg(req *http.Request) (*http.Response, error) {
	token, err := c.tokenStore.GetTokenForActiveOrg()
	if err != nil {
		return nil, fmt.Errorf("error while retrieving access token for active org: %w", err)
	}
	return c.doWithToken(req, token)
}

func (c *IPHTTPClient) Do(req *http.Request, orgId string) (*http.Response, error) {
	// Get the access token for the active org
	// If the token is expired, it will be automatically refreshed & returned
	token, err := c.tokenStore.GetTokenForOrg(orgId, false)
	if err != nil {
		return nil, fmt.Errorf("error while retrieving access token: %w", err)
	}
	return c.doWithToken(req, token)
}

func (c *IPHTTPClient) doWithToken(req *http.Request, token string) (*http.Response, error) {
	var err error
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("User-Agent", "Integration Platform CLI")
	req.Header.Set("x-origin-cloud", "devant")
	req.Header.Add("Authorization", "Bearer "+token)

	// Buffer the request body once, up front. Reading it inside the retry
	// closure drained it on the first attempt, so every retry re-sent the
	// request with an empty body.
	var bodyBytes []byte
	if req.Body != nil {
		bodyBytes, err = io.ReadAll(req.Body)
		req.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("reading request body: %w", err)
		}
	}

	var resp *http.Response
	err = retry.Do(
		func() error {
			if bodyBytes != nil {
				req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			}

			resp, err = c.client.Do(req)
			if err != nil {
				return err
			}
			return nil
		},
		retry.Attempts(NO_OF_RETRIES),
		retry.Delay(RETRY_DELAY),
		retry.OnRetry(func(n uint, err error) {
			// Must be stderr: in MCP stdio mode stdout carries the JSON-RPC
			// stream, and writing to it corrupts the session.
			fmt.Fprintf(os.Stderr, "Retry attempt #%d failed with error: %s\n", n, err)
		}),
	)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 300 {
		// Every path below returns an error, and no caller reads the response
		// on error — they only match with errors.Is. Close the body so the
		// connection is released; leaking it on each 401/5xx accumulates
		// sockets in the long-lived MCP server.
		defer resp.Body.Close()

		if resp.StatusCode == 401 {
			// tokenStore.GetTokenForActiveOrg() will automatically refresh the token if it is expired
			// So, if we reach here, it means that the token is not valid for the active org
			return nil, ErrTokenNotValid
		} else if resp.StatusCode == 403 {
			return nil, ErrForbidden
		} else if resp.StatusCode == 404 {
			return nil, ErrNotFound
		} else {
			bodyBytes, _ := io.ReadAll(resp.Body)
			var errorResponse ErrorResponse
			// Try to decode the error response from the bodyBytes
			if err := json.Unmarshal(bodyBytes, &errorResponse); err == nil && errorResponse.StatusCode != "" {
				return nil, fmt.Errorf("%s", errorResponse.Message)
			}
			return nil, fmt.Errorf("unexpected response: %s: status code: %d", summarizeBody(bodyBytes), resp.StatusCode)
		}
	}

	return resp, nil
}
