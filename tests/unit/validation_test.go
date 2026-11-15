package unit

import (
	"strings"
	"testing"

	"user-management/internal/validation"
)

func TestNameValidation(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"English name", "Alice", true},
		{"Full name", "John Smith", true},
		{"Thai name", "สมชาย ใจดี", true},
		{"Hyphen", "Jean-Luc", true},
		{"Apostrophe", "O'Neil", true},
		{"Leading spaces", " Alice ", true},
		{"Contains number", "Alice123", false},
		{"Special character", "Alice@", false},
		{"Single character", "A", false},
		{"Empty", "", false},
		{"Only spaces", "   ", false},
		{"Leading hyphen", "-Alice", false},
		{"Trailing hyphen", "Alice-", false},
		{"Repeated spaces", "John  Smith", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validation.Name(tt.input)
			if got != tt.want {
				t.Errorf("Name(%q) = %v, want %v",
					tt.input, got, tt.want)
			}
		})
	}
}

func TestEmailValidation(t *testing.T) {
	longLocal := strings.Repeat("a", 65)
	longDomainLabel := strings.Repeat("a", 64)

	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"Normal email", "alice@example.com", true},
		{"Email with numbers", "alice123@example.com", true},
		{"Email with plus", "alice+tag@example.com", true},
		{"Subdomain", "alice@mail.example.com", true},
		{"Country domain", "alice@example.co.th", true},
		{"Domain hyphen", "alice@my-domain.com", true},
		{"Leading trailing spaces", " alice@example.com ", true},

		{"Empty", "", false},
		{"Missing at", "alice.example.com", false},
		{"Multiple at", "alice@@example.com", false},
		{"Missing local", "@example.com", false},
		{"Missing domain", "alice@", false},
		{"Missing TLD", "alice@example", false},
		{"Single character TLD", "alice@example.c", false},

		{"Leading local dot", ".alice@example.com", false},
		{"Trailing local dot", "alice.@example.com", false},
		{"Repeated local dots", "alice..smith@example.com", false},
		{"Repeated domain dots", "alice@example..com", false},

		{"Domain starts hyphen", "alice@-example.com", false},
		{"Domain ends hyphen", "alice@example-.com", false},
		{"Subdomain starts hyphen", "alice@mail.-example.com", false},
		{"Subdomain ends hyphen", "alice@mail.example-.com", false},
		{"Underscore in domain", "alice@exam_ple.com", false},

		{"Local part over 64 bytes", longLocal + "@example.com", false},
		{"Domain label over 63 bytes", "alice@" + longDomainLabel + ".com", false},
		{"Domain label empty", "alice@.example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validation.Email(tt.input)
			if got != tt.want {
				t.Errorf("Email(%q) = %v, want %v",
					tt.input, got, tt.want)
			}
		})
	}
}

func TestPasswordRegister(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"Minimum 8 bytes", "12345678", true},
		{"Normal password", "password123", true},
		{"Seven bytes", "1234567", false},
		{"Empty password", "", false},
		{"72 bytes", strings.Repeat("a", 72), true},
		{"73 bytes", strings.Repeat("a", 73), false},
		{"Password with spaces", "pass word", true},
		{"Unicode password", "รหัส1234", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validation.PasswordRegister(tt.input)
			if got != tt.want {
				t.Errorf("PasswordRegister() = %v, want %v",
					got, tt.want)
			}
		})
	}
}

func TestPasswordLogin(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"Normal password", "password123", true},
		{"Short existing password", "abc", true},
		{"Minimum register length", "12345678", true},
		{"Empty password", "", false},
		{"72 bytes", strings.Repeat("a", 72), true},
		{"73 bytes", strings.Repeat("a", 73), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validation.PasswordLogin(tt.input)
			if got != tt.want {
				t.Errorf("PasswordLogin() = %v, want %v",
					got, tt.want)
			}
		})
	}
}
