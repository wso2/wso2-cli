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

// This package had no tests of its own until the atomic write moved out to
// internal/atomicfile. Its behaviour was covered end to end by test/acceptance,
// which installs a module and checks that it runs — a run that succeeds whether
// the store's documents land at 0644 or at the 0600 os.CreateTemp opens a
// temporary file at. So the one property the extraction could quietly break was
// the one nothing asserted.
//
// These tests are in-package because writeAtomically is the seam that changed;
// they pin what a caller of it is entitled to assume, and nothing about how the
// store is assembled.
package install

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/wso2/wso2-cli/internal/catalog"
	"github.com/wso2/wso2-cli/internal/modules"
	"github.com/wso2/wso2-cli/sdk/problem"
)

func TestWriteAtomicallyLeavesStoreDocumentsReadable(t *testing.T) {
	// The store's documents are 0644 and the receipts beside them are read by
	// tooling that is not the installing user. os.CreateTemp opens at 0600, so
	// a mode that is not carried through to the target makes every receipt
	// private without failing anything.
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("file modes are not meaningful here")
	}
	target := filepath.Join(t.TempDir(), "receipt.json")
	if err := writeAtomically("recording the receipt", target, []byte("{}\n")); err != nil {
		t.Fatalf("writeAtomically returned %v", err)
	}

	info, err := os.Stat(target)
	if err != nil {
		t.Fatalf("Stat returned %v", err)
	}
	if info.Mode().Perm() != 0o644 {
		t.Errorf("mode = %v, want 0644", info.Mode().Perm())
	}
}

func TestWriteAtomicallyLeavesNoTemporaryFileBehind(t *testing.T) {
	// A temporary file left in the store is a file the store's own readers
	// would have to learn to skip.
	directory := t.TempDir()
	if err := writeAtomically("writing the version policy",
		filepath.Join(directory, "policy.json"), []byte("{}\n")); err != nil {
		t.Fatalf("writeAtomically returned %v", err)
	}

	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatalf("ReadDir returned %v", err)
	}
	if len(entries) != 1 {
		var names []string
		for _, entry := range entries {
			names = append(names, entry.Name())
		}
		t.Errorf("the store holds %v, want only the document", names)
	}
}

func TestAnUnwrappedCauseIsStillReportedAsItself(t *testing.T) {
	// storeFailure formats the cause with %v, so a nil would reach a user as
	// "<nil>". atomicfile wraps both of its return paths today and this cannot
	// fire; it is pinned so that an edit over there degrades the message rather
	// than replacing it with nonsense.
	bare := errors.New("something the write helper did not wrap")
	if got := writeCause(bare); got != bare {
		t.Errorf("writeCause(%v) = %v, want the error itself", bare, got)
	}
}

func TestAFailedWriteNamesWhatFailedAndNotHowItIsImplemented(t *testing.T) {
	// action is in the message so the store's documents are told apart when one
	// cannot be written. The helper package that performs the rename is not:
	// a user cannot act on it, and it repeats the path the store already named.
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("directory permissions do not refuse a write here")
	}
	directory := t.TempDir()
	if err := os.Chmod(directory, 0o500); err != nil {
		t.Fatalf("Chmod returned %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(directory, 0o700) })

	err := writeAtomically("recording the receipt", filepath.Join(directory, "receipt.json"), []byte("{}\n"))
	if err == nil {
		t.Fatal("writeAtomically into an unwritable directory succeeded")
	}
	typed, ok := err.(problem.Problem)
	if !ok {
		t.Fatalf("expected a typed problem, got %T", err)
	}
	if !strings.Contains(typed.Message, "recording the receipt failed") {
		t.Errorf("the message does not name what failed: %q", typed.Message)
	}
	if !strings.Contains(typed.Message, "permission denied") {
		t.Errorf("the message does not say why: %q", typed.Message)
	}
	if strings.Contains(typed.Message, "atomicfile") {
		t.Errorf("the message leaks an internal package name: %q", typed.Message)
	}
}

func TestConcurrentInstallsActivateCleanlyMatchingExactlyOneInstall(t *testing.T) {
	installer := pinFixtureInstaller(t)
	var group sync.WaitGroup
	errs := make(chan error, 2)
	policies := []catalog.Policy{
		{Version: "1.0.0"},
		{Version: "1.1.0"},
	}

	for _, policy := range policies {
		group.Add(1)
		go func(p catalog.Policy) {
			defer group.Done()
			_, err := installer.Run(context.Background(), Request{
				Namespace: fixtureNamespace,
				Policy:    p,
			})
			errs <- err
		}(policy)
	}

	group.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent install failed: %v", err)
		}
	}

	active, err := installer.Store.ReadActive(fixtureNamespace)
	if err != nil {
		t.Fatalf("ReadActive failed: %v", err)
	}
	if active.Version != "1.0.0" && active.Version != "1.1.0" {
		t.Fatalf("unexpected active version: %q", active.Version)
	}

	policy, err := installer.Store.ReadPolicy(fixtureNamespace)
	if err != nil {
		t.Fatalf("ReadPolicy failed: %v", err)
	}
	if policy.PinnedVersion != active.Version {
		t.Errorf("policy pinned version = %q, want active version %q", policy.PinnedVersion, active.Version)
	}

	resolved, err := installer.Store.Resolve(fixtureNamespace, installer.Shell)
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
	if resolved.Receipt.ModuleVersion != active.Version {
		t.Errorf("resolved receipt version = %q, want active version %q", resolved.Receipt.ModuleVersion, active.Version)
	}

	replacedDir := installer.Store.VersionDir(fixtureNamespace, active.Version) + ".replaced"
	if _, err := os.Stat(replacedDir); !os.IsNotExist(err) {
		t.Errorf("replaced directory %q still exists", replacedDir)
	}
}

func TestInterruptedInstallLeavesPreviousVersionActive(t *testing.T) {
	archive := pinFixtureArchive(t)
	digest := sha256.Sum256(archive)
	artifact := fmt.Sprintf(`{"os": "linux", "arch": "amd64",
		"url": "https://origin.example/demo.tar.gz",
		"size": %d, "sha256": %q}`, len(archive), hex.EncodeToString(digest[:]))
	index := []byte(`{
		"schemaVersion": 1,
		"modules": [
			{"namespace": "demo", "path": "demo.json",
			 "channels": [{"channel": "stable", "version": "1.1.0"}]}
		]
	}`)
	// Version 1.0.0 is initially valid.
	// Version 1.1.0 has an invalid product descriptor audience, which passes catalog
	// selection and archive verification but causes receipt.Validate to fail during
	// activation after staging.
	validNamespace := []byte(fmt.Sprintf(`{
		"schemaVersion": 1,
		"namespace": "demo",
		"versions": [
			{"version": "1.0.0", "channel": "stable",
			 "compatibility": {"shell": ">=0.0.0", "protocolVersions": [1]},
			 "artifacts": [%s]},
			{"version": "1.1.0", "channel": "stable",
			 "compatibility": {"shell": ">=0.0.0", "protocolVersions": [1]},
			 "capabilities": {"product": {"audience": "unsupported_audience"}},
			 "artifacts": [%s]}
		]
	}`, artifact, artifact))

	// In failedReinstallNamespace, version 1.0.0 also carries an invalid audience,
	// exercising the rollback path where a same-version reinstall fails after
	// moving the existing installed version aside to .replaced.
	failedReinstallNamespace := []byte(fmt.Sprintf(`{
		"schemaVersion": 1,
		"namespace": "demo",
		"versions": [
			{"version": "1.0.0", "channel": "stable",
			 "compatibility": {"shell": ">=0.0.0", "protocolVersions": [1]},
			 "capabilities": {"product": {"audience": "unsupported_audience"}},
			 "artifacts": [%s]},
			{"version": "1.1.0", "channel": "stable",
			 "compatibility": {"shell": ">=0.0.0", "protocolVersions": [1]},
			 "capabilities": {"product": {"audience": "unsupported_audience"}},
			 "artifacts": [%s]}
		]
	}`, artifact, artifact))

	currentNamespace := validNamespace
	store := modules.NewStore(t.TempDir())
	installer := Installer{
		Store: store,
		Client: catalog.Client{
			Origin: "https://origin.example",
			HTTP: &http.Client{
				Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
					switch r.URL.Path {
					case "/" + catalog.IndexPath:
						return jsonResponse(index), nil
					case "/demo.json":
						return jsonResponse(currentNamespace), nil
					default:
						return jsonResponse(archive), nil
					}
				}),
			},
		},
		Shell: fixtureShell(),
	}

	// 1. Install version 1.0.0 successfully.
	installed, err := installer.Run(context.Background(), Request{
		Namespace: fixtureNamespace,
		Policy:    catalog.Policy{Version: "1.0.0"},
	})
	if err != nil {
		t.Fatalf("initial install of 1.0.0 failed: %v", err)
	}
	if installed.Version != "1.0.0" {
		t.Fatalf("initial installed version = %q, want 1.0.0", installed.Version)
	}

	// Place a sentinel file in the installed version directory. When a same-version
	// reinstall moves the directory to .replaced and later restores it on failure,
	// the sentinel proves that the original installation was restored rather than
	// replaced by a fresh extraction.
	versionDir := store.VersionDir(fixtureNamespace, "1.0.0")
	sentinelPath := filepath.Join(versionDir, "installed-sentinel.txt")
	if err := os.WriteFile(sentinelPath, []byte("sentinel"), 0o644); err != nil {
		t.Fatalf("writing sentinel file failed: %v", err)
	}

	// 2. Attempt a same-version reinstall of 1.0.0 with the invalid descriptor.
	// This exercises moving the existing version directory to .replaced and
	// rolling back by renaming .replaced back to the version directory.
	currentNamespace = failedReinstallNamespace
	_, err = installer.Run(context.Background(), Request{
		Namespace: fixtureNamespace,
		Policy:    catalog.Policy{Version: "1.0.0"},
	})
	if err == nil {
		t.Fatal("expected reinstall of 1.0.0 to fail due to invalid product descriptor")
	}

	// Verify rollback after same-version reinstall failure: version 1.0.0 remains active and resolvable.
	activeAfterReinstall, err := store.ReadActive(fixtureNamespace)
	if err != nil {
		t.Fatalf("ReadActive after failed same-version reinstall returned %v", err)
	}
	if activeAfterReinstall.Version != "1.0.0" {
		t.Errorf("active version after failed same-version reinstall = %q, want 1.0.0", activeAfterReinstall.Version)
	}

	policyAfterReinstall, err := store.ReadPolicy(fixtureNamespace)
	if err != nil {
		t.Fatalf("ReadPolicy after failed same-version reinstall returned %v", err)
	}
	if policyAfterReinstall.PinnedVersion != "1.0.0" {
		t.Errorf("policy after failed same-version reinstall pinned = %q, want 1.0.0", policyAfterReinstall.PinnedVersion)
	}

	resolvedAfterReinstall, err := store.Resolve(fixtureNamespace, fixtureShell())
	if err != nil {
		t.Fatalf("Resolve after failed same-version reinstall returned %v", err)
	}
	if resolvedAfterReinstall.Receipt.ModuleVersion != "1.0.0" {
		t.Errorf("resolved version = %q, want 1.0.0", resolvedAfterReinstall.Receipt.ModuleVersion)
	}

	if sentinelContent, err := os.ReadFile(sentinelPath); err != nil || string(sentinelContent) != "sentinel" {
		t.Errorf("sentinel file in restored version directory missing or altered: %v", err)
	}

	replacedDir := versionDir + ".replaced"
	if _, err := os.Stat(replacedDir); !os.IsNotExist(err) {
		t.Errorf("replaced directory %q still exists after same-version rollback", replacedDir)
	}

	// 3. Attempt to install version 1.1.0 which also fails receipt validation.
	// This exercises the rollback path where a new version fails without displacing
	// an existing directory of the same version.
	_, err = installer.Run(context.Background(), Request{
		Namespace: fixtureNamespace,
		Policy:    catalog.Policy{Version: "1.1.0"},
	})
	if err == nil {
		t.Fatal("expected install of 1.1.0 to fail due to invalid product descriptor")
	}

	// Verify rollback: previous version 1.0.0 still remains active and resolvable.
	activeAfterNewVersion, err := store.ReadActive(fixtureNamespace)
	if err != nil {
		t.Fatalf("ReadActive after failed install returned %v", err)
	}
	if activeAfterNewVersion.Version != "1.0.0" {
		t.Errorf("active version after failed install = %q, want 1.0.0", activeAfterNewVersion.Version)
	}

	policyAfterNewVersion, err := store.ReadPolicy(fixtureNamespace)
	if err != nil {
		t.Fatalf("ReadPolicy after failed install returned %v", err)
	}
	if policyAfterNewVersion.PinnedVersion != "1.0.0" {
		t.Errorf("policy after failed install pinned = %q, want 1.0.0", policyAfterNewVersion.PinnedVersion)
	}

	resolvedAfterNewVersion, err := store.Resolve(fixtureNamespace, fixtureShell())
	if err != nil {
		t.Fatalf("Resolve after failed install returned %v", err)
	}
	if resolvedAfterNewVersion.Receipt.ModuleVersion != "1.0.0" {
		t.Errorf("resolved version = %q, want 1.0.0", resolvedAfterNewVersion.Receipt.ModuleVersion)
	}

	newVersionDir := store.VersionDir(fixtureNamespace, "1.1.0")
	if _, err := os.Stat(newVersionDir); !os.IsNotExist(err) {
		t.Errorf("failed version directory %q was not removed after rollback", newVersionDir)
	}
}
