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
