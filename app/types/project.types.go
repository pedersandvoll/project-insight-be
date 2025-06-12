package types

import (
	"github.com/google/uuid"
	"github.com/pedersandvoll/project-insight-be/config/tables"
)

type CreateProjectDTO struct {
	Name          string        `json:"name"`
	Description   string        `json:"description"`
	Status        tables.Status `json:"status"`
	EstimatedCost uint          `json:"estimatedcost"`
}

type AssignUserToProjectDTO struct {
	Role   string    `json:"role"`
	UserID uuid.UUID `json:"userid"`
}
