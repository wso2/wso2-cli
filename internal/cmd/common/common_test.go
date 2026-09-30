package common

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateNameText(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr string
	}{
		{"valid simple", "MyProject", ""},
		{"valid with hyphen", "My-Project-123", ""},
		{"valid min length", "abc", ""},
		{"valid max length", strings.Repeat("a", 60), ""},
		{"too short", "ab", "at least 3 characters"},
		{"empty", "", "at least 3 characters"},
		{"too long", strings.Repeat("a", 61), "more than 60 characters"},
		{"starts with digit", "1project", "start with an alphabetic letter"},
		{"starts with hyphen", "-project", "start with an alphabetic letter"},
		{"contains slash", "my/project", "special characters"},
		{"contains dot", "my.project", "special characters"},
		{"contains space", "my project", ""}, // spaces are permitted by the regex
		{"contains at-sign", "my@project", "special characters"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateNameText(tc.input)
			if tc.wantErr == "" {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErr)
			}
		})
	}
}

func TestParseKeyValStrPair(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantKey string
		wantVal string
		wantErr bool
	}{
		{"simple pair", "key=value", "key", "value", false},
		{"empty value", "key=", "key", "", false},
		{"no equals", "keyvalue", "", "", true},
		{"multiple equals", "key=val=extra", "", "", true},
		{"empty string", "", "", "", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseKeyValStrPair(tc.input)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.wantKey, result.Key)
				assert.Equal(t, tc.wantVal, result.Val)
			}
		})
	}
}

// The wire values carry a space ("env variable"), which nobody types behind a
// flag, and the command's own examples use the hyphenated plural. Both must
// land on the value the API accepts.
func TestMatchMountType(t *testing.T) {
	for _, tc := range []struct {
		input string
		want  string
	}{
		{MountTypeEnv, MountTypeEnv},
		{"env variable", MountTypeEnv},
		{"env-variables", MountTypeEnv},
		{"ENV_VARIABLE", MountTypeEnv},
		{"env", MountTypeEnv},
		{MountTypeFile, MountTypeFile},
		{"file-mount", MountTypeFile},
		{"File Mount", MountTypeFile},
	} {
		t.Run(tc.input, func(t *testing.T) {
			got, ok := MatchMountType(tc.input)
			if !ok {
				t.Fatalf("MatchMountType(%q) did not match", tc.input)
			}
			if got != tc.want {
				t.Errorf("MatchMountType(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}

	for _, input := range []string{"", "   ", "volume", "configmap"} {
		t.Run("reject "+input, func(t *testing.T) {
			if got, ok := MatchMountType(input); ok {
				t.Errorf("MatchMountType(%q) matched %q, want no match", input, got)
			}
		})
	}
}

func TestMatchConfigType(t *testing.T) {
	for _, tc := range []struct {
		input string
		want  string
	}{
		{ConfigTypeConfigMap, ConfigTypeConfigMap},
		{"config-map", ConfigTypeConfigMap},
		{"ConfigMap", ConfigTypeConfigMap},
		{"config_map", ConfigTypeConfigMap},
		{ConfigTypeSecret, ConfigTypeSecret},
		{"Secret", ConfigTypeSecret},
	} {
		t.Run(tc.input, func(t *testing.T) {
			got, ok := MatchConfigType(tc.input)
			if !ok {
				t.Fatalf("MatchConfigType(%q) did not match", tc.input)
			}
			if got != tc.want {
				t.Errorf("MatchConfigType(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}

	// The old check was a substring test against the joined list, so "config"
	// and "ret" passed as valid types and were sent on unchanged.
	for _, input := range []string{"", "config", "ret", "map-config"} {
		t.Run("reject "+input, func(t *testing.T) {
			if got, ok := MatchConfigType(input); ok {
				t.Errorf("MatchConfigType(%q) matched %q, want no match", input, got)
			}
		})
	}
}
