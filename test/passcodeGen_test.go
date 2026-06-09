package test

import (
	"math/rand"
	"password-fuzzifier/internal"
	"testing"
)

// containsAll reports whether every rune in s belongs to charset.
func containsAll(s []rune, charset string) bool {
	set := make(map[rune]bool)
	for _, r := range charset {
		set[r] = true
	}
	for _, r := range s {
		if !set[r] {
			return false
		}
	}
	return true
}

func TestGeneratePassword(t *testing.T) {
	r := rand.New(rand.NewSource(1))

	tests := []struct {
		name        string
		charset     string
		length      int
		expectError bool
	}{
		{name: "alpha", charset: "abcdefghijklmnopqrstuvwxyz", length: 10},
		{name: "numeric", charset: "0123456789", length: 6},
		{name: "length zero", charset: "abc", length: 0},
		{name: "custom", charset: "abc123", length: 8},
		{name: "empty charset", charset: "", length: 5, expectError: true},
		{name: "negative length", charset: "abc", length: -1, expectError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := internal.GeneratePassword(tt.charset, tt.length, r)
			if tt.expectError {
				if err == nil {
					t.Fatalf("expected error, got password %q", string(got))
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != tt.length {
				t.Errorf("expected length %d, got %d", tt.length, len(got))
			}
			if !containsAll(got, tt.charset) {
				t.Errorf("password %q has characters outside charset %q", string(got), tt.charset)
			}
		})
	}
}
