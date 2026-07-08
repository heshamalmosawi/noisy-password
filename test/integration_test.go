package test

import (
	"math/rand"
	"password-fuzzifier/internal"
	"strings"
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

				// Every emitted character must come from the charset, and the
				// AddCorrect steps must spell the passcode in order.
				var correct []rune
				for _, s := range ks {
					switch s.Action {
					case internal.AddCorrect:
						correct = append(correct, s.Char)
						if !strings.ContainsRune(charSet, s.Char) {
							t.Fatalf("seed=%d charset=%s n=%d: AddCorrect char %q not in charset", seed, cs.name, n, s.Char)
						}
					case internal.AddNoise:
						if !strings.ContainsRune(charSet, s.Char) {
							t.Fatalf("seed=%d charset=%s n=%d: AddNoise char %q not in charset", seed, cs.name, n, s.Char)
						}
					}
				}
				if string(correct) != string(passcode) {
					t.Fatalf("seed=%d charset=%s n=%d: AddCorrect sequence %q != passcode %q", seed, cs.name, n, string(correct), string(passcode))
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

// TestChooseStepsRespectsWindow checks that the chosen budget stays within
// [min,max] whenever a valid value exists there, and is always valid for the
// generator (>= n and matching parity).
func TestChooseStepsRespectsWindow(t *testing.T) {
	r := rand.New(rand.NewSource(7))

	// Specific cases, including the reported regression (min==max==12, n==5).
	cases := []struct {
		min, max, n int
		wantMin     int // expected lower bound on result (inclusive)
		wantMax     int // expected upper bound on result (inclusive)
	}{
		{min: 12, max: 12, n: 5, wantMin: 11, wantMax: 12}, // wrong-parity point -> 11, never 13 (> max)
		{min: 12, max: 12, n: 4, wantMin: 12, wantMax: 12}, // valid as-is
		{min: 12, max: 18, n: 4, wantMin: 12, wantMax: 18},
		{min: 12, max: 18, n: 5, wantMin: 13, wantMax: 17},
		{min: 18, max: 12, n: 4, wantMin: 12, wantMax: 18}, // inverted window is swapped
		{min: 2, max: 3, n: 6, wantMin: 6, wantMax: 6},     // max below n -> forced to n
		{min: 1, max: 5, n: 1, wantMin: 1, wantMax: 1},     // length 1 -> exactly 1
	}
	for _, tc := range cases {
		for i := 0; i < 200; i++ {
			got := internal.ChooseSteps(tc.min, tc.max, tc.n, r)
			if got < tc.wantMin || got > tc.wantMax {
				t.Fatalf("ChooseSteps(%d,%d,%d)=%d, want in [%d,%d]", tc.min, tc.max, tc.n, got, tc.wantMin, tc.wantMax)
			}
			if got < tc.n {
				t.Fatalf("ChooseSteps(%d,%d,%d)=%d is below n", tc.min, tc.max, tc.n, got)
			}
			if tc.n > 1 && (got-tc.n)%2 != 0 {
				t.Fatalf("ChooseSteps(%d,%d,%d)=%d has wrong parity for n=%d", tc.min, tc.max, tc.n, got, tc.n)
			}
		}
	}

	// Property sweep: result is always generator-valid and never exceeds max
	// unless the window is infeasible (max < n).
	for i := 0; i < 5000; i++ {
		n := 1 + r.Intn(20)
		min := 1 + r.Intn(30)
		max := 1 + r.Intn(30)
		got := internal.ChooseSteps(min, max, n, r)

		if got < n || (n > 1 && (got-n)%2 != 0) {
			t.Fatalf("ChooseSteps(%d,%d,%d)=%d is not generator-valid", min, max, n, got)
		}
		hi := max
		if min > max {
			hi = min // window is swapped internally
		}
		if hi >= n && got > hi {
			t.Fatalf("ChooseSteps(%d,%d,%d)=%d exceeds feasible max %d", min, max, n, got, hi)
		}
		// The chosen budget must actually produce a valid sequence.
		cs, _ := internal.GetCharset("alphanumeric", false)
		pass, _ := internal.GeneratePassword(cs, n, r)
		if _, err := internal.GenerateKeystrokes(pass, cs, got, r); err != nil {
			t.Fatalf("ChooseSteps(%d,%d,%d)=%d rejected by generator: %v", min, max, n, got, err)
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
