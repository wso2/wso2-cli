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
	"maps"
	"slices"

	"github.com/wso2/wso2-cli/internal/remoteconnect"
)

// LinkTarget is one Resource output a link binds: the Resource ref and the
// output key.
type LinkTarget struct {
	Ref string `json:"ref"`
	Key string `json:"key"`
}

// LinkTargets is what a component's links bind: env var name -> output, and
// file path -> output. It is the shape of an environment's deployed release in
// the links read (environments[].release).
type LinkTargets struct {
	Env   map[string]LinkTarget `json:"env,omitempty"`
	Files map[string]LinkTarget `json:"files,omitempty"`
}

// MergeLinks is the client side of the dependency merge (the broker does none).
// It groups the environment's deployed release by Resource ref, inverting each
// target into the output -> env var (or file path) shape a resolve takes, then
// lays the workload file over it. The file wins: a release target whose env
// var or file path the file also binds is dropped, so the same destination is
// never claimed twice. The result is sorted by ref.
func MergeLinks(release LinkTargets, local []remoteconnect.ResourceDep) []remoteconnect.ResourceDep {
	boundEnv, boundFiles := map[string]bool{}, map[string]bool{}
	for _, dep := range local {
		for _, envVar := range dep.EnvBindings {
			boundEnv[envVar] = true
		}
		for _, path := range dep.FileBindings {
			boundFiles[path] = true
		}
	}

	byRef := map[string]*remoteconnect.ResourceDep{}
	entry := func(ref string) *remoteconnect.ResourceDep {
		if byRef[ref] == nil {
			byRef[ref] = &remoteconnect.ResourceDep{Ref: ref}
		}
		return byRef[ref]
	}
	for _, envVar := range slices.Sorted(maps.Keys(release.Env)) {
		if boundEnv[envVar] {
			continue
		}
		target := release.Env[envVar]
		dep := entry(target.Ref)
		if dep.EnvBindings == nil {
			dep.EnvBindings = map[string]string{}
		}
		dep.EnvBindings[target.Key] = envVar
	}
	for _, path := range slices.Sorted(maps.Keys(release.Files)) {
		if boundFiles[path] {
			continue
		}
		target := release.Files[path]
		dep := entry(target.Ref)
		if dep.FileBindings == nil {
			dep.FileBindings = map[string]string{}
		}
		dep.FileBindings[target.Key] = path
	}
	for _, dep := range local {
		merged := entry(dep.Ref)
		if len(dep.EnvBindings) > 0 && merged.EnvBindings == nil {
			merged.EnvBindings = map[string]string{}
		}
		maps.Copy(merged.EnvBindings, dep.EnvBindings)
		if len(dep.FileBindings) > 0 && merged.FileBindings == nil {
			merged.FileBindings = map[string]string{}
		}
		maps.Copy(merged.FileBindings, dep.FileBindings)
	}

	merged := make([]remoteconnect.ResourceDep, 0, len(byRef))
	for _, ref := range slices.Sorted(maps.Keys(byRef)) {
		merged = append(merged, *byRef[ref])
	}
	return merged
}
