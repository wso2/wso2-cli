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

package remote

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/wso2/wso2-cli/internal/remoteconnect"
)

const ordersWorkload = `apiVersion: openchoreo.dev/v1alpha1
kind: Workload
metadata:
  name: orders
spec:
  owner:
    projectName: shop
    componentName: orders
  dependencies:
    resources:
      - ref: orders-db
        envBindings:
          host: DB_HOST
          port: DB_PORT
          password: DB_PASSWORD_LOCAL
`

func writeWorkload(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "workload.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

// fakeBroker serves the two operations the CLI calls, records what it was sent,
// and counts the exchanges behind the tokens it saw.
type fakeBroker struct {
	t         *testing.T
	linksBody string // empty answers 404
	resolve   remoteconnect.ResolveResponse

	mu        sync.Mutex
	resolves  []map[string]any
	bearers   []string
	linkCalls int
}

func (f *fakeBroker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.bearers = append(f.bearers, r.Header.Get("Authorization"))
	switch {
	case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/links/"):
		f.linkCalls++
		if f.linksBody == "" {
			w.WriteHeader(http.StatusNotFound)
			_, _ = io.WriteString(w, `{"error":"not found","code":"not_found"}`)
			return
		}
		_, _ = io.WriteString(w, f.linksBody)
	case r.Method == http.MethodPost && r.URL.Path == "/projects/shop/remote-connect/resolve":
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		f.resolves = append(f.resolves, body)
		_ = json.NewEncoder(w).Encode(f.resolve)
	default:
		f.t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
	}
}

// countingTokens mints a distinct token per exchange.
func countingTokens() (*Tokens, *atomic.Int32) {
	var mints atomic.Int32
	return &Tokens{Mint: func(context.Context) (Grant, error) {
		n := mints.Add(1)
		return Grant{Token: "token-" + string(rune('0'+n)), ExpiresAt: time.Now().Add(time.Hour)}, nil
	}}, &mints
}

const ordersLinks = `{"environments":[
  {"name":"development","deployed":true,"release":{
    "env":{"DB_PASSWORD":{"ref":"orders-db","key":"password"},"DB_USER":{"ref":"orders-db","key":"user"}}}}]}`

func ordersResponse() remoteconnect.ResolveResponse {
	const agent = "dp-orders-development"
	return remoteconnect.ResolveResponse{
		Capability: "cap-1",
		Targets: []remoteconnect.ResolvedTarget{{
			Key: remoteconnect.ResourceTargetKey("orders-db", "client"), Proto: "tcp", AgentID: agent,
			Resource: &remoteconnect.ResourceRender{Ref: "orders-db", Address: "client", HostEnv: "DB_HOST", PortEnv: "DB_PORT"},
		}},
		Resources: []remoteconnect.ResourceBindings{{
			Ref: "orders-db", StaticEnv: map[string]string{"DB_NAME": "orders"},
			FetchEnv: map[string]string{"DB_PASSWORD": "sec/orders-db/password"}, FetchAgentID: agent,
		}},
		Agents: map[string]remoteconnect.AgentEndpoint{agent: {Endpoint: "wss://bridge.invalid/connect", ServerName: agent + ".remote-connect"}},
	}
}

// connectFixture wires a Remote to a fake broker over HTTP, the way ws wires it
// to the real one, with a tunnel that reaches a local echo server.
func connectFixture(t *testing.T, broker *fakeBroker) (*Remote, *atomic.Int32, *fakeTunnel) {
	t.Helper()
	server := httptest.NewServer(broker)
	t.Cleanup(server.Close)
	tokens, mints := countingTokens()
	api := &Broker{BaseURL: server.URL, Token: tokens.Token}
	echo := startEchoServer(t)
	agentTunnel := &fakeTunnel{addr: echo.Addr().String(), values: map[string][]byte{"sec/orders-db/password": []byte("s3cret")}}
	remote := New(api, tokens, "wso2-cli/test")
	remote.DialAgent = func(ctx context.Context, _ remoteconnect.AgentEndpoint, _ func() string) (tunnel, error) {
		// The dial presents the same token the broker was called with.
		if _, err := tokens.Token(ctx); err != nil {
			return nil, err
		}
		return agentTunnel, nil
	}
	return remote, mints, agentTunnel
}

func TestConnectMergesTheDeployedReleaseOverTheFileAndTunnels(t *testing.T) {
	broker := &fakeBroker{t: t, linksBody: ordersLinks, resolve: ordersResponse()}
	remote, mints, _ := connectFixture(t, broker)

	var env map[string]string
	roundTrip := make(chan error, 1)
	remote.RunShell = func(_ context.Context, shellEnv []string) error {
		env = envToMap(shellEnv)
		roundTrip <- tunnelRoundTrip(net.JoinHostPort(env["DB_HOST"], env["DB_PORT"]), "ping")
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := remote.Connect(ctx, ConnectParams{WorkloadPath: writeWorkload(t, ordersWorkload), Environment: "development"}, io.Discard)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	if err := <-roundTrip; err != nil {
		t.Fatalf("round trip through the tunnel: %v", err)
	}

	if len(broker.resolves) != 1 {
		t.Fatalf("%d resolves, want one", len(broker.resolves))
	}
	// DB_USER comes from the deployed release. The release also binds the
	// password output to DB_PASSWORD, but the file binds it to
	// DB_PASSWORD_LOCAL, and the file wins.
	resources, _ := broker.resolves[0]["resources"].([]any)
	if len(resources) != 1 {
		t.Fatalf("resources = %v", broker.resolves[0]["resources"])
	}
	bindings := resources[0].(map[string]any)["envBindings"].(map[string]any)
	want := map[string]string{"host": "DB_HOST", "port": "DB_PORT", "password": "DB_PASSWORD_LOCAL", "user": "DB_USER"}
	for output, envVar := range want {
		if bindings[output] != envVar {
			t.Errorf("envBindings[%s] = %v, want %s (all: %v)", output, bindings[output], envVar, bindings)
		}
	}
	if broker.resolves[0]["component"] != "orders" || broker.resolves[0]["environment"] != "development" {
		t.Errorf("body = %v", broker.resolves[0])
	}
	if env["DB_NAME"] != "orders" || env["DB_HOST"] != "127.0.0.1" {
		t.Errorf("env DB_NAME=%q DB_HOST=%q", env["DB_NAME"], env["DB_HOST"])
	}
	if env["DB_PASSWORD"] != "s3cret" {
		t.Errorf("the fetched secret did not reach the environment")
	}
	// One exchange serves the links read, the resolve and the tunnel dial.
	if got := mints.Load(); got != 1 {
		t.Errorf("%d exchanges at connect, want 1", got)
	}
	for _, bearer := range broker.bearers {
		if bearer != "Bearer token-1" {
			t.Errorf("a broker call carried %q", bearer)
		}
	}
}

func TestConnectWithoutDeployedLinksUsesTheFileAndSaysSo(t *testing.T) {
	broker := &fakeBroker{t: t, resolve: ordersResponse()} // links answer 404
	remote, _, _ := connectFixture(t, broker)
	remote.RunShell = func(context.Context, []string) error { return nil }
	var out strings.Builder

	err := remote.Connect(context.Background(), ConnectParams{WorkloadPath: writeWorkload(t, ordersWorkload), Environment: "development"}, &out)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	if !strings.Contains(out.String(), "using the dependencies in") {
		t.Errorf("no note that the file is the whole list:\n%s", out.String())
	}
	resources := broker.resolves[0]["resources"].([]any)
	bindings := resources[0].(map[string]any)["envBindings"].(map[string]any)
	if _, fromRelease := bindings["user"]; fromRelease || len(bindings) != 3 {
		t.Errorf("envBindings = %v, want the file's three only", bindings)
	}
}

func TestConnectNoSecretsAsksForNoGrantsAndFetchesNothing(t *testing.T) {
	broker := &fakeBroker{t: t, linksBody: ordersLinks, resolve: ordersResponse()}
	remote, _, tunnel := connectFixture(t, broker)
	var env map[string]string
	remote.RunShell = func(_ context.Context, shellEnv []string) error { env = envToMap(shellEnv); return nil }

	err := remote.Connect(context.Background(),
		ConnectParams{WorkloadPath: writeWorkload(t, ordersWorkload), Environment: "development", NoSecrets: true}, io.Discard)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	if broker.resolves[0]["skipSecrets"] != true {
		t.Errorf("skipSecrets = %v, want true", broker.resolves[0]["skipSecrets"])
	}
	if len(tunnel.fetched) != 0 {
		t.Errorf("fetched %v under --no-secrets", tunnel.fetched)
	}
	if _, present := env["DB_PASSWORD"]; present {
		t.Error("a secret reached the environment under --no-secrets")
	}
}

func TestConnectPrintEnvRedactsFetchedValuesUnlessShown(t *testing.T) {
	for _, show := range []bool{false, true} {
		broker := &fakeBroker{t: t, linksBody: ordersLinks, resolve: ordersResponse()}
		remote, _, _ := connectFixture(t, broker)
		remote.RunShell = func(context.Context, []string) error {
			t.Fatal("--print-env must not start a subshell")
			return nil
		}
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		var out strings.Builder
		err := remote.Connect(ctx, ConnectParams{
			WorkloadPath: writeWorkload(t, ordersWorkload), Environment: "development", PrintEnv: true, ShowSecrets: show,
		}, &out)
		cancel()
		if err != nil {
			t.Fatalf("Connect: %v", err)
		}
		printed := strings.Contains(out.String(), "export DB_PASSWORD=s3cret")
		if printed != show {
			t.Errorf("show=%v but the secret printed=%v:\n%s", show, printed, out.String())
		}
	}
}

func TestConnectNeedsAnEnvironmentAndAWorkloadWithDependencies(t *testing.T) {
	remote, _, _ := connectFixture(t, &fakeBroker{t: t, resolve: ordersResponse()})
	if err := remote.Connect(context.Background(), ConnectParams{WorkloadPath: writeWorkload(t, ordersWorkload)}, io.Discard); err == nil {
		t.Error("no --env: want an error")
	}
	bare := "kind: Workload\nspec:\n  owner: {projectName: shop, componentName: orders}\n"
	err := remote.Connect(context.Background(), ConnectParams{WorkloadPath: writeWorkload(t, bare), Environment: "development"}, io.Discard)
	if err == nil {
		t.Error("a workload with no dependencies and no deployed release: want an error")
	}
}

// Each renewal runs on its own exchange, and the renewed call carries that
// token; a refused renewal ends the session at once.
func TestRenewalExchangesAfreshAndEndsOnARefusal(t *testing.T) {
	tokens, mints := countingTokens()
	if _, err := tokens.Token(context.Background()); err != nil { // the connect's exchange
		t.Fatal(err)
	}
	out := &syncBuf{}
	u, _ := unitFixture(t, out)
	u.renewAfter = time.Millisecond

	var mu sync.Mutex
	var seen []string
	inner := &scriptedResolver{steps: []resolveStep{
		{resp: &remoteconnect.ResolveResponse{
			Capability: "cap-2", Targets: []remoteconnect.ResolvedTarget{{Key: "res/postgres/primary", Proto: "tcp", AgentID: "dp-a"}},
			Agents:            map[string]remoteconnect.AgentEndpoint{"dp-a": {ServerName: "dp-a.remote-connect"}},
			RenewAfterSeconds: 1,
		}},
		{err: &BrokerError{Status: http.StatusForbidden, Code: "session_expired", Message: "the session reached its maximum age"}},
	}}
	recording := resolverFunc(func(ctx context.Context, req remoteconnect.ResolveRequest) (*remoteconnect.ResolveResponse, error) {
		token, err := tokens.Token(ctx)
		mu.Lock()
		seen = append(seen, token)
		mu.Unlock()
		if err != nil {
			return nil, err
		}
		return inner.Resolve(ctx, req)
	})

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	done := make(chan struct{})
	go func() { u.renew(ctx, renewalResolver{inner: recording, tokens: tokens}); close(done) }()
	select {
	case <-done:
	case <-ctx.Done():
		t.Fatal("the renewal did not end on the refusal")
	}

	if want := []string{"token-2", "token-3"}; len(seen) != 2 || seen[0] != want[0] || seen[1] != want[1] {
		t.Errorf("renewals carried %v, want a fresh exchange each: %v", seen, want)
	}
	if got := mints.Load(); got != 3 {
		t.Errorf("%d exchanges, want 1 at connect and one per renewal", got)
	}
	if u.cap() != "cap-2" {
		t.Errorf("capability = %q, want the first renewal's", u.cap())
	}
	if got := out.String(); !strings.Contains(got, "maximum age") || !strings.Contains(got, "restart `wso2 remote connect`") {
		t.Errorf("the refusal's remedy is missing: %q", got)
	}
}

type resolverFunc func(context.Context, remoteconnect.ResolveRequest) (*remoteconnect.ResolveResponse, error)

func (f resolverFunc) Resolve(ctx context.Context, req remoteconnect.ResolveRequest) (*remoteconnect.ResolveResponse, error) {
	return f(ctx, req)
}
