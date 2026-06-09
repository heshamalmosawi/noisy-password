package internal

import (
	"crypto/rand"
	"errors"
	"math/big"
)

// Rand is the minimal randomness source the generators depend on.
// *math/rand.Rand satisfies it (handy for deterministic tests), as does CryptoRand.
type Rand interface {
	Intn(n int) int
}

// CryptoRand is a cryptographically secure Rand backed by crypto/rand.
type CryptoRand struct{}

// Intn returns a uniform random integer in [0, n). It panics only if the
// system CSPRNG is unavailable, which indicates a broken environment.
func (CryptoRand) Intn(n int) int {
	if n <= 0 {
		return 0
	}
	v, err := rand.Int(rand.Reader, big.NewInt(int64(n)))
	if err != nil {
		panic("crypto/rand unavailable: " + err.Error())
	}
	return int(v.Int64())
}

// GeneratePassword builds a random passcode of the given length drawn uniformly
// from charset.
func GeneratePassword(charset string, length int, r Rand) ([]rune, error) {
	if charset == "" {
		return nil, errors.New("charset cannot be empty")
	}
	if length < 0 {
		return nil, errors.New("length cannot be negative")
	}
	runes := []rune(charset)
	password := make([]rune, length)
	for i := range password {
		password[i] = runes[r.Intn(len(runes))]
	}
	return password, nil
}
