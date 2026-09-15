package component

import (
	"fmt"
	"os"
	"sync"
	"time"
)

var (
	debugOnce    sync.Once
	debugEnabled bool
)

func isDebugEnabled() bool {
	debugOnce.Do(func() {
		debugEnabled = os.Getenv("MCP_DEBUG") == "true"
	})
	return debugEnabled
}

// debugf writes a diagnostic line to stderr when MCP_DEBUG=true.
// Never include tokens, passwords, or auth headers in the format string or args.
func debugf(format string, args ...any) {
	if !isDebugEnabled() {
		return
	}
	ts := time.Now().Format("15:04:05.000")
	fmt.Fprintf(os.Stderr, "[MCP_DEBUG %s] %s\n", ts, fmt.Sprintf(format, args...))
}
