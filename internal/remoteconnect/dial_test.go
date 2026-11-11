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
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

const testServerName = "agent.remote-connect.test"

// agentIdentity is the self-signed certificate a remote-agent presents and the client pins.
type agentIdentity struct {
	cert  tls.Certificate
	caPEM string
}

func newAgentIdentity(t *testing.T) agentIdentity {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: testServerName},
		DNSNames:     []string{testServerName},
		NotBefore:    time.Now().Add(-time.Minute),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	return agentIdentity{
		cert:  tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key},
		caPEM: string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})),
	}
}

// echoAgent serves the agent side of the tunnel on an established, TLS-wrapped conn.
func echoAgent(t *testing.T, tlsConn net.Conn, seen chan<- StreamOpen) {
	t.Helper()
	fakeAgent(t, tlsConn, func(stream net.Conn) {
		defer func() { _ = stream.Close() }()
		var so StreamOpen
		if err := ReadMessage(stream, &so); err != nil {
			return
		}
		if seen != nil {
			seen <- so
		}
		_ = WriteMessage(stream, StreamResult{OK: true})
		_, _ = io.Copy(stream, stream)
	})
}

func assertEchoes(t *testing.T, c *TunnelClient, key string) {
	t.Helper()
	stream, err := c.OpenStream(key)
	if err != nil {
		t.Fatalf("open stream: %v", err)
	}
	defer func() { _ = stream.Close() }()
	_, _ = stream.Write([]byte("hi"))
	buf := make([]byte, 2)
	if _, err := io.ReadFull(stream, buf); err != nil || string(buf) != "hi" {
		t.Fatalf("echo = %q, %v", buf, err)
	}
}

// bridge is an in-process stand-in for the remote-connect bridge: a TLS HTTP server
// that upgrades /v1/tunnel to a WebSocket and runs the agent behind it.
type bridge struct {
	*httptest.Server
	identity agentIdentity
	mu       sync.Mutex
	headers  []http.Header
	attempts atomic.Int32
	// refuse, when set, answers an attempt with this status instead of upgrading.
	refuse func(attempt int32) int
	seen   chan StreamOpen
}

func newBridge(t *testing.T) *bridge { return newBridgeALPN(t, false) }

// newBridgeALPN is newBridge whose TLS listener, when http2 is set, also offers h2 in
// ALPN, as an Envoy or kgateway HTTPS listener does.
func newBridgeALPN(t *testing.T, http2 bool) *bridge {
	t.Helper()
	b := &bridge{identity: newAgentIdentity(t), seen: make(chan StreamOpen, 4)}
	b.Server = httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt := b.attempts.Add(1)
		b.mu.Lock()
		b.headers = append(b.headers, r.Header.Clone())
		b.mu.Unlock()
		if b.refuse != nil {
			if status := b.refuse(attempt); status != 0 {
				w.WriteHeader(status)
				return
			}
		}
		ws, err := (&websocket.Upgrader{}).Upgrade(w, r, nil)
		if err != nil {
			return
		}
		conn := newWebSocketConn(ws)
		tlsConn := tls.Server(conn, &tls.Config{Certificates: []tls.Certificate{b.identity.cert}, MinVersion: tls.VersionTLS12})
		echoAgent(t, tlsConn, b.seen)
		<-r.Context().Done()
	}))
	b.EnableHTTP2 = http2
	b.StartTLS()
	t.Cleanup(b.Close)
	return b
}

func (b *bridge) endpoint() string { return "wss" + strings.TrimPrefix(b.URL, "https") + "/v1/tunnel" }

func (b *bridge) options(token string) DialOptions {
	return DialOptions{
		BearerToken:  token,
		UserAgent:    "wso2-cli/test",
		HTTPClient:   b.Client(),
		RetryInitial: 5 * time.Millisecond,
		RetryMax:     20 * time.Millisecond,
	}
}

func TestDialTunnelOverWebSocketForWssEndpoint(t *testing.T) {
	b := newBridge(t)
	ctx, cancel := contextWithTestTimeout(t)
	defer cancel()

	c, err := DialTunnel(ctx, b.endpoint(), b.identity.caPEM, testServerName, func() string { return "cap-ws" }, b.options("tok-123"))
	if err != nil {
		t.Fatalf("DialTunnel: %v", err)
	}
	t.Cleanup(func() { _ = c.Close() })

	assertEchoes(t, c, "ep/p/c/http")
	if got := (<-b.seen); got.Capability != "cap-ws" {
		t.Errorf("capability = %q", got.Capability)
	}
	h := b.headers[0]
	if h.Get("Authorization") != "Bearer tok-123" {
		t.Errorf("Authorization header not sent as a Bearer token")
	}
	if h.Get("User-Agent") != "wso2-cli/test" {
		t.Errorf("User-Agent = %q", h.Get("User-Agent"))
	}
}

func TestDialTunnelUsesPlainTLSForOtherEndpoints(t *testing.T) {
	id := newAgentIdentity(t)
	ln, err := tls.Listen("tcp", "127.0.0.1:0", &tls.Config{Certificates: []tls.Certificate{id.cert}, MinVersion: tls.VersionTLS12})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ln.Close() })
	go func() {
		conn, aerr := ln.Accept()
		if aerr != nil {
			return
		}
		echoAgent(t, conn, nil)
	}()
	ctx, cancel := contextWithTestTimeout(t)
	defer cancel()

	// A bearer token is configured but must not be used: there is no HTTP request.
	c, err := DialTunnel(ctx, ln.Addr().String(), id.caPEM, testServerName, func() string { return "cap" }, DialOptions{BearerToken: "tok"})
	if err != nil {
		t.Fatalf("DialTunnel: %v", err)
	}
	t.Cleanup(func() { _ = c.Close() })
	assertEchoes(t, c, "ep/p/c/http")
}

func TestDialTunnelRetriesTransientWebSocketFailures(t *testing.T) {
	b := newBridge(t)
	b.refuse = func(attempt int32) int {
		if attempt < 3 {
			return http.StatusServiceUnavailable
		}
		return 0
	}
	ctx, cancel := contextWithTestTimeout(t)
	defer cancel()

	c, err := DialTunnel(ctx, b.endpoint(), b.identity.caPEM, testServerName, func() string { return "cap" }, b.options("tok"))
	if err != nil {
		t.Fatalf("DialTunnel: %v", err)
	}
	t.Cleanup(func() { _ = c.Close() })
	if got := b.attempts.Load(); got != 3 {
		t.Errorf("attempts = %d, want 3", got)
	}
}

func TestDialTunnelDoesNotRetryARefusedToken(t *testing.T) {
	b := newBridge(t)
	b.refuse = func(int32) int { return http.StatusUnauthorized }
	ctx, cancel := contextWithTestTimeout(t)
	defer cancel()

	_, err := DialTunnel(ctx, b.endpoint(), b.identity.caPEM, testServerName, func() string { return "cap" }, b.options("tok"))
	if err == nil {
		t.Fatal("expected a refusal")
	}
	if got := b.attempts.Load(); got != 1 {
		t.Errorf("attempts = %d, want 1 (a 401 cannot succeed on retry)", got)
	}
	if !strings.Contains(err.Error(), "401") {
		t.Errorf("error %q does not carry the status", err)
	}
}

func TestDialTunnelStillPinsTheAgentCertificateOverWebSocket(t *testing.T) {
	b := newBridge(t)
	other := newAgentIdentity(t)
	ctx, cancel := contextWithTestTimeout(t)
	defer cancel()

	_, err := DialTunnel(ctx, b.endpoint(), other.caPEM, testServerName, func() string { return "cap" }, b.options("tok"))
	if err == nil {
		t.Fatal("expected the pinned-certificate check to fail")
	}
}

func contextWithTestTimeout(t *testing.T) (context.Context, context.CancelFunc) {
	t.Helper()
	return context.WithTimeout(context.Background(), 10*time.Second)
}
