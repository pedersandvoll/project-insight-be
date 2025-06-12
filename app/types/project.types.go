package types

import "github.com/pedersandvoll/project-insight-be/config/tables"

type CreateProjectDTO struct {
	Name          string        `json:"name"`
	Description   string        `json:"description"`
	Status        tables.Status `json:"status"`
	EstimatedCost uint          `json:"estimatedcost"`
}
