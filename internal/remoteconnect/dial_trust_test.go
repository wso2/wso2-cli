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

package remoteconnect

import (
	"net/http"
	"testing"
	"time"
)

// installDefaultTransport stands in for the CLI's WSO2_CA_FILE trust, which replaces
// http.DefaultTransport (internal/app/trust.go).
func installDefaultTransport(t *testing.T, rt http.RoundTripper) {
	t.Helper()
	prev := http.DefaultTransport
	http.DefaultTransport = rt
	t.Cleanup(func() { http.DefaultTransport = prev })
}

func TestDialTunnelHonoursTheCLITrustWhenNoHTTPClientIsGiven(t *testing.T) {
	b := newBridge(t) // served with a certificate from a private CA
	installDefaultTransport(t, b.Client().Transport)
	ctx, cancel := contextWithTestTimeout(t)
	defer cancel()

	opts := b.options("tok")
	opts.HTTPClient = nil
	c, err := DialTunnel(ctx, b.endpoint(), b.identity.caPEM, testServerName, func() string { return "cap" }, opts)
	if err != nil {
		t.Fatalf("DialTunnel with the private CA trusted: %v", err)
	}
	t.Cleanup(func() { _ = c.Close() })
	assertEchoes(t, c, "ep/p/c/http")
}

func TestDialTunnelRefusesAnUntrustedBridgeCertificate(t *testing.T) {
	b := newBridge(t)
	installDefaultTransport(t, http.DefaultTransport.(*http.Transport).Clone())
	ctx, cancel := contextWithTestTimeout(t)
	defer cancel()

	opts := b.options("tok")
	opts.HTTPClient = nil
	opts.RetryInitial, opts.RetryMax = time.Millisecond, 5*time.Millisecond
	if c, err := DialTunnel(ctx, b.endpoint(), b.identity.caPEM, testServerName, func() string { return "cap" }, opts); err == nil {
		_ = c.Close()
		t.Fatal("DialTunnel succeeded against a bridge whose CA is not trusted")
	}
}
