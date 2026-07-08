package test

import (
	"path/filepath"
	"password-fuzzifier/internal"
	"testing"
)

// TestFullPipeline is the critical end-to-end safety net. Using the real
// production RNG (CryptoRand), it runs the entire chain a passcode travels
// through in actual use and asserts every stage stays aligned with the
// originally generated passcode:
//
//	generate -> save(base64,file) -> load -> decode -> keystrokes -> replay
//
// If any link drifts, the user would type the wrong code into their device,
// so this is checked exhaustively across charsets and lengths.
func TestFullPipeline(t *testing.T) {
	rng := internal.CryptoRand{}
	dir := t.TempDir()

	charsets := []struct {
		name      string
		lowercase bool
	}{
		{"numeric", false},
		{"alphabet", true},
		{"alphanumeric", false},
		{"all", false},
	}

	for iter := 0; iter < 2000; iter++ {
		cs := charsets[iter%len(charsets)]
		charSet, err := internal.GetCharset(cs.name, cs.lowercase)
		if err != nil {
			t.Fatalf("GetCharset: %v", err)
		}
		n := 1 + (iter % 12) // lengths 1..12, covering the common 4 and 6

		// 1. Generate (ground truth for this iteration).
		passcode, err := internal.GeneratePassword(charSet, n, rng)
		if err != nil {
			t.Fatalf("GeneratePassword: %v", err)
		}

		// 2. Save to disk and 3. load it back.
		path := filepath.Join(dir, "code.enc")
		if err := internal.SavePasscode(path, passcode); err != nil {
			t.Fatalf("SavePasscode: %v", err)
		}
		loaded, err := internal.LoadPasscode(path)
		if err != nil {
			t.Fatalf("LoadPasscode: %v", err)
		}
		if string(loaded) != string(passcode) {
			t.Fatalf("iter=%d charset=%s n=%d: loaded %q != generated %q", iter, cs.name, n, string(loaded), string(passcode))
		}

		// 4. Build keystrokes from the LOADED passcode (as the enter path does)
		//    and 5. replay them. The result must equal the originally generated code.
		steps := internal.AdjustSteps(n+rng.Intn(16), n)
		ks, err := internal.GenerateKeystrokes(loaded, charSet, steps, rng)
		if err != nil {
			t.Fatalf("iter=%d charset=%s n=%d steps=%d: %v", iter, cs.name, n, steps, err)
		}
		if got := replay(ks); string(got) != string(passcode) {
			t.Fatalf("iter=%d charset=%s n=%d: keystroke replay %q != generated %q", iter, cs.name, n, string(got), string(passcode))
		}
	}
}

// TestCryptoRandInRange verifies the production RNG never returns a value
// outside [0,n) — an out-of-range value would panic on a slice index during
// generation.
func TestCryptoRandInRange(t *testing.T) {
	rng := internal.CryptoRand{}

	if got := rng.Intn(0); got != 0 {
		t.Errorf("Intn(0)=%d, want 0", got)
	}
	if got := rng.Intn(1); got != 0 {
		t.Errorf("Intn(1)=%d, want 0", got)
	}

	for _, n := range []int{2, 7, 10, 36, 84} {
		for i := 0; i < 5000; i++ {
			v := rng.Intn(n)
			if v < 0 || v >= n {
				t.Fatalf("Intn(%d)=%d out of range", n, v)
			}
		}
	}
}

// TestCryptoRandCoverage sanity-checks that the RNG actually spans its range
// (catches a stuck/biased implementation that always returns the same value).
func TestCryptoRandCoverage(t *testing.T) {
	rng := internal.CryptoRand{}
	const n = 10
	seen := make(map[int]bool)
	for i := 0; i < 5000; i++ {
		seen[rng.Intn(n)] = true
	}
	if len(seen) != n {
		t.Errorf("Intn(%d) only produced %d distinct values in 5000 draws: %v", n, len(seen), seen)
	}
}
