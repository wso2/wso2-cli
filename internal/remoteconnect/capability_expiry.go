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
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"
)

// CapabilityExpiry reads a capability's expiry without verifying its signature. The
// client has no verification key, but it still has to tell the developer when the
// session ends, otherwise expiry looks like the dependency going away. Never use this
// to make a security decision; only the control plane's verified check governs access.
//
// Adapted from openchoreo internal/remoteconnect/capability.go @
// 69f36d373d0e7f9ad3f9dd21cf0c4960be75cd3d (v1.4.0-m.2): reads the payload with the
// standard library instead of golang-jwt.
func CapabilityExpiry(token string) (time.Time, bool) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return time.Time{}, false
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return time.Time{}, false
	}
	var claims struct {
		Exp *json.Number `json:"exp"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil || claims.Exp == nil {
		return time.Time{}, false
	}
	secs, err := claims.Exp.Float64()
	if err != nil {
		return time.Time{}, false
	}
	return time.Unix(0, int64(secs*float64(time.Second))), true
}
