package test

import (
	"math/rand"
	"password-fuzzifier/internal"
	"testing"
)

// replay returns the buffer produced by following a keystroke sequence on an
// empty buffer — i.e. what the device would end up holding.
func replay(steps []internal.Step) []rune {
	buf := make([]rune, 0, len(steps))
	for _, s := range steps {
		switch s.Action {
		case internal.Backspace:
			if len(buf) > 0 {
				buf = buf[:len(buf)-1]
			}
		default: // AddCorrect / AddNoise
			buf = append(buf, s.Char)
		}
	}
	return buf
}

// TestKeystrokesReconstructAndRespectCap exhaustively checks the core
// guarantees across many seeds, lengths and charsets:
//   - the keystroke count equals the requested budget,
//   - the auto-submit cap is never violated (buffer < N until the final keystroke),
//   - replaying the keystrokes reproduces the exact passcode.
func TestKeystrokesReconstructAndRespectCap(t *testing.T) {
	charsets := []struct {
		name      string
		lowercase bool
	}{
		{"numeric", false},
		{"alphabet", true},
		{"alphanumeric", false},
		{"all", false},
	}

	for seed := int64(0); seed < 300; seed++ {
		r := rand.New(rand.NewSource(seed))
		for _, cs := range charsets {
			charSet, err := internal.GetCharset(cs.name, cs.lowercase)
			if err != nil {
				t.Fatalf("GetCharset(%q): %v", cs.name, err)
			}
			for n := 1; n <= 20; n++ {
				passcode, err := internal.GeneratePassword(charSet, n, r)
				if err != nil {
					t.Fatalf("GeneratePassword: %v", err)
				}

				steps := internal.AdjustSteps(n+r.Intn(16), n)
				ks, err := internal.GenerateKeystrokes(passcode, charSet, steps, r)
				if err != nil {
					t.Fatalf("seed=%d charset=%s n=%d steps=%d: %v", seed, cs.name, n, steps, err)
				}

				if len(ks) != steps {
					t.Fatalf("seed=%d n=%d: got %d keystrokes, want %d", seed, n, len(ks), steps)
				}

				// Walk the buffer length, enforcing the auto-submit cap.
				buf := 0
				for i, s := range ks {
					if s.Action == internal.Backspace {
						if buf > 0 {
							buf--
						}
					} else {
						buf++
					}
					if i < len(ks)-1 && buf >= n {
						t.Fatalf("seed=%d charset=%s n=%d: cap violated at step %d (buffer=%d)", seed, cs.name, n, i, buf)
					}
					if i == len(ks)-1 && buf != n {
						t.Fatalf("seed=%d charset=%s n=%d: final buffer=%d, want %d", seed, cs.name, n, buf, n)
					}
				}

				if got := replay(ks); string(got) != string(passcode) {
					t.Fatalf("seed=%d charset=%s n=%d: replay %q != passcode %q", seed, cs.name, n, string(got), string(passcode))
				}
			}
		}
	}
}

func TestAdjustSteps(t *testing.T) {
	tests := []struct {
		requested, n, want int
	}{
		{0, 1, 1},   // n<=1 clamps to n
		{100, 1, 1}, // no noise possible for length 1
		{0, 4, 4},   // floor to n
		{5, 4, 6},   // bump parity (5-4 odd -> 6)
		{6, 4, 6},   // already valid
		{3, 6, 6},   // floor to n
		{13, 6, 14}, // bump parity
	}
	for _, tt := range tests {
		if got := internal.AdjustSteps(tt.requested, tt.n); got != tt.want {
			t.Errorf("AdjustSteps(%d,%d)=%d, want %d", tt.requested, tt.n, got, tt.want)
		}
		got := internal.AdjustSteps(tt.requested, tt.n)
		if tt.n > 1 && (got-tt.n)%2 != 0 {
			t.Errorf("AdjustSteps(%d,%d)=%d has wrong parity for n=%d", tt.requested, tt.n, got, tt.n)
		}
	}
}

func TestGenerateKeystrokesRejectsBadInput(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	cs, _ := internal.GetCharset("numeric", false)

	if _, err := internal.GenerateKeystrokes([]rune("123"), cs, 2, r); err == nil {
		t.Error("expected error when steps < passcode length")
	}
	if _, err := internal.GenerateKeystrokes([]rune("123"), cs, 4, r); err == nil {
		t.Error("expected error when steps parity differs from passcode length")
	}
	if _, err := internal.GenerateKeystrokes([]rune{}, cs, 0, r); err == nil {
		t.Error("expected error for empty passcode")
	}
	if _, err := internal.GenerateKeystrokes([]rune("1"), cs, 3, r); err == nil {
		t.Error("expected error for 1-char passcode with steps != 1")
	}
}

// TestFreshNoisePerRun confirms the same passcode yields different keystroke
// sequences on different runs (the property that resists memorization).
func TestFreshNoisePerRun(t *testing.T) {
	cs, _ := internal.GetCharset("numeric", false)
	passcode := []rune("4827")

	a, err := internal.GenerateKeystrokes(passcode, cs, 16, rand.New(rand.NewSource(1)))
	if err != nil {
		t.Fatal(err)
	}
	b, err := internal.GenerateKeystrokes(passcode, cs, 16, rand.New(rand.NewSource(2)))
	if err != nil {
		t.Fatal(err)
	}

	if string(replay(a)) != string(passcode) || string(replay(b)) != string(passcode) {
		t.Fatal("both sequences must reconstruct the passcode")
	}
	same := len(a) == len(b)
	if same {
		for i := range a {
			if a[i] != b[i] {
				same = false
				break
			}
		}
	}
	if same {
		t.Error("expected different keystroke sequences for different seeds")
	}
}
