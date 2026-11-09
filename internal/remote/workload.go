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
	"bytes"
	"errors"
	"fmt"
	"io"

	"go.yaml.in/yaml/v3"

	"github.com/wso2/wso2-cli/internal/remoteconnect"
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
	Kind string `yaml:"kind"`
	Spec struct {
		Owner struct {
			ProjectName   string `yaml:"projectName"`
			ComponentName string `yaml:"componentName"`
		} `yaml:"owner"`
		Dependencies struct {
			Endpoints []endpointDependency `yaml:"endpoints"`
			Resources []resourceDependency `yaml:"resources"`
		} `yaml:"dependencies"`
	} `yaml:"spec"`
}

type endpointDependency struct {
	Project     string `yaml:"project"`
	Component   string `yaml:"component"`
	ReleaseLine string `yaml:"releaseLine"`
	Name        string `yaml:"name"`
	Visibility  string `yaml:"visibility"`
	EnvBindings struct {
		Address  string `yaml:"address"`
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		BasePath string `yaml:"basePath"`
	} `yaml:"envBindings"`
}

type resourceDependency struct {
	Ref          string            `yaml:"ref"`
	EnvBindings  map[string]string `yaml:"envBindings"`
	FileBindings map[string]string `yaml:"fileBindings"`
}

// ParseWorkload reads the one Workload a file holds. Documents of any other
// kind are skipped. A file with no Workload, more than one, or one that names
// no owner is refused: a session is for one component.
func ParseWorkload(data []byte) (Workload, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	var found []workloadDocument
	for {
		var doc workloadDocument
		err := decoder.Decode(&doc)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return Workload{}, fmt.Errorf("the file is not valid YAML: %w", err)
		}
		if doc.Kind == "Workload" {
			found = append(found, doc)
		}
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
