package test

import (
	"password-fuzzifier/internal"
	"testing"
)

// helper to check if all runes in s are in charset
func isValidpassword(s []rune, charset string) bool {
	charsetRunes := make(map[rune]bool)
	for _, r := range charset {
		charsetRunes[r] = true
	}
	for _, r := range s {
		if !charsetRunes[r] {
			return false
		}
	}
	return true
}

func TestGeneratepassword(t *testing.T) {
	tests := []struct {
		name        string
		charset     string
		length      int
		expectError bool
	}{
		{
			name:    "valid alpha",
			charset: "abcdefghijklmnopqrstuvwxyz",
			length:  10,
		},
		{
			name:    "valid numeric",
			charset: "0123456789",
			length:  6,
		},
		{
			name:    "length zero",
			charset: "abc",
			length:  0,
		},
		{
			name:        "empty charset",
			charset:     "",
			length:      5,
			expectError: true,
		},
		{
			name:        "negative length",
			charset:     "abc",
			length:      -1,
			expectError: true,
		},
		{
			name:    "custom charset",
			charset: "abc123",
			length:  8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := internal.GeneratePassword(tt.charset, tt.length)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error, got nil & password %q", got)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if len(got) != tt.length {
				t.Errorf("expected length %d, got %d", tt.length, len(got))
			}

			if !isValidpassword(got, tt.charset) {
				t.Errorf("password %q contains characters not in charset %q", got, tt.charset)
			}
		})
	}
}
