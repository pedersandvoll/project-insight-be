package utils

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// GetFilterValue retrieves a value from the queries map using either the direct key
// or a formatted key with "filters[key]" pattern. Returns an empty string if the key
// is not found in either format.
func GetFilterValue(queries map[string]string, key string) string {
	if val, exists := queries[key]; exists {
		return val
	}
	if val, exists := queries["filters["+key+"]"]; exists {
		return val
	}
	return ""
}

func ParseUUIDParam(c *fiber.Ctx, paramName string) (uuid.UUID, error) {
	paramStr := c.Params(paramName)
	if paramStr == "" {
		return uuid.Nil, fmt.Errorf("%s is required in the URL", paramName)
	}

	paramUUID, err := uuid.Parse(paramStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid %s format", paramName)
	}

	return paramUUID, nil
}
