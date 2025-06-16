package types

import (
	"github.com/google/uuid"
	"github.com/pedersandvoll/project-insight-be/config/tables"
)

type AssignRoleToUserDTO struct {
	Role   tables.Role `json:"role"`
	UserID uuid.UUID   `json:"userid"`
}
