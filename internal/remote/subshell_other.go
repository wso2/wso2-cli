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
// Adapted from openchoreo internal/occ/cmd/remote/subshell_other.go @ 69f36d373d0e7f9ad3f9dd21cf0c4960be75cd3d (v1.4.0-m.2).

//go:build !unix

package remote

import (
	"os/exec"
)

// setSubshellProcessGroup is a no-op on platforms without POSIX process groups.
func setSubshellProcessGroup(_ *exec.Cmd) {}

// restoreTerminalForeground is a no-op on platforms without POSIX job control.
func restoreTerminalForeground() {}
