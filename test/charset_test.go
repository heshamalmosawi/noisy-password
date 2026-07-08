package test

import (
	"password-fuzzifier/internal"
	"strings"
	"testing"
)

func TestGetCharset(t *testing.T) {
	tests := []struct {
		name        string
		charset     string
		lowercase   bool
		want        string
		expectError bool
	}{
		{name: "alphabet lowercase", charset: "alphabet", lowercase: true, want: "abcdefghijklmnopqrstuvwxyz"},
		{name: "alphabet uppercase", charset: "alphabet", lowercase: false, want: "ABCDEFGHIJKLMNOPQRSTUVWXYZ"},
		{name: "numeric lowercase", charset: "numeric", lowercase: true, want: "0123456789"},
		{name: "numeric uppercase", charset: "numeric", lowercase: false, want: "0123456789"},
		{name: "alphanumeric lowercase", charset: "alphanumeric", lowercase: true, want: "abcdefghijklmnopqrstuvwxyz0123456789"},
		{name: "alphanumeric uppercase", charset: "alphanumeric", lowercase: false, want: "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"},
		{name: "all lowercase", charset: "all", lowercase: true, want: "abcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*()-_=+[]{}|;:,.<>?/"},
		{name: "all full", charset: "all", lowercase: false, want: "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*()-_=+[]{}|;:,.<>?/"},
		{name: "invalid charset", charset: "invalid", lowercase: false, expectError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := internal.GetCharset(tt.charset, tt.lowercase)
			if tt.expectError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

// TestLowercaseEnforced guards the --lowercase contract: every charset's
// lowercase variant must contain no uppercase letters.
func TestLowercaseEnforced(t *testing.T) {
	for _, name := range []string{"alphabet", "numeric", "alphanumeric", "all"} {
		got, err := internal.GetCharset(name, true)
		if err != nil {
			t.Fatalf("GetCharset(%q, true): %v", name, err)
		}
		if strings.ContainsAny(got, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") {
			t.Errorf("lowercase charset %q contains uppercase letters: %q", name, got)
		}
	}
}
