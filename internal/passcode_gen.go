package internal

import (
	"errors"
	"math/rand"
)

// This function generates a random password based on the provided character set and length.
// It returns an error if the charset is empty or if the length is negative.
func GeneratePassword(charset string, length int) ([]rune, error) {
	if charset == "" {
		return nil, errors.New("charset cannot be empty")
	}
	if length < 0 {
		return nil, errors.New("length must be greater than 0")
	}
	password := make([]rune, length)
	for i := range password {
		password[i] = rune(charset[rand.Intn(len(charset))])
	}

	return password, nil
}
