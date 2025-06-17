package utils_test

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/pedersandvoll/project-insight-be/utils"
	"testing"
)

func TestGetUserIDFromClaims(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		claims  jwt.MapClaims
		want    uuid.UUID
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := utils.GetUserIDFromClaims(tt.claims)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetUserIDFromClaims() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetUserIDFromClaims() succeeded unexpectedly")
			}
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("GetUserIDFromClaims() = %v, want %v", got, tt.want)
			}
		})
	}
}
