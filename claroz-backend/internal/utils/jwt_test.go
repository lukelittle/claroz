package utils

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestGenerateToken(t *testing.T) {
	userID := uuid.New()

	token, err := GenerateToken(userID)
	if err != nil {
		t.Errorf("GenerateToken() error = %v", err)
		return
	}

	if token == "" {
		t.Error("GenerateToken() returned empty token")
	}

	// Verify token can be parsed
	claims, err := ValidateToken(token)
	if err != nil {
		t.Errorf("Generated token validation failed: %v", err)
		return
	}

	if claims.UserID != userID {
		t.Errorf("Token claims UserID = %v, want %v", claims.UserID, userID)
	}
}

func TestValidateToken(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name      string
		setupFunc func() string
		wantErr   bool
		wantID    uuid.UUID
	}{
		{
			name: "valid token",
			setupFunc: func() string {
				token, _ := GenerateToken(userID)
				return token
			},
			wantErr: false,
			wantID:  userID,
		},
		{
			name: "expired token",
			setupFunc: func() string {
				claims := Claims{
					UserID: userID,
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(-24 * time.Hour)),
						IssuedAt:  jwt.NewNumericDate(time.Now().Add(-48 * time.Hour)),
					},
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, _ := token.SignedString(jwtSecret)
				return tokenString
			},
			wantErr: true,
			wantID:  uuid.Nil,
		},
		{
			name: "invalid signing method",
			setupFunc: func() string {
				claims := Claims{
					UserID: userID,
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
						IssuedAt:  jwt.NewNumericDate(time.Now()),
					},
				}
				token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
				// Using the secret as a private key is invalid, but that's what we want for this test
				tokenString, _ := token.SignedString(jwtSecret)
				return tokenString
			},
			wantErr: true,
			wantID:  uuid.Nil,
		},
		{
			name: "malformed token",
			setupFunc: func() string {
				return "malformed.token.string"
			},
			wantErr: true,
			wantID:  uuid.Nil,
		},
		{
			name: "empty token",
			setupFunc: func() string {
				return ""
			},
			wantErr: true,
			wantID:  uuid.Nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenString := tt.setupFunc()
			claims, err := ValidateToken(tokenString)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if claims == nil {
					t.Error("ValidateToken() claims is nil, want non-nil")
					return
				}
				if claims.UserID != tt.wantID {
					t.Errorf("ValidateToken() UserID = %v, want %v", claims.UserID, tt.wantID)
				}
			}
		})
	}
}

func TestTokenExpiration(t *testing.T) {
	userID := uuid.New()
	token, err := GenerateToken(userID)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	// Verify token is valid now
	claims, err := ValidateToken(token)
	if err != nil {
		t.Errorf("ValidateToken() error = %v", err)
	}
	if claims.UserID != userID {
		t.Errorf("ValidateToken() UserID = %v, want %v", claims.UserID, userID)
	}

	// Verify expiration time is set correctly
	if claims.ExpiresAt == nil {
		t.Error("Token expiration time not set")
	} else {
		expectedExpiration := time.Now().Add(24 * time.Hour)
		if claims.ExpiresAt.Time.Sub(expectedExpiration) > time.Minute {
			t.Errorf("Token expiration = %v, want close to %v", claims.ExpiresAt.Time, expectedExpiration)
		}
	}
}
