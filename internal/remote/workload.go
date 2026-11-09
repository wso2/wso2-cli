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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/wso2/wso2-cli/internal/remoteconnect"
	"github.com/wso2/wso2-cli/internal/yamldoc"
)

// Workload is what remote connect reads of a workload file: the owning
// component and the declared dependencies. The file is an OpenChoreo Workload;
// every other field of it is ignored.
type Workload struct {
	Project   string
	Component string
	Endpoints []remoteconnect.EndpointDep
	Resources []remoteconnect.ResourceDep
}

// workloadDocument is the decode shape of one Workload document.
type workloadDocument struct {
	Kind string `json:"kind"`
	Spec struct {
		Owner struct {
			ProjectName   string `json:"projectName"`
			ComponentName string `json:"componentName"`
		} `json:"owner"`
		Dependencies struct {
			Endpoints []endpointDependency `json:"endpoints"`
			Resources []resourceDependency `json:"resources"`
		} `json:"dependencies"`
	} `json:"spec"`
}

type endpointDependency struct {
	Project     string `json:"project"`
	Component   string `json:"component"`
	ReleaseLine string `json:"releaseLine"`
	Name        string `json:"name"`
	Visibility  string `json:"visibility"`
	EnvBindings struct {
		Address  string `json:"address"`
		Host     string `json:"host"`
		Port     string `json:"port"`
		BasePath string `json:"basePath"`
	} `json:"envBindings"`
}

type resourceDependency struct {
	Ref          string            `json:"ref"`
	EnvBindings  map[string]string `json:"envBindings"`
	FileBindings map[string]string `json:"fileBindings"`
}

// ParseWorkload reads the one Workload a file holds. Documents of any other
// kind are skipped. A file with no Workload, more than one, or one that names
// no owner is refused: a session is for one component.
func ParseWorkload(data []byte) (Workload, error) {
	var found []workloadDocument
	// A leading separator opens the first document; the prepended newline lets
	// one split rule cover it.
	for _, part := range bytes.Split(append([]byte("\n"), data...), []byte("\n---")) {
		if len(bytes.TrimSpace(part)) == 0 {
			continue
		}
		converted, err := yamldoc.ToJSON(part)
		if err != nil {
			// A document that does not parse is reported only when it means to
			// be a Workload, which keeps templated YAML beside it from failing
			// the read.
			if bytes.Contains(part, []byte("kind: Workload")) {
				return Workload{}, fmt.Errorf("a Workload document is not readable YAML: %w", err)
			}
			continue
		}
		// Decided on the kind alone, so another kind's shape under spec can
		// never fail the read.
		var probe struct {
			Kind string `json:"kind"`
		}
		if json.Unmarshal(converted, &probe) != nil || probe.Kind != "Workload" {
			continue
		}
		var doc workloadDocument
		if err := json.Unmarshal(converted, &doc); err != nil {
			return Workload{}, fmt.Errorf("a Workload document is malformed: %w", err)
		}
		found = append(found, doc)
	}
	switch len(found) {
	case 0:
		return Workload{}, errors.New("the file holds no Workload document")
	case 1:
	default:
		return Workload{}, fmt.Errorf("the file holds %d Workload documents; remote connect takes exactly one", len(found))
	}
	doc := found[0]
	owner := doc.Spec.Owner
	if owner.ProjectName == "" || owner.ComponentName == "" {
		return Workload{}, errors.New("the Workload has no spec.owner.projectName or spec.owner.componentName")
	}
	workload := Workload{Project: owner.ProjectName, Component: owner.ComponentName}
	for _, e := range doc.Spec.Dependencies.Endpoints {
		workload.Endpoints = append(workload.Endpoints, remoteconnect.EndpointDep{
			Project: e.Project, Component: e.Component, ReleaseLine: e.ReleaseLine,
			Name: e.Name, Visibility: e.Visibility,
			EnvBindings: remoteconnect.EndpointEnvBindings{
				Address: e.EnvBindings.Address, Host: e.EnvBindings.Host,
				Port: e.EnvBindings.Port, BasePath: e.EnvBindings.BasePath,
			},
		})
	}
	for _, r := range doc.Spec.Dependencies.Resources {
		workload.Resources = append(workload.Resources, remoteconnect.ResourceDep{
			Ref: r.Ref, EnvBindings: r.EnvBindings, FileBindings: r.FileBindings,
		})
	}
	return workload, nil
}
