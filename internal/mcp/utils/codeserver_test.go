package utils

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/wso2/integration-platform-tools/pkg/api"
)

// The binding is only meaningful inside the cloud editor. BASE_URL is the same
// signal get_started uses to switch to Source Control UI git instructions, so
// the two stay consistent.
func TestIsCloudEditor(t *testing.T) {
	original, had := os.LookupEnv("BASE_URL")
	t.Cleanup(func() {
		if had {
			os.Setenv("BASE_URL", original)
		} else {
			os.Unsetenv("BASE_URL")
		}
	})

	os.Unsetenv("BASE_URL")
	if IsCloudEditor() {
		t.Error("IsCloudEditor() = true with BASE_URL unset, want false")
	}

	os.Setenv("BASE_URL", "")
	if IsCloudEditor() {
		t.Error("IsCloudEditor() = true with BASE_URL empty, want false")
	}

	os.Setenv("BASE_URL", "https://cloud.example")
	if !IsCloudEditor() {
		t.Error("IsCloudEditor() = false with BASE_URL set, want true")
	}
}

// Argument validation must happen before any client call, so a caller that
// passes an incomplete set gets a clear error rather than a nil dereference or a
// malformed mutation with empty ids.
func TestBindCodeServerValidatesArguments(t *testing.T) {
	ctx := context.Background()
	org := &api.Organization{ID: "53019", UUID: "u-1", Handle: "acme"}

	cases := []struct {
		name        string
		org         *api.Organization
		projectID   string
		componentID string
		wantErr     string
	}{
		{"nil org", nil, "p-1", "c-1", "organization is required"},
		{"empty project id", org, "", "c-1", "project id and component id are required"},
		{"empty component id", org, "p-1", "", "project id and component id are required"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := BindCodeServer(ctx, tc.org, tc.projectID, tc.componentID, "")
			if err == nil {
				t.Fatal("expected an error")
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("error = %q, want it to contain %q", err, tc.wantErr)
			}
		})
	}
}

// The Code Server image is looked up from the samples registry by name rather
// than hardcoded, so the name must match what the registry publishes.
func TestCodeServerSampleName(t *testing.T) {
	if codeServerSampleName != "Code Server" {
		t.Errorf("codeServerSampleName = %q, want \"Code Server\"", codeServerSampleName)
	}
}
