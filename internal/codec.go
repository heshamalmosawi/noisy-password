package internal

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"
)

// EncodePasscode renders a passcode as the Base64 text that gets persisted.
func EncodePasscode(passcode []rune) string {
	return base64.StdEncoding.EncodeToString([]byte(string(passcode)))
}

// DecodePasscode is the exact inverse of EncodePasscode. Surrounding whitespace
// (e.g. a trailing newline in a file) is tolerated.
func DecodePasscode(encoded string) ([]rune, error) {
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(encoded))
	if err != nil {
		return nil, fmt.Errorf("decoding passcode: %w", err)
	}
	return []rune(string(decoded)), nil
}

// SavePasscode writes a passcode to path as Base64 with owner-only permissions.
func SavePasscode(path string, passcode []rune) error {
	if err := os.WriteFile(path, []byte(EncodePasscode(passcode)), 0o600); err != nil {
		return fmt.Errorf("writing %q: %w", path, err)
	}
	return nil
}

// LoadPasscode reads and decodes a passcode previously written by SavePasscode.
func LoadPasscode(path string) ([]rune, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %q: %w", path, err)
	}
	passcode, err := DecodePasscode(string(data))
	if err != nil {
		return nil, fmt.Errorf("file %q: %w", path, err)
	}
	if len(passcode) == 0 {
		return nil, fmt.Errorf("file %q contains no passcode", path)
	}
	return passcode, nil
}
