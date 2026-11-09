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
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// ResolveError is a refusal or failure the broker answered with. Its envelope is
// the Resources API's: {"error": text, "code": string}.
type ResolveError struct {
	Status  int
	Code    string
	Message string
}

func newResolveError(status int, body []byte) *ResolveError {
	refused := &ResolveError{Status: status}
	var envelope struct {
		Error string `json:"error"`
		Code  string `json:"code"`
	}
	if json.Unmarshal(body, &envelope) == nil {
		refused.Code, refused.Message = envelope.Code, envelope.Error
	}
	return refused
}

func (e *ResolveError) Error() string {
	switch {
	case e.Message != "" && e.Code != "":
		return fmt.Sprintf("the broker answered %d %s: %s", e.Status, e.Code, e.Message)
	case e.Message != "":
		return fmt.Sprintf("the broker answered %d: %s", e.Status, e.Message)
	default:
		return fmt.Sprintf("the broker answered %d", e.Status)
	}
}

// Terminal reports whether retrying is pointless: a 401 means the login can no
// longer be used, a 403 that the broker refused on policy (permission, a
// production environment, or the session's maximum age).
func (e *ResolveError) Terminal() bool {
	return e.Status == http.StatusUnauthorized || e.Status == http.StatusForbidden
}

// Reason turns a terminal refusal into the developer's next step.
func (e *ResolveError) Reason() string {
	switch e.Status {
	case http.StatusUnauthorized:
		return "your login is no longer valid; run `ws login`, then restart `ws remote connect`"
	case http.StatusForbidden:
		message := strings.TrimSpace(e.Message)
		if message == "" {
			message = "the broker refused to renew it"
		}
		return message + "; restart `ws remote connect` to start a new session"
	default:
		return e.Error()
	}
}
