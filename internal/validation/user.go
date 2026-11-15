package validation

import (
	"net/mail"
	"regexp"
	"strings"
	"unicode/utf8"
)

var (
	// Supports Unicode letters/marks (including Thai), separated by spaces,
	// apostrophes, or hyphens.
	namePattern  = regexp.MustCompile(`^[\p{L}][\p{L}\p{M}]*(?:[ '\-][\p{L}][\p{L}\p{M}]*)*$`)
	emailPattern = regexp.MustCompile(
		`^[A-Za-z0-9_%+-]+(?:\.[A-Za-z0-9_%+-]+)*@(?:[A-Za-z0-9](?:[A-Za-z0-9-]*[A-Za-z0-9])?\.)+[A-Za-z]{2,}$`,
	)
)

func Name(s string) bool {
	s = strings.TrimSpace(s)
	n := utf8.RuneCountInString(s)
	return utf8.ValidString(s) && n >= 2 && n <= 100 && namePattern.MatchString(s)
}

func Email(s string) bool {
	s = strings.TrimSpace(s)

	if len(s) > 254 || !emailPattern.MatchString(s) {
		return false
	}

	parts := strings.Split(s, "@")
	if len(parts) != 2 {
		return false
	}

	localPart := parts[0]
	domain := parts[1]

	if len(localPart) > 64 || len(domain) > 253 {
		return false
	}

	for _, label := range strings.Split(domain, ".") {
		if len(label) > 63 {
			return false
		}
	}

	parsed, err := mail.ParseAddress(s)
	return err == nil && parsed.Address == s
}

// Password is intentionally not trimmed: spaces may be part of a password.
// bcrypt accepts at most 72 bytes.
func PasswordRegister(s string) bool { return len(s) >= 8 && len(s) <= 72 }
func PasswordLogin(s string) bool    { return len(s) >= 1 && len(s) <= 72 }
