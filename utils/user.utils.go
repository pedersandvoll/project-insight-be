package utils

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func GetUserIDFromClaims(claims jwt.MapClaims) (uuid.UUID, error) {
	userIDStr, ok := claims["userid"].(string)
	if !ok {
		return uuid.Nil, fmt.Errorf("invalid User ID type in JWT claims, expected string")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to parse User ID from JWT: %w", err)
	}

	return userID, nil
}
