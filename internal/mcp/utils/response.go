package utils

import (
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
)

// MCPResponse defines the standard structure for tool responses.
type MCPResponse struct {
	Data       any      `json:"data"`
	Error      any      `json:"error"`
	NextSteps  []string `json:"next_steps,omitempty"`
	Conclusion string   `json:"conclusion"`
}

// NewMCPResponse creates a successful MCPResponse and wraps it in mcp.CallToolResult.
func NewMCPResponse(data any, conclusion string, nextSteps []string) (*mcp.CallToolResult, error) {
	response := MCPResponse{
		Data:       data,
		Error:      nil,
		NextSteps:  nextSteps,
		Conclusion: conclusion,
	}
	jsonBytes, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}
	return mcp.NewToolResultText(string(jsonBytes)), nil
}

// NewMCPErrorResponse creates an error MCPResponse and wraps it in mcp.CallToolResult.
func NewMCPErrorResponse(err error, conclusion string) *mcp.CallToolResult {
	// Callers sometimes reach an error branch where the local err is already
	// nil (e.g. a validation failure after a successful lookup). Calling
	// err.Error() there panics, and the recovered panic replaces the intended
	// diagnostic with a generic failure — so fall back to the conclusion.
	message := conclusion
	if err != nil {
		message = err.Error()
	}
	response := MCPResponse{
		Data:       nil,
		Error:      map[string]string{"message": message},
		NextSteps:  nil,
		Conclusion: conclusion,
	}
	jsonBytes, _ := json.Marshal(response)
	return mcp.NewToolResultError(string(jsonBytes))
}
