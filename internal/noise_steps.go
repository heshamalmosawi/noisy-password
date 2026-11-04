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

// ChooseSteps picks a keystroke budget within [min,max] that is valid for a
// passcode of length n: at least n, and sharing n's parity (each noise excursion
// costs two keystrokes). It selects uniformly among the valid values inside the
// window, so the result never exceeds max. The window is only overridden when it
// genuinely cannot be honored — when max is below n, or when min==max lands on
// the wrong parity (in which case the nearest valid value at or below max is used).
func ChooseSteps(min, max, n int, r Rand) int {
	// A 0/1-length passcode admits no noise: the only valid budget is n itself.
	if n <= 1 {
		return n
	}
	if max < min {
		min, max = max, min
	}

	// Snap the window inward to values with the correct parity and >= n.
	lo := min
	if lo < n {
		lo = n
	}
	if (lo-n)%2 != 0 {
		lo++ // up to the next value with n's parity
	}
	hi := max
	if (hi-n)%2 != 0 {
		hi-- // down to the previous value with n's parity
	}

	switch {
	case hi < n:
		// max is below the passcode length: impossible to honor, use the minimum.
		return n
	case lo > hi:
		// No valid value lies in the window (e.g. min==max with the wrong parity);
		// use the largest valid value that does not exceed max.
		return hi
	default:
		count := (hi-lo)/2 + 1
		return lo + 2*r.Intn(count)
	}
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
// throwaway characters may sit on top. Backspaces only ever remove garbage,
// which makes the final result provably equal to passcode.
//
// Noise is planned up front rather than chosen per keystroke. A gap is the point
// just before passcode[k] is committed, and it can hold (n-1)-k garbage characters
// before the field would submit. The noise budget is shared randomly across the
// gaps, favouring those with more room, so short passcodes don't spend it all
// adding and deleting one character in the last slot that accepts noise.
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

	noise := allocateNoise((steps-n)/2, n, r)

	out := make([]Step, 0, steps)
	for k, c := range passcode {
		if k < len(noise) {
			out = appendNoiseWalk(out, noise[k], n-1-k, cs, r)
		}
		out = append(out, Step{Action: AddCorrect, Char: c})
	}

	if len(out) != steps {
		return nil, fmt.Errorf("internal error: built %d keystrokes, want %d", len(out), steps)
	}
	if got := replay(out); string(got) != string(passcode) {
		return nil, fmt.Errorf("internal error: replay %q != passcode %q", string(got), string(passcode))
	}
	return out, nil
}

// allocateNoise randomly shares count noise characters across the gaps of a
// passcode of length n, returning how many each gap gets. Only gaps 0..n-2 can
// hold noise. Each character picks a gap with probability proportional to that
// gap's headroom, so counts vary per run and the roomier gaps receive more.
func allocateNoise(count, n int, r Rand) []int {
	if n < 2 {
		return nil
	}
	alloc := make([]int, n-1)
	totalHeadroom := n * (n - 1) / 2 // (n-1) + (n-2) + ... + 1
	for i := 0; i < count; i++ {
		pick := r.Intn(totalHeadroom)
		for k := range alloc {
			headroom := n - 1 - k
			if pick < headroom {
				alloc[k]++
				break
			}
			pick -= headroom
		}
	}
	return alloc
}

// appendNoiseWalk appends a random excursion of count noise characters, each
// later deleted, with at most headroom of them in the buffer at once. It starts
// and ends with no garbage on top, leaving the committed prefix untouched.
func appendNoiseWalk(out []Step, count, headroom int, cs []rune, r Rand) []Step {
	depth := 0
	for count > 0 || depth > 0 {
		canAdd := count > 0 && depth < headroom
		canDelete := depth > 0
		if canAdd && (!canDelete || r.Intn(2) == 0) {
			out = append(out, Step{Action: AddNoise, Char: cs[r.Intn(len(cs))]})
			count--
			depth++
		} else {
			out = append(out, Step{Action: Backspace})
			depth--
		}
	}
	return out
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
