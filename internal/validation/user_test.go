package validation

import (
	"strings"
	"testing"
)

func TestName(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want bool
	}{
		{"valid english", "John Doe", true},
		{"valid thai with marks", "สมชาย ใจดี", true},
		{"valid hyphen and apostrophe", "Jean-Luc O'Connor", true},
		{"valid with outer spaces", "  Alice  ", true},
		{"min length 2 runes", "กข", true},
		{"max length 100 runes", strings.Repeat("ก", 100), true},
		{"too short 1 rune", "A", false},
		{"too long 101 runes", strings.Repeat("a", 101), false},
		{"empty or spaces only", "   ", false},
		{"consecutive spaces", "John  Doe", false},
		{"contains digits", "John123", false},
		{"invalid utf8", string([]byte{0xff, 0xfe, 0xfd}), false},
	}
	for _, tc := range cases {
		if got := Name(tc.in); got != tc.want {
			t.Errorf("%s: Name(%q) = %v, want %v", tc.name, tc.in, got, tc.want)
		}
	}
}

func TestEmail(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want bool
	}{
		{"valid standard", "user@example.com", true},
		{"valid subdomain and plus", "first.last+tag@sub.example.co.th", true},
		{"valid trimmed spaces", "  user@example.com  ", true},
		{"max local part 64 chars", strings.Repeat("a", 64) + "@example.com", true},
		{"local part exceeds 64 chars", strings.Repeat("a", 65) + "@example.com", false},
		{"domain label exceeds 63 chars", "u@" + strings.Repeat("a", 64) + ".com", false},
		{"total length exceeds 254 chars", "u@" + strings.Repeat("a.", 125) + "com", false},
		{"missing @", "userexample.com", false},
		{"multiple @", "a@b@example.com", false},
		{"leading dot in local", ".user@example.com", false},
		{"consecutive dots", "user..name@example.com", false},
		{"leading hyphen in domain", "user@-example.com", false},
		{"single char TLD", "user@example.c", false},
	}
	for _, tc := range cases {
		if got := Email(tc.in); got != tc.want {
			t.Errorf("%s: Email(%q) = %v, want %v", tc.name, tc.in, got, tc.want)
		}
	}
}

func TestPassword(t *testing.T) {
	if !PasswordRegister("12345678") || !PasswordRegister(strings.Repeat("a", 72)) || !PasswordRegister("        ") {
		t.Fatal("expected valid PasswordRegister")
	}
	if PasswordRegister("1234567") || PasswordRegister(strings.Repeat("a", 73)) {
		t.Fatal("expected invalid PasswordRegister")
	}

	if !PasswordLogin("1") || !PasswordLogin(strings.Repeat("a", 72)) {
		t.Fatal("expected valid PasswordLogin")
	}
	if PasswordLogin("") || PasswordLogin(strings.Repeat("a", 73)) {
		t.Fatal("expected invalid PasswordLogin")
	}
}