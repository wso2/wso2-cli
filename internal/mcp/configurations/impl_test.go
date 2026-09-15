package configurations

import (
	"testing"

	"github.com/wso2/integration-platform-tools/internal/cmd/common"
)

// The tool schema's enum advertised "env variables" while the handler compared
// against common.MountTypeEnv ("env variable"). mcp-go does not validate enums
// server-side, so the plural reached the handler, matched neither the env nor
// the file branch, and the requested configuration was silently dropped — while
// CreateConfigMapping was still called and the tool reported success and
// restarted the component.
func TestNormalizeMountType(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		// canonical
		{"env variable", common.MountTypeEnv},
		{"file mount", common.MountTypeFile},

		// the plural the schema enum used to advertise — the actual bug
		{"env variables", common.MountTypeEnv},

		// shorthand and casing/whitespace tolerance
		{"env", common.MountTypeEnv},
		{"file", common.MountTypeFile},
		{"ENV VARIABLE", common.MountTypeEnv},
		{"  file mount  ", common.MountTypeFile},
		{"Env Variables", common.MountTypeEnv},
	}

	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			if got := normalizeMountType(tc.in); got != tc.want {
				t.Errorf("normalizeMountType(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

// Unrecognised values must NOT normalise onto a valid constant — the handler
// rejects them explicitly rather than falling through to a silent no-op.
func TestNormalizeMountTypeRejectsUnknown(t *testing.T) {
	for _, in := range []string{"", "volume", "configmap", "secret", "env-variable"} {
		got := normalizeMountType(in)
		if got == common.MountTypeEnv || got == common.MountTypeFile {
			t.Errorf("normalizeMountType(%q) = %q — must not map an unknown value onto a valid mount type", in, got)
		}
	}
}

// Buildpack detection drives the mount-type warnings. DisplayType prefixes
// follow the convention used in internal/cmd/build/create/impl.go.
func TestBuildpackDetection(t *testing.T) {
	ballerina := []string{"ballerinaService", "ballerinaEventHandler", "BallerinaService", "ballerina"}
	mi := []string{"miRestApi", "miEventHandler", "miApiService", "miCronjob", "miJob", "MiRestApi"}
	neither := []string{"", "byocService", "proxy", "buildpackService", "service"}

	for _, d := range ballerina {
		if !isBallerinaComponent(d) {
			t.Errorf("isBallerinaComponent(%q) = false, want true", d)
		}
		if isMicroIntegratorComponent(d) {
			t.Errorf("isMicroIntegratorComponent(%q) = true, want false", d)
		}
	}
	for _, d := range mi {
		if !isMicroIntegratorComponent(d) {
			t.Errorf("isMicroIntegratorComponent(%q) = false, want true", d)
		}
		if isBallerinaComponent(d) {
			t.Errorf("isBallerinaComponent(%q) = true, want false", d)
		}
	}
	for _, d := range neither {
		if isBallerinaComponent(d) || isMicroIntegratorComponent(d) {
			t.Errorf("%q classified as ballerina=%v mi=%v, want both false",
				d, isBallerinaComponent(d), isMicroIntegratorComponent(d))
		}
	}
}

// The Ballerina runtime reads Config.toml only from this path; a file mounted
// anywhere else applies cleanly and is then ignored.
func TestBallerinaConfigMountPath(t *testing.T) {
	if BallerinaConfigMountPath != "/config/Config.toml" {
		t.Errorf("BallerinaConfigMountPath = %q, want /config/Config.toml", BallerinaConfigMountPath)
	}
}
