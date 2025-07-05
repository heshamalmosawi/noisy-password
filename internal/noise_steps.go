package internal

import (
	"fmt"
	"math/rand"
	"strings"
)

// GenerateDynamicSequence emulates the JS style:
// - Guarantees you add all correct digits in order
// - Adds random noise with wrong digits and deletes
// - Final result always equals finalPasscode
func GenerateDynamicSequence(finalPasscode []rune, charSet string, steps int, r *rand.Rand) ([]string, error) {
	if len(finalPasscode) == 0 {
		return nil, fmt.Errorf("passcode cannot be empty")
	} else if steps < len(finalPasscode) {
		return nil, fmt.Errorf("steps must be at least the passcode length")
	}

	var output []string
	display := make([]rune, 0)
	correctDigitIndex := 0

	for i := 0; i < steps; i++ {
		remainingSteps := steps - i
		remainingCorrect := len(finalPasscode) - correctDigitIndex

		// If we must add the remaining correct digits to finish on time, do so
		if remainingCorrect >= remainingSteps {
			// Must add correct digit now
			if len(display) == correctDigitIndex {
				c := finalPasscode[correctDigitIndex]
				display = append(display, c)
				correctDigitIndex++
				output = append(output, "Add '"+string(c)+"' → "+stringifyAndHide(display))
				continue
			} else {
				// If display length != correctDigitIndex, must delete to fix position
				removed := display[len(display)-1]
				display = display[:len(display)-1]
				output = append(output, "Delete '"+string(removed)+"' → "+stringifyAndHide(display))
				continue
			}
		}

		// Can add correct digit only if display length == correctDigitIndex
		canAdd := len(display) < len(finalPasscode) && correctDigitIndex < len(finalPasscode)
		// Can delete only if there's something to delete AND it's not a correct digit in correct position
		canDelete := len(display) > 0 && !isCorrectDigitInCorrectPosition(display, finalPasscode)

		// Decide: add or delete?
		action := ""
		if canAdd && canDelete {
			if r.Intn(2) == 0 {
				action = "add"
			} else {
				action = "delete"
			}
		} else if canAdd {
			action = "add"
		} else {
			action = "delete"
		}

		switch action {
		case "add":
			// Only add correct digit if display length == correctDigitIndex
			if remainingCorrect > 0 && len(display) == correctDigitIndex && r.Intn(100) < 50 {
				c := finalPasscode[correctDigitIndex]
				display = append(display, c)
				correctDigitIndex++
				output = append(output, "Add '"+string(c)+"' → "+stringifyAndHide(display))
			} else {
				// Add wrong digit
				var c rune
				if len(charSet) == 0 {
					c = 'X'
				} else {
					c = rune(charSet[r.Intn(len(charSet))])
				}
				display = append(display, c)
				output = append(output, "Add '"+string(c)+"' → "+stringifyAndHide(display))
			}
		case "delete":
			removed := display[len(display)-1]
			display = display[:len(display)-1]
			output = append(output, "Delete '"+string(removed)+"' → "+stringifyAndHide(display))
		}
	}

	// Ensure all correct digits are added at the end
	for correctDigitIndex < len(finalPasscode) {
		c := finalPasscode[correctDigitIndex]
		display = append(display, c)
		correctDigitIndex++
		output = append(output, "Add '"+string(c)+"' → "+stringifyAndHide(display))
	}

	output = append(output, "Final display → "+stringifyAndHide(display))
	return output, nil
}

func stringifyAndHide(runes []rune) string {
	if len(runes) == 0 {
		return "[]"
	}
	var sb strings.Builder
	for i, r := range runes {
		if i == len(runes)-1 {
			sb.WriteRune(r) // Keep the last character as is
		} else {
			sb.WriteRune('X') // Replace all other characters with 'X'
		}
	}
	return sb.String()
}

// isCorrectDigitInCorrectPosition checks if the last character in display
// is a correct digit in its correct position
func isCorrectDigitInCorrectPosition(display []rune, finalPasscode []rune) bool {
	if len(display) == 0 {
		return false
	}

	lastIndex := len(display) - 1
	// If we're beyond the passcode length, it can't be correct
	if lastIndex >= len(finalPasscode) {
		return false
	}

	// Check if the last character matches the expected character at that position
	return display[lastIndex] == finalPasscode[lastIndex]
}
