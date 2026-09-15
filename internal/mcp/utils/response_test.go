package utils

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func textOf(t *testing.T, result *mcp.CallToolResult) string {
	t.Helper()
	require.Len(t, result.Content, 1)
	tc, ok := mcp.AsTextContent(result.Content[0])
	require.True(t, ok, "content item must be TextContent")
	return tc.Text
}

func TestNewMCPResponse_Structure(t *testing.T) {
	data := map[string]string{"project": "test-project"}
	conclusion := "Found 1 project."
	nextSteps := []string{"Run get_integrations to list integrations"}

	result, err := NewMCPResponse(data, conclusion, nextSteps)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.IsError)

	var got MCPResponse
	require.NoError(t, json.Unmarshal([]byte(textOf(t, result)), &got))

	assert.Equal(t, conclusion, got.Conclusion)
	assert.Nil(t, got.Error)
	assert.Equal(t, nextSteps, got.NextSteps)
}

func TestNewMCPResponse_NoNextSteps(t *testing.T) {
	result, err := NewMCPResponse("some-data", "Done.", nil)
	require.NoError(t, err)

	var got MCPResponse
	require.NoError(t, json.Unmarshal([]byte(textOf(t, result)), &got))
	assert.Nil(t, got.NextSteps)
}

func TestNewMCPErrorResponse_IsError(t *testing.T) {
	e := errors.New("something went wrong")
	result := NewMCPErrorResponse(e, "Failed to fetch projects.")

	require.NotNil(t, result)
	assert.True(t, result.IsError)

	var got MCPResponse
	require.NoError(t, json.Unmarshal([]byte(textOf(t, result)), &got))

	assert.Equal(t, "Failed to fetch projects.", got.Conclusion)
	assert.Nil(t, got.Data)

	errMap, ok := got.Error.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "something went wrong", errMap["message"])
}

// Several handlers reach a validation-failure branch where the local err is
// already nil (the preceding lookup succeeded). Passing that nil to
// NewMCPErrorResponse used to call err.Error() and panic; with WithRecovery()
// enabled the server survived but the intended diagnostic was replaced by a
// generic recovered-panic error. internal/mcp/logs/impl.go did exactly this for
// unsupported component types.
func TestNewMCPErrorResponseToleratesNilError(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("panicked on nil error: %v", r)
		}
	}()

	result := NewMCPErrorResponse(nil, "Failed to get execution logs.")
	require.NotNil(t, result)
	assert.True(t, result.IsError)

	var got MCPResponse
	require.NoError(t, json.Unmarshal([]byte(textOf(t, result)), &got))

	// With no error to report, the conclusion stands in as the message so the
	// caller still learns something.
	errMap, ok := got.Error.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "Failed to get execution logs.", errMap["message"])
	assert.Equal(t, "Failed to get execution logs.", got.Conclusion)
}
