package utils_test

import (
	"testing"

	"github.com/pedersandvoll/project-insight-be/utils"
	"golang.org/x/crypto/bcrypt"
)

func TestVerifyPassword(t *testing.T) {
	validPassword := "password123"
	validHash, _ := bcrypt.GenerateFromPassword(
		[]byte(validPassword),
		bcrypt.DefaultCost,
	)

	tests := []struct {
		name     string
		password string
		hash     string
		want     bool
	}{
		{
			name:     "correct password",
			password: validPassword,
			hash:     string(validHash),
			want:     true,
		},
		{
			name:     "incorrect password",
			password: "wrongpassword",
			hash:     string(validHash),
			want:     false,
		},
		{
			name:     "empty password",
			password: "",
			hash:     string(validHash),
			want:     false,
		},
		{
			name:     "empty hash",
			password: validPassword,
			hash:     "",
			want:     false,
		},
		{
			name:     "nil hash (invalid input)",
			password: validPassword,
			hash:     string([]byte{0}),
			want:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := utils.VerifyPassword(tt.password, tt.hash)
			if got != tt.want {
				t.Errorf("VerifyPassword() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid password",
			password: "password123",
			wantErr:  false,
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  false,
		},
		{
			name:     "long password",
			password: "thisisareallylongpasswordthatshouldstillbehashedcorrectly",
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := utils.HashPassword(tt.password)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("HashPassword() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("HashPassword() succeeded unexpectedly")
			}

			// Verify the hashed password
			err := bcrypt.CompareHashAndPassword([]byte(got), []byte(tt.password))
			if err != nil {
				t.Errorf(
					"HashPassword() produced an un-verifiable hash: %v",
					err,
				)
			}
		})
	}
}
