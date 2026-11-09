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

package remote

import (
	"reflect"
	"testing"

	"github.com/wso2/wso2-cli/internal/remoteconnect"
)

func TestMergeLinksGroupsReleaseByRefAndFileWins(t *testing.T) {
	release := LinkTargets{
		Env: map[string]LinkTarget{
			"DB_PASSWORD": {Ref: "orders-db", Key: "password"},
			"DB_HOST":     {Ref: "orders-db", Key: "host"},
			"CACHE_URL":   {Ref: "cache", Key: "url"},
		},
		Files: map[string]LinkTarget{
			"/etc/db/ca.pem":  {Ref: "orders-db", Key: "ca"},
			"/etc/db/old.pem": {Ref: "orders-db", Key: "old"},
		},
	}
	local := []remoteconnect.ResourceDep{
		{
			Ref:         "orders-db",
			EnvBindings: map[string]string{"pw": "DB_PASSWORD"},
			// The local file maps a different output to the same mount path.
			FileBindings: map[string]string{"cert": "/etc/db/ca.pem"},
		},
		{Ref: "queue", EnvBindings: map[string]string{"url": "QUEUE_URL"}},
	}

	got := MergeLinks(release, local)

	want := []remoteconnect.ResourceDep{
		{Ref: "cache", EnvBindings: map[string]string{"url": "CACHE_URL"}},
		{
			Ref:         "orders-db",
			EnvBindings: map[string]string{"host": "DB_HOST", "pw": "DB_PASSWORD"},
			FileBindings: map[string]string{
				"old":  "/etc/db/old.pem",
				"cert": "/etc/db/ca.pem",
			},
		},
		{Ref: "queue", EnvBindings: map[string]string{"url": "QUEUE_URL"}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("merged\n got %+v\nwant %+v", got, want)
	}
}

func TestMergeLinksWithNoReleaseIsTheFile(t *testing.T) {
	local := []remoteconnect.ResourceDep{{Ref: "a", EnvBindings: map[string]string{"x": "X"}}}
	if got := MergeLinks(LinkTargets{}, local); !reflect.DeepEqual(got, local) {
		t.Fatalf("got %+v, want the file's resources unchanged", got)
	}
}

func TestLoadWorkloadReadsOwnerAndDependencies(t *testing.T) {
	doc := `
apiVersion: openchoreo.dev/v1alpha1
kind: Workload
metadata:
  name: orders
spec:
  owner:
    projectName: shop
    componentName: orders
  dependencies:
    endpoints:
      - component: inventory
        name: http
        visibility: project
        envBindings:
          address: INVENTORY_URL
    resources:
      - ref: orders-db
        envBindings:
          password: DB_PASSWORD
        fileBindings:
          ca: /etc/db/ca.pem
`
	wl, err := ParseWorkload([]byte(doc))
	if err != nil {
		t.Fatal(err)
	}
	if wl.Project != "shop" || wl.Component != "orders" {
		t.Fatalf("owner = %q/%q", wl.Project, wl.Component)
	}
	wantEndpoints := []remoteconnect.EndpointDep{{
		Component: "inventory", Name: "http", Visibility: "project",
		EnvBindings: remoteconnect.EndpointEnvBindings{Address: "INVENTORY_URL"},
	}}
	if !reflect.DeepEqual(wl.Endpoints, wantEndpoints) {
		t.Fatalf("endpoints = %+v", wl.Endpoints)
	}
	wantResources := []remoteconnect.ResourceDep{{
		Ref:          "orders-db",
		EnvBindings:  map[string]string{"password": "DB_PASSWORD"},
		FileBindings: map[string]string{"ca": "/etc/db/ca.pem"},
	}}
	if !reflect.DeepEqual(wl.Resources, wantResources) {
		t.Fatalf("resources = %+v", wl.Resources)
	}
}

func TestParseWorkloadRefusesAFileWithoutOneOwnedWorkload(t *testing.T) {
	cases := map[string]string{
		"no workload":   "kind: ConfigMap\nmetadata:\n  name: x\n",
		"no owner":      "kind: Workload\nspec: {}\n",
		"two workloads": "kind: Workload\nspec:\n  owner: {projectName: p, componentName: a}\n---\nkind: Workload\nspec:\n  owner: {projectName: p, componentName: b}\n",
	}
	for name, doc := range cases {
		if _, err := ParseWorkload([]byte(doc)); err == nil {
			t.Errorf("%s: want an error", name)
		}
	}
}
