package test

import (
	"os"
	"path/filepath"
	"password-fuzzifier/internal"
	"testing"
)

// TestEncodeDecodeRoundTrip verifies the saved representation is a lossless
// inverse for every kind of passcode, including special characters.
func TestEncodeDecodeRoundTrip(t *testing.T) {
	cases := []string{
		"1",
		"6398",
		"705128",
		"ABC",
		"abcXYZ123",
		"P@ssw0rd!",
		"!@#$%^&*()-_=+[]{}|;:,.<>?/",
		"VeryLongPasswordWithSpecialChars!@#$%^&*()1234567890",
	}
	for _, c := range cases {
		encoded := internal.EncodePasscode([]rune(c))
		decoded, err := internal.DecodePasscode(encoded)
		if err != nil {
			t.Fatalf("DecodePasscode(%q): %v", encoded, err)
		}
		if string(decoded) != c {
			t.Errorf("round trip: got %q, want %q", string(decoded), c)
		}
	}
}

// TestDecodeToleratesWhitespace ensures a trailing newline (as written to a
// file) does not corrupt the decoded passcode.
func TestDecodeToleratesWhitespace(t *testing.T) {
	encoded := internal.EncodePasscode([]rune("6398"))
	decoded, err := internal.DecodePasscode("  " + encoded + "\n")
	if err != nil {
		t.Fatal(err)
	}
	if string(decoded) != "6398" {
		t.Errorf("got %q, want %q", string(decoded), "6398")
	}
}

func TestDecodeRejectsGarbage(t *testing.T) {
	if _, err := internal.DecodePasscode("not valid base64!!!"); err == nil {
		t.Error("expected error decoding invalid base64")
	}
}

// TestSaveLoadFileRoundTrip exercises the real on-disk persistence path.
func TestSaveLoadFileRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "code.enc")

	want := []rune("8120")
	if err := internal.SavePasscode(path, want); err != nil {
		t.Fatalf("SavePasscode: %v", err)
	}

	// File should hold exactly the Base64 encoding (no extra wrapping).
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != internal.EncodePasscode(want) {
		t.Errorf("file contents %q != encoding %q", string(raw), internal.EncodePasscode(want))
	}

	got, err := internal.LoadPasscode(path)
	if err != nil {
		t.Fatalf("LoadPasscode: %v", err)
	}
	if string(got) != string(want) {
		t.Errorf("loaded %q, want %q", string(got), string(want))
	}
}

func TestLoadPasscodeErrors(t *testing.T) {
	dir := t.TempDir()

	if _, err := internal.LoadPasscode(filepath.Join(dir, "missing.enc")); err == nil {
		t.Error("expected error for missing file")
	}

	empty := filepath.Join(dir, "empty.enc")
	if err := os.WriteFile(empty, []byte(""), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := internal.LoadPasscode(empty); err == nil {
		t.Error("expected error for empty file")
	}
}
