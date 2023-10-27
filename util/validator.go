package util

import (
	"github.com/go-playground/validator/v10"
	"strings"
)

func UsernameValidation(fl validator.FieldLevel) bool {
	str := fl.Field().String()

	if len(str) < 6 || len(str) > 12 {
		return false
	}

	if strings.HasSuffix(str, "test") {
		return false
	}

	hasRestricted := false
	for _, char := range str {
		if (char < 'a' || 'z' < char) && (char < '0' || '9' < char) {
			hasRestricted = true
			break
		}
	}

	return !hasRestricted
}

func PasswordValidation(fl validator.FieldLevel) bool {
	str := fl.Field().String()

	if len(str) < 8 {
		return false
	}

	hasUppercase := false
	hasLowercase := false
	hasNumber := false
	hasSpecialChar := false
	for _, char := range str {
		if 'A' <= char && char <= 'Z' {
			hasUppercase = true
		} else if 'a' <= char && char <= 'z' {
			hasLowercase = true
		} else if '0' <= char && char <= '9' {
			hasNumber = true
		} else if specialCharacter(char) {
			hasSpecialChar = true
		}
	}

	return hasUppercase && hasLowercase && hasNumber && hasSpecialChar
}

func specialCharacter(char rune) bool {
	specialChars := "!@#$%^&*()-_=+[]{}|;:'\",.<>/?"
	for _, sChar := range specialChars {
		if char == sChar {
			return true
		}
	}
	return false
}
