package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

var infoLogger *log.Logger
var errorLogger *log.Logger

// TODO: Implement the logger functions
// Commented due to restriction on file creation when spawned from the vscode extension host

// Info logs an informational message
func Info(message string) {
	infoLogger.Println(message)
}

// Error logs an error message
func Error(message string) {
	errorLogger.Println(message)
}

func init() {
	files, err := filepath.Glob(filepath.Join(os.TempDir(), "wso2ip-rpc-server-log-*"))

	if err != nil {
		log.Fatal(err)
	}

	logfile, err := os.OpenFile(
		filepath.Join(os.TempDir(), fmt.Sprintf("wso2ip-rpc-server-log-%d", len(files))),
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
		0666,
	)

	if err != nil {
		log.Fatal(err)
	}

	infoLogger = log.New(logfile, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	errorLogger = log.New(logfile, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}
