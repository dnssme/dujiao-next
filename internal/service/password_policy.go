package service

import (
	"unicode"

	"github.com/dujiao-next/internal/config"
)

type passwordPolicyError struct {
	key  string
	args []interface{}
}

func (e passwordPolicyError) Error() string {
	return e.key
}

func (e passwordPolicyError) Is(target error) bool {
	return target == ErrWeakPassword
}

func (e passwordPolicyError) Key() string {
	return e.key
}

func (e passwordPolicyError) Args() []interface{} {
	return e.args
}

// BcryptMaxPasswordBytes is the maximum number of bytes bcrypt will hash.
// Passwords longer than this are silently truncated, so we reject them to
// avoid giving users a false sense of security (CIS 5.2).
const BcryptMaxPasswordBytes = 72

func validatePassword(policy config.PasswordPolicyConfig, password string) error {
	if len(password) > BcryptMaxPasswordBytes {
		return passwordPolicyError{key: "error.password_max_length", args: []interface{}{BcryptMaxPasswordBytes}}
	}

	if policy.MinLength <= 0 &&
		!policy.RequireUpper &&
		!policy.RequireLower &&
		!policy.RequireNumber &&
		!policy.RequireSpecial {
		return nil
	}

	if policy.MinLength > 0 {
		if len([]rune(password)) < policy.MinLength {
			return passwordPolicyError{key: "error.password_min_length", args: []interface{}{policy.MinLength}}
		}
	}

	var hasUpper, hasLower, hasNumber, hasSpecial bool
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasNumber = true
		default:
			hasSpecial = true
		}
	}

	if policy.RequireUpper && !hasUpper {
		return passwordPolicyError{key: "error.password_require_upper"}
	}
	if policy.RequireLower && !hasLower {
		return passwordPolicyError{key: "error.password_require_lower"}
	}
	if policy.RequireNumber && !hasNumber {
		return passwordPolicyError{key: "error.password_require_number"}
	}
	if policy.RequireSpecial && !hasSpecial {
		return passwordPolicyError{key: "error.password_require_special"}
	}

	return nil
}
