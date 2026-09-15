package clikeyring

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/wso2/integration-platform-tools/pkg/util/constants"
	"github.com/zalando/go-keyring"
)

func useFileStorage() bool {
	return os.Getenv("SKIP_KEYRING") == "true"
}

func setToFile(entryName, value string) error {
	homeDir, _ := os.UserHomeDir()
	filePath := filepath.Join(homeDir, constants.CLI_HOME_DIR, constants.CLI_HOME_AUTH_DIR, entryName)
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(filePath, []byte(value), 0600); err != nil {
		return fmt.Errorf("failed to write to file %s: %w", entryName, err)
	}
	return nil
}

func getFromFile(entryName string) (string, error) {
	homeDir, _ := os.UserHomeDir()
	filePath := filepath.Join(homeDir, constants.CLI_HOME_DIR, constants.CLI_HOME_AUTH_DIR, entryName)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", fmt.Errorf("file %s does not exist", entryName)
	}
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", entryName, err)
	}
	return string(data), nil
}

func deleteFile(entryName string) error {
	homeDir, _ := os.UserHomeDir()
	filePath := filepath.Join(homeDir, constants.CLI_HOME_DIR, constants.CLI_HOME_AUTH_DIR, entryName)
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file %s: %w", entryName, err)
	}
	return nil
}

func Set(serviceName, entryName, value string) error {
	if useFileStorage() {
		return setToFile(entryName, value)
	}
	// Attempt system keyring; fall back to file storage when unavailable
	// (e.g. headless environments without dbus).
	if err := keyring.Set(serviceName, entryName, value); err != nil {
		return setToFile(entryName, value)
	}
	return nil
}

func Get(serviceName, entryName string) (string, error) {
	if useFileStorage() {
		return getFromFile(entryName)
	}
	// Attempt system keyring; fall back to file storage when unavailable.
	if value, err := keyring.Get(serviceName, entryName); err == nil {
		return value, nil
	}
	return getFromFile(entryName)
}

func Delete(serviceName, entryName string) error {
	if useFileStorage() {
		return deleteFile(entryName)
	}
	// Try both locations — the entry may exist in either depending on whether
	// the keyring was available when it was written.
	_ = keyring.Delete(serviceName, entryName)
	return deleteFile(entryName)
}
