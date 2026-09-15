package utils

import (
	"fmt"
	"os"
	"sync"
	"time"
)

var (
	mcpDebugOnce    sync.Once
	mcpDebugEnabled bool
)

func isMCPDebugEnabled() bool {
	mcpDebugOnce.Do(func() {
		mcpDebugEnabled = os.Getenv("MCP_DEBUG") == "true"
	})
	return mcpDebugEnabled
}

// MCPDebugf writes a diagnostic line to stderr when MCP_DEBUG=true.
// Never include tokens, passwords, or auth headers in the format string or args.
func MCPDebugf(format string, args ...any) {
	if !isMCPDebugEnabled() {
		return
	}
	ts := time.Now().Format("15:04:05.000")
	fmt.Fprintf(os.Stderr, "[MCP_DEBUG %s] %s\n", ts, fmt.Sprintf(format, args...))
}
