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

// BrokerError is a refusal or failure the broker answered with. Its envelope is
// the Resources API's: {"error": text, "code": string}.
type BrokerError struct {
	Status  int
	Code    string
	Message string
}

func newBrokerError(status int, body []byte) *BrokerError {
	refused := &BrokerError{Status: status}
	var envelope struct {
		Error string `json:"error"`
		Code  string `json:"code"`
	}
	if json.Unmarshal(body, &envelope) == nil {
		refused.Code, refused.Message = envelope.Code, envelope.Error
	}
	return refused
}

func (e *BrokerError) Error() string {
	switch {
	case e.Message != "" && e.Code != "":
		return fmt.Sprintf("the broker answered %d %s: %s", e.Status, e.Code, e.Message)
	case e.Message != "":
		return fmt.Sprintf("the broker answered %d: %s", e.Status, e.Message)
	default:
		return fmt.Sprintf("the broker answered %d", e.Status)
	}
}

// Terminal reports whether retrying is pointless: the login can no longer be
// used (401), the broker refused on policy (403), or the request is itself
// wrong (400). A 5xx or a transport failure may pass, so those are retried.
func (e *BrokerError) Terminal() bool {
	switch e.Status {
	case http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden:
		return true
	}
	return false
}

// Reason turns a terminal refusal into the developer's next step. Only an ended
// session is cured by starting a new one; every other refusal names what has to
// change first.
func (e *BrokerError) Reason() string {
	message := strings.TrimSpace(e.Message)
	switch {
	case e.Status == http.StatusUnauthorized:
		return "your login is no longer valid; run `" + loginCommand + "`, then restart `" + connectCommand + "`"
	case e.Status == http.StatusBadRequest:
		return withMessage(message, "the broker rejected the request") +
			"; correct the workload file or --env and run `" + connectCommand + "` again"
	case e.Status == http.StatusForbidden && e.Code == "session_expired":
		return withMessage(message, "the session reached its maximum age") +
			"; restart `" + connectCommand + "` to start a new session"
	case e.Status == http.StatusForbidden && e.Code == "permission_denied":
		return withMessage(message, "you do not have access to this") +
			"; ask an administrator of the organization for the access it names"
	case e.Status == http.StatusForbidden && e.Code == "production_environment":
		return withMessage(message, "remote connect is not available for a production environment") +
			"; connect to another environment"
	case e.Status == http.StatusForbidden && e.Code == "org_needs_migration":
		return withMessage(message, "the organization has not been migrated") +
			"; the organization has to be migrated before remote connect can be used"
	case e.Status == http.StatusForbidden:
		return withMessage(message, "the broker refused the request")
	default:
		return e.Error()
	}
}

func withMessage(message, fallback string) string {
	if message == "" {
		return fallback
	}
	return message
}

// LoginRequired is a token exchange that failed because the login session can
// no longer be exchanged. Like a terminal broker refusal, no retry mends it.
type LoginRequired struct{ Cause error }

func (e LoginRequired) Error() string { return e.Cause.Error() }
func (e LoginRequired) Unwrap() error { return e.Cause }

// Terminal is always true: only a new login helps.
func (LoginRequired) Terminal() bool { return true }

// Reason names the login as the next step.
func (LoginRequired) Reason() string {
	return "your login has expired or was refused; run `" + loginCommand + "`, then restart `" + connectCommand + "`"
}

// Terminated is a failure no retry can mend, with the developer's next step.
type Terminated interface {
	error
	Terminal() bool
	Reason() string
}

// The command names the user's next steps spell, in the form every suggestion
// in this binary uses.
const (
	loginCommand   = "wso2 login"
	connectCommand = "wso2 remote connect"
)
