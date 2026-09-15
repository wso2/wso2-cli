package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseOutputFormat(t *testing.T) {
	tests := []struct {
		input   string
		want    OutputFormat
		wantErr bool
	}{
		{"", OutputFormatTable, false},
		{"table", OutputFormatTable, false},
		{"TABLE", OutputFormatTable, false},
		{"Table", OutputFormatTable, false},
		{"  table  ", OutputFormatTable, false},
		{"json", OutputFormatJSON, false},
		{"JSON", OutputFormatJSON, false},
		{"Json", OutputFormatJSON, false},
		{"  json  ", OutputFormatJSON, false},
		{"csv", "", true},
		{"yaml", "", true},
		{"invalid", "", true},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got, err := ParseOutputFormat(tc.input)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid output format")
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.want, got)
			}
		})
	}
}

func TestOutputFormat_IsStructured(t *testing.T) {
	assert.True(t, OutputFormatJSON.IsStructured())
	assert.False(t, OutputFormatTable.IsStructured())
	assert.False(t, OutputFormat("").IsStructured())
}

func TestRenderStructured_JSON(t *testing.T) {
	data := map[string]string{"key": "value"}
	out, err := RenderStructured(OutputFormatJSON, data)
	require.NoError(t, err)
	assert.Contains(t, out, `"key"`)
	assert.Contains(t, out, `"value"`)
}

func TestRenderStructured_Table(t *testing.T) {
	_, err := RenderStructured(OutputFormatTable, map[string]string{"k": "v"})
	assert.Error(t, err, "table format is not a structured format and should error")
}
