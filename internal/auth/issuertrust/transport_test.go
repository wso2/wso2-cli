// Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
//
// WSO2 LLC. licenses this file to you under the Apache License,
// Version 2.0 (the "License"); you may not use this file except
// in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package issuertrust_test

import (
	"bytes"
	"encoding/json"
	"net/url"
	"strings"
	"testing"

	"github.com/wso2/wso2-cli/internal/auth/issuertrust"
)

func TestPlaintextRefusesAnyDiscoveredEndpointOffLoopbackHTTP(t *testing.T) {
	secure := `{"issuer": "https://id.example.test", "token_endpoint": "https://id.example.test/token",
		"jwks_uri": "https://id.example.test/jwks"}`
	for document, want := range map[string]bool{
		secure: false,
		`{"issuer": "http://localhost:9443", "token_endpoint": "http://127.0.0.1:9443/token"}`:    false,
		`{"issuer": "https://id.example.test", "token_endpoint": "http://id.example.test/token"}`: true,
		`{"issuer": "https://id.example.test", "jwks_uri": "http://id.example.test/jwks"}`:        true,
		`{"issuer": "https://id.example.test", "device_authorization_endpoint": "http://d.test"}`: true,
		`{"issuer": "https://id.example.test", "revocation_endpoint": "http://r.test/revoke"}`:    true,
		`{"issuer": "https://id.example.test", "end_session_endpoint": "http://e.test/logout"}`:   true,
		`{"issuer": "http://id.example.test"}`:                                                    true,
		`{"issuer": "https://id.example.test", "token_endpoint": 7}`:                              true,
	} {
		claims := func(into any) error { return json.Unmarshal([]byte(document), into) }
		if got := issuertrust.Plaintext(claims); got != want {
			t.Errorf("Plaintext(%s) = %v, want %v", document, got, want)
		}
	}
}

// For a declared issuer identifier the issuer member alone is exempt: it is
// compared with the identifier instead, and every endpoint keeps the rule.
func TestPlaintextEndpointsExemptsOnlyTheIssuerMember(t *testing.T) {
	for document, want := range map[string]bool{
		`{"issuer": "example-idp", "token_endpoint": "https://id.example.test/token"}`: false,
		`{"issuer": "example-idp", "token_endpoint": "http://id.example.test/token"}`:  true,
		`{"issuer": "example-idp", "jwks_uri": "http://id.example.test/jwks"}`:         true,
	} {
		claims := func(into any) error { return json.Unmarshal([]byte(document), into) }
		if got := issuertrust.PlaintextEndpoints(claims); got != want {
			t.Errorf("PlaintextEndpoints(%s) = %v, want %v", document, got, want)
		}
	}
	claims := func(into any) error { return json.Unmarshal([]byte(`{"issuer": "example-idp"}`), into) }
	if !issuertrust.Plaintext(claims) {
		t.Error("Plaintext passed an issuer member that is no URL")
	}
}

func TestSecureAllowsPlainHTTPOnlyOnLoopback(t *testing.T) {
	for raw, want := range map[string]bool{
		"https://login.example.test/oauth2/token": true,
		"HTTPS://login.example.test":              true,
		"http://login.example.test/oauth2/token":  false,
		"http://10.0.0.5:9443":                    false,
		"http://192.168.1.10":                     false,
		"http://localhost.example.test":           false,
		"http://127.0.0.1.example.test":           false,
		"http://localhost:9443/oauth2/token":      true,
		"http://LocalHost:9443":                   true,
		"http://127.0.0.1:8090":                   true,
		"http://127.255.255.254":                  true,
		"http://[::1]:9443":                       true,
		"ftp://login.example.test":                false,
		"login.example.test":                      false,
	} {
		parsed, err := url.Parse(raw)
		if err != nil {
			t.Fatalf("parse %q: %v", raw, err)
		}
		if got := issuertrust.Secure(parsed); got != want {
			t.Errorf("Secure(%q) = %v, want %v", raw, got, want)
		}
	}
}

func secureFor(t *testing.T, raw string) bool {
	t.Helper()
	parsed, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("parse %q: %v", raw, err)
	}
	return issuertrust.Secure(parsed)
}

func TestSecureRefusesPlainHTTPWhenNoDevPlaintextHostIsListed(t *testing.T) {
	t.Setenv(issuertrust.DevPlaintextHostsEnv, "")
	var warnings bytes.Buffer
	defer issuertrust.CaptureDevPlaintextWarnings(&warnings)()

	if secureFor(t, "http://idp.local:8090/oauth2/token") {
		t.Error("Secure allowed plain http off loopback with the override unset")
	}
	if warnings.Len() != 0 {
		t.Errorf("unexpected warning: %q", warnings.String())
	}
}

func TestSecureAllowsAListedDevPlaintextHostAndWarnsOnce(t *testing.T) {
	t.Setenv(issuertrust.DevPlaintextHostsEnv, " idp.local:8090 , Other.Local ")
	var warnings bytes.Buffer
	defer issuertrust.CaptureDevPlaintextWarnings(&warnings)()

	for i := 0; i < 3; i++ {
		if !secureFor(t, "http://idp.local:8090/.well-known/openid-configuration") {
			t.Fatal("Secure refused a host listed in the override")
		}
	}
	if !secureFor(t, "http://other.local/token") {
		t.Error("Secure refused a listed host with no port")
	}
	got := warnings.String()
	if strings.Count(got, "idp.local:8090") != 1 || strings.Count(got, "other.local") != 1 {
		t.Errorf("want one warning per host, got %q", got)
	}
	if !strings.Contains(got, issuertrust.DevPlaintextHostsEnv) {
		t.Errorf("warning does not name the variable: %q", got)
	}
}

func TestSecureStillRefusesUnlistedHostsWithTheOverrideSet(t *testing.T) {
	t.Setenv(issuertrust.DevPlaintextHostsEnv, "idp.local:8090")
	var warnings bytes.Buffer
	defer issuertrust.CaptureDevPlaintextWarnings(&warnings)()

	for _, raw := range []string{
		"http://idp.local:9999/token", // other port
		"http://idp.local/token",      // no port
		"http://evil.idp.local:8090",  // no wildcard or suffix match
		"http://other.test:8090",
		"ftp://idp.local:8090",
	} {
		if secureFor(t, raw) {
			t.Errorf("Secure allowed %q, which is not listed", raw)
		}
	}
	if warnings.Len() != 0 {
		t.Errorf("unexpected warning: %q", warnings.String())
	}
}

func TestSecureLoopbackIsUnchangedByTheOverrideAndNeverWarns(t *testing.T) {
	t.Setenv(issuertrust.DevPlaintextHostsEnv, "idp.local:8090")
	var warnings bytes.Buffer
	defer issuertrust.CaptureDevPlaintextWarnings(&warnings)()

	if !secureFor(t, "http://localhost:9443") || !secureFor(t, "http://127.0.0.1:8090") {
		t.Error("loopback http was refused")
	}
	if !secureFor(t, "https://login.example.test") {
		t.Error("https was refused")
	}
	if warnings.Len() != 0 {
		t.Errorf("unexpected warning: %q", warnings.String())
	}
}

func TestPlaintextAcceptsListedDevHostsInDiscoveryDocuments(t *testing.T) {
	t.Setenv(issuertrust.DevPlaintextHostsEnv, "idp.local:8090")
	defer issuertrust.CaptureDevPlaintextWarnings(&bytes.Buffer{})()
	doc := `{"issuer": "http://idp.local:8090", "token_endpoint": "http://idp.local:8090/token"}`
	claims := func(into any) error { return json.Unmarshal([]byte(doc), into) }
	if issuertrust.Plaintext(claims) {
		t.Error("Plaintext refused a document whose hosts are all listed")
	}
	other := `{"issuer": "http://idp.local:8090", "jwks_uri": "http://unlisted.local/jwks"}`
	claims = func(into any) error { return json.Unmarshal([]byte(other), into) }
	if !issuertrust.Plaintext(claims) {
		t.Error("Plaintext accepted an unlisted endpoint host")
	}
}
