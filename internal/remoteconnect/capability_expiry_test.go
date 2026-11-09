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
	"fmt"
	"testing"
	"time"
)

// jwtWithPayload builds an unsigned compact JWT; CapabilityExpiry never checks the signature.
func jwtWithPayload(payload string) string {
	enc := base64.RawURLEncoding.EncodeToString
	return enc([]byte(`{"alg":"EdDSA"}`)) + "." + enc([]byte(payload)) + ".sig"
}

func TestCapabilityExpiryReadsExp(t *testing.T) {
	got, ok := CapabilityExpiry(jwtWithPayload(`{"exp":1893456000}`))
	if !ok {
		t.Fatal("expected an expiry")
	}
	if want := time.Unix(1893456000, 0); !got.Equal(want) {
		t.Fatalf("expiry = %v, want %v", got, want)
	}
}

func TestCapabilityExpiryRefusesWhatIsNotAnExpiringJWT(t *testing.T) {
	for name, token := range map[string]string{
		"no exp":      jwtWithPayload(`{"sub":"x"}`),
		"not a jwt":   "garbage",
		"bad base64":  "a.!!!.c",
		"bad json":    jwtWithPayload(`{`),
		"empty":       "",
		"exp as text": jwtWithPayload(fmt.Sprintf(`{"exp":%q}`, "soon")),
	} {
		t.Run(name, func(t *testing.T) {
			if _, ok := CapabilityExpiry(token); ok {
				t.Fatal("expected no expiry")
			}
		})
	}
}
