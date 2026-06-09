package internal

import "fmt"

// Action is a single keystroke primitive available on a passcode field.
type Action int

const (
	// AddCorrect appends the next correct passcode character.
	AddCorrect Action = iota
	// AddNoise appends a throwaway character that will later be deleted.
	AddNoise
	// Backspace removes the last character in the buffer.
	Backspace
)

// Step is one keystroke instruction. Char is set for AddCorrect/AddNoise and
// unused for Backspace. AddCorrect and AddNoise are deliberately indistinguishable
// to whoever follows the steps — both are shown as "Press <Char>".
type Step struct {
	Action Action
	Char   rune
}

// AdjustSteps returns a valid keystroke budget for a passcode of length n:
//   - at least n,
//   - the same parity as n (each noise excursion costs two keystrokes: an add
//     plus the backspace that removes it),
//   - clamped to n when n <= 1, since the auto-submit cap leaves no room for
//     noise in that case.
func AdjustSteps(requested, n int) int {
	if n <= 1 {
		return n
	}
	s := requested
	if s < n {
		s = n
	}
	if (s-n)%2 != 0 {
		s++
	}
	return s
}

// GenerateKeystrokes builds a fresh ADD/BACKSPACE keystroke sequence that, when
// replayed on an empty buffer, yields exactly passcode. Real keystrokes are
// interleaved with noise so that following the steps does not reveal which
// presses matter.
//
// It targets an auto-submitting field (e.g. an iPhone passcode), so the buffer
// length never reaches len(passcode) until the single final keystroke.
//
// The construction maintains a committed-prefix invariant: the first k buffer
// characters are always the correct prefix passcode[:k] and are never deleted;
// g throwaway characters may sit on top. Backspaces only ever remove garbage,
// which makes the final result provably equal to passcode.
func GenerateKeystrokes(passcode []rune, charset string, steps int, r Rand) ([]Step, error) {
	n := len(passcode)
	if n == 0 {
		return nil, fmt.Errorf("passcode cannot be empty")
	}
	cs := []rune(charset)
	if len(cs) == 0 {
		return nil, fmt.Errorf("charset cannot be empty")
	}
	if steps < n || (steps-n)%2 != 0 {
		return nil, fmt.Errorf("steps (%d) must be >= passcode length (%d) and of equal parity", steps, n)
	}
	if n == 1 && steps != 1 {
		return nil, fmt.Errorf("a 1-character passcode allows no noise; steps must be 1")
	}

	out := make([]Step, 0, steps)
	k := 0 // committed correct prefix length (buffer[:k] == passcode[:k])
	g := 0 // garbage characters currently on top

	for i := 0; i < steps; i++ {
		remaining := steps - i
		// minNeeded = backspaces to clear current garbage + remaining correct adds.
		minNeeded := g + (n - k)

		if remaining == minNeeded {
			// Forced finish: drain all garbage, then add the remaining correct chars.
			if g > 0 {
				out = append(out, Step{Action: Backspace})
				g--
			} else {
				out = append(out, Step{Action: AddCorrect, Char: passcode[k]})
				k++
			}
			continue
		}

		// Free choice. The parity invariant guarantees remaining-minNeeded is
		// even and >= 2 here, so adding noise (which costs 2 to undo) stays in budget.
		var opts []Action
		// Advance a correct char only while doing so keeps noise room (k+1 <= n-2);
		// the last two correct chars are placed in the forced finish, since the
		// auto-submit cap forbids noise once the buffer is one short of full.
		if g == 0 && k <= n-3 {
			opts = append(opts, AddCorrect)
		}
		if g > 0 {
			opts = append(opts, Backspace)
		}
		// Noise is only allowed below the cap: k+g must stay <= n-2 so the next
		// add cannot reach the full length and submit early.
		if k+g <= n-2 {
			opts = append(opts, AddNoise)
		}

		switch opts[r.Intn(len(opts))] {
		case AddCorrect:
			out = append(out, Step{Action: AddCorrect, Char: passcode[k]})
			k++
		case AddNoise:
			out = append(out, Step{Action: AddNoise, Char: cs[r.Intn(len(cs))]})
			g++
		case Backspace:
			out = append(out, Step{Action: Backspace})
			g--
		}
	}

	if k != n || g != 0 {
		return nil, fmt.Errorf("internal error: ended at k=%d g=%d for passcode length %d", k, g, n)
	}
	if got := replay(out); string(got) != string(passcode) {
		return nil, fmt.Errorf("internal error: replay %q != passcode %q", string(got), string(passcode))
	}
	return out, nil
}

// replay returns the buffer produced by following steps on an empty buffer.
func replay(steps []Step) []rune {
	buf := make([]rune, 0, len(steps))
	for _, s := range steps {
		switch s.Action {
		case AddCorrect, AddNoise:
			buf = append(buf, s.Char)
		case Backspace:
			if len(buf) > 0 {
				buf = buf[:len(buf)-1]
			}
		}
	}
	return buf
}
