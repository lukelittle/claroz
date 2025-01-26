package utils

import (
	"strings"
	"testing"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid password",
			password: "mySecurePassword123",
			wantErr:  false,
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  false, // bcrypt allows empty passwords
		},
		{
			name:     "long password",
			password: strings.Repeat("a", 72), // bcrypt's maximum length
			wantErr:  false,
		},
		{
			name:     "too long password",
			password: strings.Repeat("a", 73), // exceeds bcrypt's maximum length
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hashedPassword, err := HashPassword(tt.password)

			if (err != nil) != tt.wantErr {
				t.Errorf("HashPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if hashedPassword == "" {
					t.Error("HashPassword() returned empty hash")
				}

				if hashedPassword == tt.password {
					t.Error("HashPassword() returned unhashed password")
				}

				// Verify the hash starts with bcrypt identifier
				if !strings.HasPrefix(hashedPassword, "$2a$") {
					t.Error("HashPassword() did not return a bcrypt hash")
				}
			}
		})
	}
}

func TestComparePasswords(t *testing.T) {
	password := "mySecurePassword123"
	hashedPassword, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password for test setup: %v", err)
	}

	tests := []struct {
		name          string
		hashedPwd     string
		plainPwd      string
		wantErr       bool
		errorContains string
	}{
		{
			name:      "matching passwords",
			hashedPwd: hashedPassword,
			plainPwd:  password,
			wantErr:   false,
		},
		{
			name:          "wrong password",
			hashedPwd:     hashedPassword,
			plainPwd:      "wrongPassword",
			wantErr:       true,
			errorContains: "crypto/bcrypt: hashedPassword is not the hash of the given password",
		},
		{
			name:          "empty password",
			hashedPwd:     hashedPassword,
			plainPwd:      "",
			wantErr:       true,
			errorContains: "crypto/bcrypt: hashedPassword is not the hash of the given password",
		},
		{
			name:          "invalid hash format",
			hashedPwd:     "invalid-hash-format",
			plainPwd:      password,
			wantErr:       true,
			errorContains: "crypto/bcrypt: hashedSecret too short to be a bcrypted password",
		},
		{
			name:          "empty hash",
			hashedPwd:     "",
			plainPwd:      password,
			wantErr:       true,
			errorContains: "crypto/bcrypt: hashedSecret too short to be a bcrypted password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ComparePasswords(tt.hashedPwd, tt.plainPwd)

			if (err != nil) != tt.wantErr {
				t.Errorf("ComparePasswords() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errorContains != "" {
				if err == nil || !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("ComparePasswords() error = %v, want error containing %q", err, tt.errorContains)
				}
			}
		})
	}
}

func TestPasswordRoundTrip(t *testing.T) {
	originalPassword := "mySecurePassword123"

	// Hash the password
	hashedPassword, err := HashPassword(originalPassword)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	// Verify the hash can be compared successfully
	err = ComparePasswords(hashedPassword, originalPassword)
	if err != nil {
		t.Errorf("Failed to compare matching passwords: %v", err)
	}

	// Verify a different password fails comparison
	err = ComparePasswords(hashedPassword, "differentPassword")
	if err == nil {
		t.Error("ComparePasswords() succeeded with wrong password")
	}
}
