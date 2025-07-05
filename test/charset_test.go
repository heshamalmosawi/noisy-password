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
		// "alphabet" cases
		{
			name:      "alphabet lowercase true",
			charset:   "alphabet",
			lowercase: true,
			want:      "abcdefghijklmnopqrstuvwxyz",
		},
		{
			name:      "alphabet lowercase false",
			charset:   "alphabet",
			lowercase: false,
			want:      "ABCDEFGHIJKLMNOPQRSTUVWXYZ",
		},

		// "numeric" case
		{
			name:      "numeric",
			charset:   "numeric",
			lowercase: true, // Should not affect numeric
			want:      "0123456789",
		},
		{
			name:      "numeric lowercase false",
			charset:   "numeric",
			lowercase: false,
			want:      "0123456789",
		},

		// "alphanumeric" cases
		{
			name:      "alphanumeric lowercase true",
			charset:   "alphanumeric",
			lowercase: true,
			want:      "abcdefghijklmnopqrstuvwxyz0123456789",
		},
		{
			name:      "alphanumeric lowercase false",
			charset:   "alphanumeric",
			lowercase: false,
			want:      "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789",
		},

		// "all" cases
		{
			name:      "all lowercase true",
			charset:   "all",
			lowercase: true,
			want:      "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789",
		},
		{
			name:      "all lowercase false",
			charset:   "all",
			lowercase: false,
			want:      "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*()-_=+[]{}|;:,.<>?/",
		},

		// Invalid case
		{
			name:        "invalid charset",
			charset:     "invalid",
			lowercase:   false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := internal.GetCharset(tt.charset, tt.lowercase)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if got != tt.want {
					t.Errorf("got %q, want %q", got, tt.want)
				}
			}
		})
	}
}
