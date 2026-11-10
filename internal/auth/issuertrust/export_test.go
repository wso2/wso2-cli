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

package issuertrust

import "io"

// CaptureDevPlaintextWarnings sends the dev-only plaintext warnings to w and
// forgets which hosts were already warned about, returning a restore func.
func CaptureDevPlaintextWarnings(w io.Writer) func() {
	previous := devPlaintextWarnings
	devPlaintextWarnings = w
	warnedHosts.Range(func(key, _ any) bool { warnedHosts.Delete(key); return true })
	return func() { devPlaintextWarnings = previous }
}
