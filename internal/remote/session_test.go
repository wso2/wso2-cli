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

// Copyright 2026 The OpenChoreo Authors
// SPDX-License-Identifier: Apache-2.0
//
// Adapted from openchoreo internal/occ/cmd/remote/connect_test.go @ 69f36d373d0e7f9ad3f9dd21cf0c4960be75cd3d (v1.4.0-m.2).

package remote

import (
	"bytes"
	"errors"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/wso2/wso2-cli/internal/remoteconnect"
)

// TestForwardReportsExpiredSession: once the capability expires the agent rejects every
// new stream. Swallowing that left the app with a bare connection reset and no clue, so
// the first failure per dependency must say what happened and how to recover. Reaching
// this state at all means renewal did not land, which is what the message says.
func TestForwardReportsExpiredSession(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ln.Close() })

	out := &syncBuf{}
	expired, ok := remoteconnect.CapabilityExpiry(mintExpiredCapability(t))
	if !ok {
		t.Fatal("expected an expiry on the minted capability")
	}
	unit := &fakeRenewable{expiry: expired}
	reporter := newStreamErrorReporter(out, unit)
	openErr := errors.New("stream rejected: not authorized")
	reported := make(chan struct{}, 2)
	report := func(k string, e error) { reporter.report(k, e); reported <- struct{}{} }
	go forward(ln, "ep/finance/ledger-svc/http", func() (net.Conn, error) { return nil, openErr }, report) //nolint:unparam // always-error open is the point

	// Two connections: the message must appear once, not per connection.
	for range 2 {
		c, derr := net.Dial("tcp", ln.Addr().String())
		if derr != nil {
			t.Fatal(derr)
		}
		<-reported
		_ = c.Close()
	}

	got := out.String()
	if !strings.Contains(got, "ep/finance/ledger-svc/http") {
		t.Errorf("output does not name the dependency: %q", got)
	}
	if !strings.Contains(got, "expired") || !strings.Contains(got, "being renewed now") {
		t.Errorf("expired session must be explained with a remedy, got: %q", got)
	}
	if n := strings.Count(got, "ep/finance/ledger-svc/http"); n != 1 {
		t.Errorf("reported %d times, want once per dependency", n)
	}
	// Every refused connection asks for a renewal, not only the one that printed.
	if n := unit.nudges.Load(); n != 2 {
		t.Errorf("nudged %d times, want one per refused connection", n)
	}
}

// Once renewal has given up the remedy is to restart rather than to wait.
func TestReportExpiredSessionAfterRenewalStopped(t *testing.T) {
	out := &syncBuf{}
	reporter := newStreamErrorReporter(out, &fakeRenewable{
		expiry:  time.Now().Add(-time.Minute),
		stopped: true,
	})
	reporter.report("res/local-dev-postgres/client", errors.New("stream rejected: not authorized"))

	got := out.String()
	if !strings.Contains(got, "no longer being renewed") || !strings.Contains(got, "ws remote connect") {
		t.Errorf("a stopped session must tell the developer to restart, got: %q", got)
	}
}

// TestForwardReportsNonExpiryFailure: a failure that is not expiry surfaces the
// underlying error rather than blaming the session.
func TestForwardReportsNonExpiryFailure(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ln.Close() })

	out := &syncBuf{}
	reporter := newStreamErrorReporter(out, &fakeRenewable{}) // no expiry available
	reported := make(chan struct{}, 1)
	report := func(k string, e error) { reporter.report(k, e); reported <- struct{}{} }
	go forward(ln, "ep/doclet/backend-api/http", func() (net.Conn, error) {
		return nil, errors.New("dial remote-agent: connection refused")
	}, report)

	c, derr := net.Dial("tcp", ln.Addr().String())
	if derr != nil {
		t.Fatal(derr)
	}
	<-reported
	_ = c.Close()

	got := out.String()
	if !strings.Contains(got, "connection refused") {
		t.Errorf("underlying error not surfaced: %q", got)
	}
	if strings.Contains(got, "the session expired") {
		t.Errorf("misreported a dial failure as expiry: %q", got)
	}
}

// A composed connection URL must be re-pointed at the tunnel, so an app that reads
// only the URL works from the developer's machine.
func TestRewriteAddrsComposedValues(t *testing.T) {
	swaps := []addrSwap{newAddrSwap("cache.ns.svc.cluster.local:6379", "55445")}
	env := map[string]string{
		"URLOUT_URL": "redis://:urlonlycase@cache.ns.svc.cluster.local:6379",
		"CONNSTR":    "cache.ns.svc.cluster.local:6379,password=pw",
		"BUCKET":     "assets",
	}
	var out bytes.Buffer
	got := rewriteAddrs(env, swaps, "tc-urlout", &out)

	if got["URLOUT_URL"] != "redis://:urlonlycase@127.0.0.1:55445" {
		t.Errorf("URL not re-pointed: %q", got["URLOUT_URL"])
	}
	// Not a URL at all -- substitution has to work on the raw pair, not via a parser.
	if got["CONNSTR"] != "127.0.0.1:55445,password=pw" {
		t.Errorf("connection string not re-pointed: %q", got["CONNSTR"])
	}
	if got["BUCKET"] != "assets" {
		t.Errorf("unrelated value modified: %q", got["BUCKET"])
	}
	if out.Len() != 0 {
		t.Errorf("unexpected warning: %q", out.String())
	}
}

// Only the full host:port pair is replaced. A value naming the host with a different
// port stays as resolved and is reported, rather than being silently rewritten.
func TestRewriteAddrsLeavesOtherPortsAlone(t *testing.T) {
	swaps := []addrSwap{newAddrSwap("pg.ns.svc.cluster.local:5432", "60001")}
	env := map[string]string{"ADMIN": "http://pg.ns.svc.cluster.local:8080/admin"}
	var out bytes.Buffer
	got := rewriteAddrs(env, swaps, "tc-pg", &out)

	if got["ADMIN"] != "http://pg.ns.svc.cluster.local:8080/admin" {
		t.Errorf("value on another port should be untouched: %q", got["ADMIN"])
	}
	if !strings.Contains(out.String(), "still points at pg.ns.svc.cluster.local") {
		t.Errorf("expected a warning about the unreachable address, got %q", out.String())
	}
}

// A bare host must never be rewritten -- doing so would break TLS verification for a
// value that is a server name rather than a dial target.
func TestRewriteAddrsNeverRewritesBareHost(t *testing.T) {
	swaps := []addrSwap{newAddrSwap("pg.ns.svc.cluster.local:5432", "60001")}
	env := map[string]string{"SSL_SERVERNAME": "pg.ns.svc.cluster.local"}
	var out bytes.Buffer
	got := rewriteAddrs(env, swaps, "tc-pg", &out)

	if got["SSL_SERVERNAME"] != "pg.ns.svc.cluster.local" {
		t.Errorf("bare host was rewritten: %q", got["SSL_SERVERNAME"])
	}
}

// A value holding the host and the port apart cannot be matched as a pair, so each is
// substituted on its own. This is the shape a resource type takes when its only output
// is a driver connection string: there are no discrete host and port outputs to name.
func TestRewriteAddrsSplitHostAndPort(t *testing.T) {
	swaps := []addrSwap{newAddrSwap("cache.ns.svc.cluster.local:6379", "64168")}
	env := map[string]string{
		"CONNSTR": "host=cache.ns.svc.cluster.local,port=6379,password=mys3cret",
	}
	var out bytes.Buffer
	got := rewriteAddrs(env, swaps, "tc-split", &out)

	if want := "host=127.0.0.1,port=64168,password=mys3cret"; got["CONNSTR"] != want {
		t.Errorf("CONNSTR = %q, want %q", got["CONNSTR"], want)
	}
	// Substituting a bare host is the weaker inference, so it is never silent.
	if !strings.Contains(out.String(), "CONNSTR had its host and port re-pointed separately") {
		t.Errorf("expected the split rewrite to be reported, got %q", out.String())
	}
}

// Endpoints of one resource share a host, so a split rewrite must be decided against
// the original value -- otherwise the first endpoint's rewrite hides the host from the
// second and its port is left pointing into the cluster.
func TestRewriteAddrsSplitAcrossTwoEndpoints(t *testing.T) {
	swaps := []addrSwap{
		newAddrSwap("nats.ns.svc.cluster.local:4222", "64168"),
		newAddrSwap("nats.ns.svc.cluster.local:8222", "64169"),
	}
	env := map[string]string{
		"CONNSTR": "host=nats.ns.svc.cluster.local,port=4222,monitor=8222",
	}
	var out bytes.Buffer
	got := rewriteAddrs(env, swaps, "tc-broker", &out)

	if want := "host=127.0.0.1,port=64168,monitor=64169"; got["CONNSTR"] != want {
		t.Errorf("CONNSTR = %q, want %q", got["CONNSTR"], want)
	}
}

// The split path needs the port as a standalone token. A digit run that merely contains
// it, or a hostname that ends in it, is not this endpoint's port.
func TestRewriteAddrsSplitRequiresDelimitedPort(t *testing.T) {
	swaps := []addrSwap{newAddrSwap("cache.ns.svc.cluster.local:6379", "64168")}
	env := map[string]string{
		"LONGER": "host=cache.ns.svc.cluster.local,port=63790",
		"SUFFIX": "host=cache.ns.svc.cluster.local,peer=cache6379.ns",
	}
	var out bytes.Buffer
	got := rewriteAddrs(env, swaps, "tc-split", &out)

	for k, v := range env {
		if got[k] != v {
			t.Errorf("%s should be untouched: %q", k, got[k])
		}
	}
	if !strings.Contains(out.String(), "still points at cache.ns.svc.cluster.local") {
		t.Errorf("expected the unreachable-address warning, got %q", out.String())
	}
}

// A value carrying the fused pair must not also take the split path: the host may
// appear a second time as a server name, which a bare-host rewrite would corrupt.
func TestRewriteAddrsFusedPairSuppressesSplit(t *testing.T) {
	swaps := []addrSwap{newAddrSwap("pg.ns.svc.cluster.local:5432", "60001")}
	env := map[string]string{
		"DSN": "postgres://pg.ns.svc.cluster.local:5432/db?sslmode=verify-full&host=pg.ns.svc.cluster.local",
	}
	var out bytes.Buffer
	got := rewriteAddrs(env, swaps, "tc-pg", &out)

	want := "postgres://127.0.0.1:60001/db?sslmode=verify-full&host=pg.ns.svc.cluster.local"
	if got["DSN"] != want {
		t.Errorf("DSN = %q, want %q", got["DSN"], want)
	}
}

// A resource output mounted as a file has no local equivalent -- the value lives on the
// A file binding whose value cannot be fetched is named with the reason, not dropped.
// A missing mount usually presents as a bug in the app, so silence is the worst answer.
func TestFetchBindingsReportsUnfetchableFile(t *testing.T) {
	resp := &remoteconnect.ResolveResponse{
		Resources: []remoteconnect.ResourceBindings{{
			Ref:          "doclet-postgres",
			FetchFile:    map[string]string{"/etc/tls/ca.crt": "sec/doclet-postgres/caCert"},
			FetchAgentID: "dp-a",
		}},
	}
	tn := &fakeTunnel{fetchErr: errors.New("read failed")}
	store := newFileStore()
	t.Cleanup(store.cleanup)

	var buf bytes.Buffer
	fetchBindings(map[string]string{}, map[string]bool{}, &buf, resp,
		map[string]tunnel{"dp-a": tn}, store, nil, false)

	out := buf.String()
	for _, want := range []string{"res/doclet-postgres", "/etc/tls/ca.crt", "read failed"} {
		if !strings.Contains(out, want) {
			t.Errorf("report missing %q; got:\n%s", want, out)
		}
	}
}

// --no-secrets fetches nothing at all: not one stream is opened, and every binding is
// reported as skipped. The point of the flag is that no credential enters the process,
// so a fetch that still happened would defeat it entirely.
func TestFetchBindingsNoSecretsFetchesNothing(t *testing.T) {
	resp := &remoteconnect.ResolveResponse{
		Resources: []remoteconnect.ResourceBindings{{
			Ref:          "doclet-postgres",
			FetchEnv:     map[string]string{"DB_PASSWORD": "sec/doclet-postgres/password"},
			FetchFile:    map[string]string{"/etc/tls/ca.crt": "sec/doclet-postgres/caCert"},
			FetchAgentID: "dp-a",
		}},
	}
	tn := &fakeTunnel{values: map[string][]byte{
		"sec/doclet-postgres/password": []byte("s3cr3t"),
		"sec/doclet-postgres/caCert":   []byte("PEM"),
	}}
	store := newFileStore()
	t.Cleanup(store.cleanup)

	overrides := map[string]string{}
	var buf bytes.Buffer
	fetchBindings(overrides, map[string]bool{}, &buf, resp,
		map[string]tunnel{"dp-a": tn}, store, nil, true)

	if len(tn.fetched) != 0 {
		t.Fatalf("--no-secrets still fetched %v", tn.fetched)
	}
	if _, set := overrides["DB_PASSWORD"]; set {
		t.Fatalf("--no-secrets still set DB_PASSWORD")
	}
	if !strings.Contains(buf.String(), "--no-secrets") {
		t.Errorf("skipped bindings not attributed to the flag:\n%s", buf.String())
	}
}

// A fetched value reaches the environment but never the terminal. occ's output is the
// developer's scrollback and often a pasted bug report, so printing the value there
// would undo the design's whole point of keeping it off the control plane.
func TestFetchBindingsNeverPrintsTheValue(t *testing.T) {
	const secret = "pa55w0rd-do-not-print"
	resp := &remoteconnect.ResolveResponse{
		Resources: []remoteconnect.ResourceBindings{{
			Ref:          "doclet-postgres",
			FetchEnv:     map[string]string{"DB_PASSWORD": "sec/doclet-postgres/password"},
			FetchAgentID: "dp-a",
		}},
	}
	tn := &fakeTunnel{values: map[string][]byte{"sec/doclet-postgres/password": []byte(secret)}}
	store := newFileStore()
	t.Cleanup(store.cleanup)

	overrides := map[string]string{}
	sensitive := map[string]bool{}
	var buf bytes.Buffer
	fetchBindings(overrides, sensitive, &buf, resp, map[string]tunnel{"dp-a": tn}, store, nil, false)

	if overrides["DB_PASSWORD"] != secret {
		t.Fatalf("value did not reach the env: %q", overrides["DB_PASSWORD"])
	}
	if !sensitive["DB_PASSWORD"] {
		t.Errorf("fetched value not marked sensitive; --print-env would leak it")
	}
	if strings.Contains(buf.String(), secret) {
		t.Errorf("value printed to output:\n%s", buf.String())
	}

	// And --print-env redacts it.
	var envBuf bytes.Buffer
	printEnvBindings(&envBuf, overrides, sensitive, false)
	if strings.Contains(envBuf.String(), secret) {
		t.Errorf("--print-env leaked the value:\n%s", envBuf.String())
	}
	if !strings.Contains(envBuf.String(), "DB_PASSWORD") {
		t.Errorf("--print-env hid the binding entirely:\n%s", envBuf.String())
	}
}

// A split rewrite must not consume a port an earlier endpoint's rewrite just produced.
// Ephemeral local ports are OS-assigned and can coincide with another endpoint's
// in-cluster port; substituting into that would silently undo the first rewrite.
func TestRewriteAddrsSkipsPortThatIsAnotherEndpointsLocalPort(t *testing.T) {
	swaps := []addrSwap{
		newAddrSwap("pg.ns.svc:5432", "40001"),
		// This endpoint's in-cluster port is the first endpoint's local port.
		newAddrSwap("pg.ns.svc:40001", "50002"),
	}
	env := map[string]string{"DSN": "host=pg.ns.svc,port=5432,fallbackPort=40001"}
	var buf bytes.Buffer
	got := rewriteAddrs(env, swaps, "pg", &buf)

	// The first endpoint's rewrite stands: its local port is not re-substituted.
	if !strings.Contains(got["DSN"], "port=40001") {
		t.Errorf("first rewrite was clobbered: %q", got["DSN"])
	}
	if !strings.Contains(got["DSN"], localHost) {
		t.Errorf("host was not re-pointed: %q", got["DSN"])
	}
	if strings.Contains(got["DSN"], "port=50002") {
		t.Errorf("second endpoint consumed the first's local port: %q", got["DSN"])
	}
}

// applyResourceBindings reports a resource nothing was tunneled for, naming what the
// values actually are — a binding pinned to a ResourceRelease cut before its type
// declared endpoints exports in-cluster addresses this machine cannot resolve, and
// saying nothing there reads as success.
func TestApplyResourceBindingsReportsUntunneledResource(t *testing.T) {
	resp := &remoteconnect.ResolveResponse{
		Resources: []remoteconnect.ResourceBindings{{
			Ref:              "doclet-postgres",
			StaticEnv:        map[string]string{"DB_HOST": "pg.ns.svc.cluster.local", "DB_PORT": "5432"},
			OmittedSecretEnv: []remoteconnect.OmittedBinding{{Target: "DB_PASSWORD", Reason: "secret-backed"}},
		}},
	}
	overrides := map[string]string{}
	var buf bytes.Buffer
	applyResourceBindings(overrides, &buf, resp, nil)

	if overrides["DB_HOST"] != "pg.ns.svc.cluster.local" {
		t.Errorf("static bindings should pass through unchanged, got %+v", overrides)
	}
	out := buf.String()
	for _, want := range []string{"res/doclet-postgres", "no address tunneled", "as published in-cluster", "DB_PASSWORD"} {
		if !strings.Contains(out, want) {
			t.Errorf("report missing %q; got:\n%s", want, out)
		}
	}
}

// A resource that publishes its connection URL through a Secret must have the tunneled
// address substituted in the FETCHED value, not just in the values the control plane
// resolved. Before value resolution this binding was omitted with a warning; resolving it
// without re-pointing would hand the app the in-cluster address instead — reachable from
// the cluster, not from the developer's machine, and silent about it.
func TestFetchBindingsRepointsTunneledAddressInFetchedValue(t *testing.T) {
	const remote = "pg.dp-ns.svc.cluster.local:5432"
	resp := &remoteconnect.ResolveResponse{
		Resources: []remoteconnect.ResourceBindings{{
			Ref:          "local-dev-postgres",
			FetchEnv:     map[string]string{"DB_URL": "sec/local-dev-postgres/url"},
			FetchAgentID: "dp-a",
		}},
	}
	tn := &fakeTunnel{values: map[string][]byte{
		"sec/local-dev-postgres/url": []byte("postgres://app:pw@" + remote + "/doclet"),
	}}
	store := newFileStore()
	t.Cleanup(store.cleanup)

	localAddrs := map[string][]addrSwap{
		"local-dev-postgres": {newAddrSwap(remote, "40001")},
	}
	overrides := map[string]string{}
	var buf bytes.Buffer
	fetchBindings(overrides, map[string]bool{}, &buf, resp,
		map[string]tunnel{"dp-a": tn}, store, localAddrs, false)

	got := overrides["DB_URL"]
	if strings.Contains(got, remote) {
		t.Errorf("fetched URL still points into the cluster: %q", got)
	}
	if !strings.Contains(got, localHost+":40001") {
		t.Errorf("fetched URL was not re-pointed at the tunnel: %q", got)
	}
	// The credential inside the URL must survive the rewrite untouched.
	if !strings.Contains(got, "app:pw@") {
		t.Errorf("rewrite mangled the value: %q", got)
	}
	if strings.Contains(buf.String(), "pw@") {
		t.Errorf("value printed to output:\n%s", buf.String())
	}
}

// Two mount paths that flatten to the same name get distinct local files.
func TestFileStoreKeepsFlatteningCollisionsDistinct(t *testing.T) {
	store := newFileStore()
	t.Cleanup(store.cleanup)

	// Both flatten to "etc_tls_certs_ca.pem".
	first, err := store.write("/etc/tls_certs/ca.pem", []byte("FIRST"))
	if err != nil {
		t.Fatalf("write first: %v", err)
	}
	second, err := store.write("/etc/tls/certs/ca.pem", []byte("SECOND"))
	if err != nil {
		t.Fatalf("write second: %v", err)
	}

	if first == second {
		t.Fatalf("both mount paths resolved to one local file: %s", first)
	}
	for path, want := range map[string]string{first: "FIRST", second: "SECOND"} {
		got, rerr := os.ReadFile(path)
		if rerr != nil {
			t.Fatalf("read %s: %v", path, rerr)
		}
		if string(got) != want {
			t.Fatalf("%s holds %q, want %q", path, got, want)
		}
	}
}

// go test's stdin is not a terminal, which is the same condition a pipe presents. The
// prompt must answer itself rather than block, and must not ask where no one can see it.
func TestConfirmShowSecretsDeclinesWithoutATerminal(t *testing.T) {
	var buf bytes.Buffer
	if confirmShowSecrets(&buf, []string{"DB_PASSWORD"}) {
		t.Error("confirmShowSecrets said yes with no terminal on stdin")
	}
	if buf.Len() != 0 {
		t.Errorf("prompted with no terminal to read it:\n%s", buf.String())
	}
}
