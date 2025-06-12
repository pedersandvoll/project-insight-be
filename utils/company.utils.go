package utils

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func GetCompanyIDFromClaims(claims jwt.MapClaims) (uuid.UUID, error) {
	companyIDInterface, exists := claims["companyid"]
	if !exists {
		return uuid.Nil, fmt.Errorf("company ID not found in JWT claims")
	}

	companyIDStr, ok := companyIDInterface.(string)
	if !ok {
		return uuid.Nil, fmt.Errorf("invalid Company ID type in JWT claims, expected string")
	}

	companyID, err := uuid.Parse(companyIDStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to parse Company ID from JWT: %w", err)
	}

	return companyID, nil
}
