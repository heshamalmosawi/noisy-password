package internal

import "errors"

// charsets maps a charset name to its lowercase (true) / uppercase-or-full
// (false) variant. "numeric" is identical for both keys.
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
		true:  "abcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*()-_=+[]{}|;:,.<>?/",
		false: "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*()-_=+[]{}|;:,.<>?/",
	},
}

// GetCharset returns the character set for the given name and lowercase option.
// Supported names: "alphabet", "numeric", "alphanumeric", "all".
func GetCharset(charset string, lowercase bool) (string, error) {
	if options, ok := charsets[charset]; ok {
		return options[lowercase], nil
	}
	return "", errors.New("invalid charset")
}
