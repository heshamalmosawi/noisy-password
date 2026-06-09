package main

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"strings"

	"password-fuzzifier/internal"

	"github.com/urfave/cli/v2"
)

func main() {
	// No timestamps on log output.
	log.SetFlags(0)

	app := &cli.App{
		Name:  "passgen",
		Usage: "Generate a passcode you don't memorize, then enter it via noisy keystroke instructions",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "charset", Aliases: []string{"c"}, Value: "numeric", Usage: "Character set: alphabet | numeric | alphanumeric | all"},
			&cli.IntFlag{Name: "length", Aliases: []string{"l"}, Value: 4, Usage: "Passcode length"},
			&cli.IntFlag{Name: "min", Aliases: []string{"m"}, Value: 12, Usage: "Minimum keystroke steps"},
			&cli.IntFlag{Name: "max", Aliases: []string{"x"}, Value: 18, Usage: "Maximum keystroke steps"},
			&cli.StringFlag{Name: "output", Aliases: []string{"o"}, Value: "passcode.enc", Usage: "File to save the generated (base64) passcode to"},
			&cli.BoolFlag{Name: "lowercase", Aliases: []string{"L"}, Usage: "Use lowercase characters only"},
			&cli.BoolFlag{Name: "generate-only", Aliases: []string{"g"}, Usage: "Only generate and save the passcode; skip the entry walkthrough"},
			&cli.StringFlag{Name: "enter", Aliases: []string{"e"}, Usage: "Skip generation; load a saved passcode from this file and walk through entering it"},
		},
		Action: run,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(c *cli.Context) error {
	rng := internal.CryptoRand{}

	charSet, err := internal.GetCharset(c.String("charset"), c.Bool("lowercase"))
	if err != nil {
		return err
	}

	var passcode []rune

	if enterFile := c.String("enter"); enterFile != "" {
		// Enter-only: load a previously saved passcode and re-enter it with fresh noise.
		passcode, err = loadPasscode(enterFile)
		if err != nil {
			return err
		}
	} else {
		// Generate a new passcode and save it.
		length := c.Int("length")
		if length < 1 {
			return fmt.Errorf("length must be at least 1")
		}
		passcode, err = internal.GeneratePassword(charSet, length, rng)
		if err != nil {
			return err
		}

		outputFile := c.String("output")
		encoded := base64.StdEncoding.EncodeToString([]byte(string(passcode)))
		if err := os.WriteFile(outputFile, []byte(encoded), 0o600); err != nil {
			return fmt.Errorf("writing %q: %w", outputFile, err)
		}
		fmt.Printf("Passcode generated and saved (base64) to %q.\n", outputFile)

		if c.Bool("generate-only") {
			return nil
		}
		fmt.Printf("Re-enter it any time with:  passgen --enter %s\n\n", outputFile)
	}

	// Build a fresh noisy keystroke sequence and walk through it.
	steps := chooseSteps(c.Int("min"), c.Int("max"), len(passcode), rng)
	keystrokes, err := internal.GenerateKeystrokes(passcode, charSet, steps, rng)
	if err != nil {
		return err
	}
	return walkthrough(keystrokes)
}

// loadPasscode reads and base64-decodes a saved passcode file.
func loadPasscode(file string) ([]rune, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("reading %q: %w", file, err)
	}
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(data)))
	if err != nil {
		return nil, fmt.Errorf("decoding %q: %w", file, err)
	}
	passcode := []rune(string(decoded))
	if len(passcode) == 0 {
		return nil, fmt.Errorf("file %q contains no passcode", file)
	}
	return passcode, nil
}

// chooseSteps picks a keystroke budget in [min,max], then adjusts it to a value
// valid for a passcode of length n.
func chooseSteps(min, max, n int, r internal.Rand) int {
	if min < 1 {
		min = 1
	}
	if max < min {
		min, max = max, min
	}
	steps := min + r.Intn(max-min+1)
	return internal.AdjustSteps(steps, n)
}

// walkthrough prints one keystroke instruction at a time, waiting for Enter
// after each and clearing it before showing the next. The resulting passcode
// is never displayed, and real vs. noise presses look identical.
func walkthrough(steps []internal.Step) error {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Follow these keystrokes on your device. Press Enter after each one.")
	fmt.Println()

	for i, s := range steps {
		instruction := "Press " + string(s.Char)
		if s.Action == internal.Backspace {
			instruction = "Backspace"
		}
		fmt.Printf("[%d/%d] %s", i+1, len(steps), instruction)

		if _, err := reader.ReadString('\n'); err != nil {
			// EOF or read error: stop cleanly without revealing anything.
			fmt.Println()
			return nil
		}
		deleteLine() // erase the instruction line so the next one reuses the space
	}

	fmt.Println("Done. Your passcode has been entered.")
	return nil
}

// deleteLine moves the cursor up one line and clears it.
func deleteLine() {
	fmt.Print("\033[1A") // cursor up one line
	fmt.Print("\033[2K") // clear the entire line
	fmt.Print("\033[G")  // cursor to column 1
}
