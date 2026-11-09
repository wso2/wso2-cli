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

package app_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wso2/wso2-cli/internal/contexts"
	"github.com/wso2/wso2-cli/internal/exit"
)

const remoteTestWorkload = `kind: Workload
spec:
  owner:
    projectName: shop
    componentName: orders
  dependencies:
    resources:
      - ref: orders-db
        envBindings:
          host: DB_HOST
`

// remoteDocument is one context that records the remote product the way the
// documentation says to: an exchanged product at the cloud audience, with the
// two permissions the broker checks.
func remoteDocument(mutate func(*contexts.Product)) contexts.Document {
	product := contexts.Product{
		Endpoint: "https://broker.example/admin/cloud/v1", Audience: contexts.RemoteConnectAudience,
		Scopes: []string{"cloud:remote-connect:connect", "cloud:remote-connect:read-secrets"},
		Grant:  &contexts.Grant{Kind: contexts.GrantExchange},
	}
	if mutate != nil {
		mutate(&product)
	}
	document := whoamiSeededDocument()
	document.Contexts[0].Organization = ""
	document.Contexts[0].Products = map[string]contexts.Product{contexts.RemoteProduct: product}
	return document
}

func writeRemoteWorkload(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "workload.yaml")
	if err := os.WriteFile(path, []byte(remoteTestWorkload), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestRemoteConnectNeedsARemoteProductOnTheContext(t *testing.T) {
	shell, out, errOut := newLoginShell(t)
	installLogin(t, shell, whoamiSeededDocument()) // no products

	code := shell.Run([]string{"remote", "connect", writeRemoteWorkload(t), "--env", "development"})
	if code == exit.OK {
		t.Fatalf("exit %d with no remote product; stdout: %s", code, out)
	}
	if !strings.Contains(errOut.String(), "remote.product_not_configured") ||
		!strings.Contains(errOut.String(), "products.remote") {
		t.Errorf("the refusal does not name the missing product:\n%s", errOut)
	}
}

func TestRemoteConnectRefusesAProductThatIsNotTheCloudExchange(t *testing.T) {
	for name, mutate := range map[string]func(*contexts.Product){
		"another audience":  func(p *contexts.Product) { p.Audience = "https://elsewhere.example" },
		"no exchange grant": func(p *contexts.Product) { p.Grant = nil },
	} {
		shell, _, errOut := newLoginShell(t)
		installLogin(t, shell, remoteDocument(mutate))
		code := shell.Run([]string{"remote", "connect", writeRemoteWorkload(t), "--env", "development"})
		if code == exit.OK || !strings.Contains(errOut.String(), "remote.product_not_configured") {
			t.Errorf("%s: exit %d, stderr:\n%s", name, code, errOut)
		}
	}
}

func TestRemoteConnectAsksForTheLoginBeforeTouchingTheNetwork(t *testing.T) {
	shell, out, errOut := newLoginShell(t)
	installLogin(t, shell, remoteDocument(nil))

	code := shell.Run([]string{"remote", "connect", writeRemoteWorkload(t), "--env", "development"})
	if code == exit.OK {
		t.Fatalf("exit %d with no login session", code)
	}
	if !strings.Contains(errOut.String(), "login") {
		t.Errorf("the refusal does not point at the login:\n%s", errOut)
	}
	if out.Len() != 0 {
		t.Errorf("stdout = %q, want nothing before a session exists", out)
	}
}

func TestRemoteConnectTakesOneWorkloadAndTheEnvironment(t *testing.T) {
	shell, _, errOut := newLoginShell(t)
	installLogin(t, shell, remoteDocument(nil))
	for name, args := range map[string][]string{
		"no file":        {"remote", "connect", "--env", "development"},
		"two files":      {"remote", "connect", "a.yaml", "b.yaml", "--env", "development"},
		"no environment": {"remote", "connect", writeRemoteWorkload(t)},
		"both secret flags": {"remote", "connect", writeRemoteWorkload(t), "--env", "development",
			"--no-secrets", "--show-secrets"},
	} {
		errOut.Reset()
		if code := shell.Run(args); code == exit.OK {
			t.Errorf("%s: exit %d, want a usage refusal", name, code)
		}
	}
}
