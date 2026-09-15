package auth

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListenForAuthCode(t *testing.T) {
	port, err := getRandomPort()
	require.NoError(t, err, "getRandomPort failed")

	timeout := 60 * time.Second
	authCodeChan := make(chan string)
	authErrChan := make(chan error)

	go func() {
		fmt.Println("Starting server...")
		authCode, err := ListenForAuthCode(port, timeout)
		if err != nil {
			fmt.Println("Server error:", err)
			authErrChan <- err
		} else {
			fmt.Println("Server received auth code:", authCode)
			authCodeChan <- authCode
		}
	}()

	go func() {
		fmt.Println("Waiting for server to be ready...")
		time.Sleep(5 * time.Second) // wait for server to be ready

		client := &http.Client{}
		req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:%d/auth-callback?code=1234", port), nil)
		require.NoError(t, err, "NewRequest failed")

		fmt.Println("Sending request...")
		resp, err := client.Do(req) // FIXME, ignore error for now
		require.NoError(t, err, "Do failed")
		// check resp.StatusCode
		assert.Equal(t, 200, resp.StatusCode, "Do returned unexpected status code")
		// check resp.Body
		assert.Equal(t, "text/html; charset=utf-8", resp.Header.Get("Content-Type"), "Do returned unexpected Content-Type")
		bodyBytes, err := io.ReadAll(resp.Body)
		require.NoError(t, err, "ReadAll failed")
		bodyString := string(bodyBytes)
		assert.Equal(t, oauthSuccessPage, bodyString, "Do returned unexpected body")
	}()

	select {
	case authCode := <-authCodeChan:
		require.Equal(t, "1234", authCode, "listenForAuthCode returned unexpected authCode")
	case err := <-authErrChan:
		t.Errorf("listenForAuthCode failed with error: %v", err)
	}
}

// The auth callback page used to render its wordmark as SVG vector outlines,
// which spelled "WSO2 Developer Platform" — the wrong product. Because the
// letters were <path> geometry rather than text, no grep could find it and the
// rebrand missed it entirely. The wordmark is now real text, so it is greppable
// and this test can hold it.
func TestOauthSuccessPageBranding(t *testing.T) {
	if !strings.Contains(oauthSuccessPage, "Integration Platform") {
		t.Error("success page must name the Integration Platform")
	}
	if strings.Contains(oauthSuccessPage, "Developer Platform") {
		t.Error("success page must not name the Developer Platform")
	}

	// Guard against the wordmark regressing to vector outlines. The circular
	// WSO2 pulse mark is legitimately two paths; a wordmark would add ~20 more.
	if got := strings.Count(oauthSuccessPage, "<path"); got > 2 {
		t.Errorf("success page has %d <path> elements; a text wordmark needs at most 2 (the icon). "+
			"Vector-outline text cannot be grep-checked and is how the wrong brand name survived the rebrand", got)
	}
}
