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
// Adapted from openchoreo internal/occ/cmd/remote/subshell_unix.go @ 69f36d373d0e7f9ad3f9dd21cf0c4960be75cd3d (v1.4.0-m.2).

//go:build unix

package remote

import (
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

// setSubshellProcessGroup puts the subshell into a new process group and, when stdin is
// a terminal, makes it that terminal's FOREGROUND process group (Foreground implies
// Setpgid). Without this the subshell would be a background group and block on SIGTTIN
// the instant the interactive shell read the TTY — the user could type but the shell
// would never receive input.
func setSubshellProcessGroup(cmd *exec.Cmd) {
	attr := &syscall.SysProcAttr{Setpgid: true}
	if f, ok := cmd.Stdin.(*os.File); ok && term.IsTerminal(int(f.Fd())) {
		attr.Foreground = true
		attr.Ctty = int(f.Fd())
	}
	cmd.SysProcAttr = attr
}

// restoreTerminalForeground reclaims the controlling terminal's foreground process group
// for ws after the subshell (which held the foreground) exits. ws is now a background
// group, so the TIOCSPGRP ioctl would itself raise SIGTTOU and stop ws — we ignore
// SIGTTOU across the call. No-op when stdin is not a terminal.
func restoreTerminalForeground() {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return
	}
	signal.Ignore(syscall.SIGTTOU)
	defer signal.Reset(syscall.SIGTTOU)
	_ = unix.IoctlSetPointerInt(fd, unix.TIOCSPGRP, syscall.Getpgrp())
}
