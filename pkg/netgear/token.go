package netgear

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// TokenManager interface for managing authentication tokens
type TokenManager interface {
	GetToken(host string) (string, error)
	SetToken(host string, token string) error
	RemoveToken(host string) error
}

// FileTokenManager implements TokenManager using local file storage
type FileTokenManager struct {
	tokenDir string
}

// NewFileTokenManager creates a new file-based token manager
func NewFileTokenManager(tokenDir string) *FileTokenManager {
	if tokenDir == "" {
		// Default to user's home directory
		homeDir, _ := os.UserHomeDir()
		tokenDir = filepath.Join(homeDir, ".netgear")
	}
	return &FileTokenManager{tokenDir: tokenDir}
}

// GetToken retrieves a stored token for the given host
func (f *FileTokenManager) GetToken(host string) (string, error) {
	tokenFile := f.getTokenFile(host)
	data, err := os.ReadFile(tokenFile)
	if err != nil {
		return "", err
	}

	var token map[string]string
	if err := json.Unmarshal(data, &token); err != nil {
		return "", err
	}

	return token["token"], nil
}

// SetToken stores a token for the given host
func (f *FileTokenManager) SetToken(host string, token string) error {
	if err := os.MkdirAll(f.tokenDir, 0700); err != nil {
		return err
	}

	tokenFile := f.getTokenFile(host)
	tokenData := map[string]string{"token": token}
	data, err := json.Marshal(tokenData)
	if err != nil {
		return err
	}

	return os.WriteFile(tokenFile, data, 0600)
}

// RemoveToken removes a stored token for the given host
func (f *FileTokenManager) RemoveToken(host string) error {
	tokenFile := f.getTokenFile(host)
	return os.Remove(tokenFile)
}

// getTokenFile returns the path to the token file for a given host
func (f *FileTokenManager) getTokenFile(host string) string {
	// Replace dots and colons with underscores to create safe filename
	safeHost := filepath.Base(host)
	return filepath.Join(f.tokenDir, fmt.Sprintf("token_%s.json", safeHost))
}