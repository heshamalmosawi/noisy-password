package test

import (
	"password-fuzzifier/internal"
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
		{name: "all lowercase", charset: "all", lowercase: true, want: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"},
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
