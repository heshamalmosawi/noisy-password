package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"math/rand"
	"os"
	"password-fuzzifier/internal"
	"time"

	"github.com/urfave/cli/v2"
)

func main() {
	// setting log flags to 0 to disable timestamps and other info
	log.SetFlags(0)

	app := &cli.App{
		Name:  "passgen",
		Usage: "Generate a random password with added noise and output encoded",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "charset",
				Aliases: []string{"c"},
				Value:   "alphanumeric",
				Usage:   "Character set: alphabet | numeric | alphanumeric | all",
			},
			&cli.IntFlag{
				Name:    "length",
				Aliases: []string{"l"},
				Value:   4,
				Usage:   "Final password length",
			},
			&cli.IntFlag{
				Name:    "min",
				Aliases: []string{"m"},
				Value:   6,
				Usage:   "Minimum random steps",
			},
			&cli.IntFlag{
				Name:    "max",
				Aliases: []string{"x"},
				Value:   6,
				Usage:   "Maximum random steps",
			},
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Value:   "password.enc",
				Usage:   "Output file name",
			},
			&cli.BoolFlag{
				Name:    "lowercase",
				Aliases: []string{"L"},
				Usage:   "Generate password with lowercase characters only",
			},
		},
		Action: func(c *cli.Context) error {
			// parsing flags
			charsetName := c.String("charset")
			plength := c.Int("length")
			minSteps := c.Int("min")
			maxSteps := c.Int("max")
			outputFile := c.String("output")
			lowercase := c.Bool("lowercase")

			if minSteps <= 0 {
				return fmt.Errorf("min steps must be greater than 0")
			} else if minSteps < plength {
				return fmt.Errorf("min steps cannot be less than password length")
			} else if maxSteps < minSteps {
				// simply reverse minSteps and maxSteps if maxSteps is lower than the minimum
				minSteps, maxSteps = maxSteps, minSteps
			}

			charSet, err := internal.GetCharset(charsetName, lowercase)
			if err != nil {
				return err
			}

			password, err := internal.GeneratePassword(charSet, plength)
			if err != nil {
				return err
			}

			r := rand.New(rand.NewSource(time.Now().UnixNano()))

			// to calculate random amount of steps we will use,
			// we generate a random number between the difference of minSteps and maxSteps
			// and add the minimum steps to it to ensure we always have at least minSteps
			// for example, if minSteps is 2 and maxSteps is 8,
			// we can genarate a random number between 0 and 6 (8 - 2)
			// and add minSteps (2) to it, resulting in a random number between 2 and 9.
			steps := r.Intn(maxSteps-minSteps+1) + minSteps
			if steps <= plength+1 {
				steps = plength + 1 // ensure we have enough steps to add all characters
			}

			chaosSteps, err := internal.GenerateDynamicSequence([]rune(password), charSet, steps, r)
			if err != nil {
				return err
			}

			fmt.Println("Enter any char to continue...")
			for _, step := range chaosSteps {
				fmt.Println(step)
				var input string
				fmt.Scanln(&input)

				deleteLine()
				deleteLine()

			}
			deleteLine()

			fmt.Println("your final password is", string(password), "TODO: Delete this after dev & testing")
			encoded := base64.StdEncoding.EncodeToString([]byte(string(password)))

			err = os.WriteFile(outputFile, []byte(encoded), 0644)
			if err != nil {
				return fmt.Errorf("error writing file: %v", err)
			}

			fmt.Printf("password saved in Base64 to '%s'\n", outputFile)
			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func deleteLine() {
	// Move cursor up one line and clear the line
	fmt.Print("\033[1A") // Move cursor up one line
	fmt.Print("\033[K")  // Clear current line
	fmt.Print("\033[G")  // Move cursor to beginning of the line
}
