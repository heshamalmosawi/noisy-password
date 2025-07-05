package internal

import "errors"

var charsets = map[string]map[bool]string{
	"alphabet": {
		true:  "abcdefghijklmnopqrstuvwxyz",
		false: "ABCDEFGHIJKLMNOPQRSTUVWXYZ",
	},
	"numeric": {
		true:  "0123456789",
		false: "0123456789",
	},
	"alphanumeric": {
		true:  "abcdefghijklmnopqrstuvwxyz0123456789",
		false: "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789",
	},
	"all": {
		true:  "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789",
		false: "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*()-_=+[]{}|;:,.<>?/",
	},
}

// GetCharset returns a character set based on the provided charset name and lowercase option.
// It supports "alphabet", "numeric", "alphanumeric", and "all" character sets.
func GetCharset(charset string, lowercase bool) (string, error) {
	if options, ok := charsets[charset]; ok {
		return options[lowercase], nil
	}
	return "", errors.New("invalid charset")
}
