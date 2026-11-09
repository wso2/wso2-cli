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
// Adapted from openchoreo internal/occ/cmd/remote/connect.go @ 69f36d373d0e7f9ad3f9dd21cf0c4960be75cd3d (v1.4.0-m.2).

package remote

import (
	"fmt"
	"io"

	"github.com/wso2/wso2-cli/internal/remoteconnect"
)

// applyResourceBindings merges the env each resource dependency contributes directly:
// values the control plane resolved itself, re-pointed at the local listeners where they
// embed a tunneled address, plus a report of the bindings that could not be resolved at
// all. Values behind a Secret/ConfigMap reference are not here — they are fetched over
// the tunnel by fetchBindings. Resources with no address at all are included, which is
// what lets a dependency with nothing to dial still supply its configuration.
func applyResourceBindings(
	overrides map[string]string,
	out io.Writer,
	resp *remoteconnect.ResolveResponse,
	localAddrs map[string][]addrSwap,
) {
	tunneledRefs := make(map[string]bool, len(resp.Targets))
	for _, t := range resp.Targets {
		if t.Resource != nil {
			tunneledRefs[t.Resource.Ref] = true
		}
	}
	for _, rb := range resp.Resources {
		mergeOverrides(overrides, out, rewriteAddrs(rb.StaticEnv, localAddrs[rb.Ref],
			remoteconnect.ResourceRefKey(rb.Ref), out))
		if !tunneledRefs[rb.Ref] {
			// Either the type declares nothing dialable, or the binding is still pinned
			// to a ResourceRelease cut before it did. Either way the addresses below are
			// the in-cluster ones as published, which this machine may not resolve.
			fmt.Fprintf(out, "  %-28s (no address tunneled; %d binding(s) resolved as published in-cluster)\n",
				remoteconnect.ResourceRefKey(rb.Ref), len(rb.StaticEnv))
		}
		for _, om := range rb.OmittedSecretEnv {
			what := "value"
			if om.File {
				what = "file"
			}
			fmt.Fprintf(out, "  ! %s: %s %s not resolved (%s)\n",
				remoteconnect.ResourceRefKey(rb.Ref), om.Target, what, om.Reason)
		}
	}
}
