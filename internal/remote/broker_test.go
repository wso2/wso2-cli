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
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/wso2/wso2-cli/internal/remoteconnect"
)

func staticToken(token string) func(context.Context) (string, error) {
	return func(context.Context) (string, error) { return token, nil }
}

func TestBrokerLinksReadsTheDeployedReleaseOfAnEnvironment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/base/links/shop/orders" {
			t.Errorf("request = %s %s", r.Method, r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer tok" {
			t.Errorf("authorization = %q", got)
		}
		_, _ = io.WriteString(w, `{"environments":[
			{"name":"stage","deployed":true,"release":{"env":{"A":{"ref":"r","key":"k"}}}},
			{"name":"development","deployed":true,"release":{"env":{"DB":{"ref":"orders-db","key":"url"}},"files":{"/p":{"ref":"orders-db","key":"ca"}}}}]}`)
	}))
	defer server.Close()
	broker := &Broker{BaseURL: server.URL + "/base/", Token: staticToken("tok")}

	release, found, err := broker.DeployedRelease(context.Background(), "shop", "orders", "development")
	if err != nil || !found {
		t.Fatalf("found=%v err=%v", found, err)
	}
	if release.Env["DB"] != (LinkTarget{Ref: "orders-db", Key: "url"}) || release.Files["/p"].Key != "ca" {
		t.Fatalf("release = %+v", release)
	}
}

func TestBrokerLinksIsAbsentForANewComponentOrAnUndeployedEnvironment(t *testing.T) {
	status := http.StatusNotFound
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if status == http.StatusNotFound {
			http.Error(w, `{"error":"no such component","code":"not_found"}`, status)
			return
		}
		_, _ = io.WriteString(w, `{"environments":[{"name":"stage","deployed":true,"release":{}}]}`)
	}))
	defer server.Close()
	broker := &Broker{BaseURL: server.URL, Token: staticToken("tok")}

	if _, found, err := broker.DeployedRelease(context.Background(), "shop", "new", "development"); err != nil || found {
		t.Fatalf("404: found=%v err=%v, want not found and no error", found, err)
	}
	status = http.StatusOK
	if _, found, err := broker.DeployedRelease(context.Background(), "shop", "new", "development"); err != nil || found {
		t.Fatalf("no release for the env: found=%v err=%v, want not found and no error", found, err)
	}
}

func TestBrokerResolveSendsOneRequestShapedByTheContract(t *testing.T) {
	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/projects/shop/remote-connect/resolve" {
			t.Errorf("request = %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer tok" {
			t.Errorf("authorization = %q", r.Header.Get("Authorization"))
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		_, _ = io.WriteString(w, `{"capability":"cap-1","renewAfterSeconds":600}`)
	}))
	defer server.Close()
	broker := &Broker{BaseURL: server.URL, Token: staticToken("tok")}

	resp, err := broker.Resolve(context.Background(), remoteconnect.ResolveRequest{
		Namespace: "ignored-org", Project: "shop", Component: "orders", Environment: "development",
		Purpose: remoteconnect.PurposeRenew, SkipSecrets: true, PreviousCapability: "cap-0",
		Resources: []remoteconnect.ResourceDep{{Ref: "orders-db"}},
	})
	if err != nil || resp.Capability != "cap-1" || resp.RenewAfterSeconds != 600 {
		t.Fatalf("resp=%+v err=%v", resp, err)
	}
	// The org comes from the token and the project from the path; purpose is
	// the broker's to set. None of the three is in the body.
	for _, absent := range []string{"namespace", "project", "purpose"} {
		if _, present := body[absent]; present {
			t.Errorf("body carries %q", absent)
		}
	}
	if body["component"] != "orders" || body["environment"] != "development" || body["previousCapability"] != "cap-0" {
		t.Errorf("body = %v", body)
	}
}

func TestBrokerRefusalsAreTerminalOnlyWhenRetryingCannotHelp(t *testing.T) {
	cases := []struct {
		status   int
		body     string
		terminal bool
		want     []string
		absent   []string
	}{
		{401, `{"error":"bad audience","code":"unauthorized"}`, true, []string{"wso2 login", "wso2 remote connect"}, nil},
		// Only an ended session is cured by starting a new one.
		{403, `{"error":"the session reached its maximum age","code":"session_expired"}`, true,
			[]string{"maximum age", "restart `wso2 remote connect`"}, nil},
		{403, `{"error":"you may not connect to this project","code":"permission_denied"}`, true,
			[]string{"you may not connect to this project", "access"}, []string{"restart"}},
		{403, `{"error":"production","code":"production_environment"}`, true,
			[]string{"production", "another environment"}, []string{"restart"}},
		{403, `{"error":"not migrated","code":"org_needs_migration"}`, true,
			[]string{"not migrated", "migrated"}, []string{"restart"}},
		{400, `{"error":"environment is required","code":"invalid_request"}`, true,
			[]string{"environment is required", "workload file"}, []string{"restart"}},
		{502, `{"error":"an upstream service failed","code":"upstream_error"}`, false, []string{"upstream"}, nil},
	}
	for _, c := range cases {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(c.status)
			_, _ = io.WriteString(w, c.body)
		}))
		broker := &Broker{BaseURL: server.URL, Token: staticToken("tok")}
		_, err := broker.Resolve(context.Background(), remoteconnect.ResolveRequest{Project: "p", Component: "c"})
		server.Close()
		var refused *BrokerError
		if !errors.As(err, &refused) {
			t.Fatalf("%d: err = %v, want a BrokerError", c.status, err)
		}
		if refused.Terminal() != c.terminal {
			t.Errorf("%d %s: terminal = %v, want %v", c.status, refused.Code, refused.Terminal(), c.terminal)
		}
		for _, want := range c.want {
			if !strings.Contains(refused.Reason(), want) {
				t.Errorf("%d %s: reason %q lacks %q", c.status, refused.Code, refused.Reason(), want)
			}
		}
		for _, absent := range c.absent {
			if strings.Contains(refused.Reason(), absent) {
				t.Errorf("%d %s: reason %q must not say %q", c.status, refused.Code, refused.Reason(), absent)
			}
		}
	}
}
