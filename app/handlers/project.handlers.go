package handlers

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/pedersandvoll/project-insight-be/app/types"
	"github.com/pedersandvoll/project-insight-be/config/tables"
	"github.com/pedersandvoll/project-insight-be/utils"

	"github.com/gofiber/fiber/v2"
)

func (h *Handlers) CreateProject(c *fiber.Ctx) error {
	token := c.Locals("user").(*jwt.Token)
	if token == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized - Missing JWT token",
		})
	}

	var body types.CreateProjectDTO

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if body.Name == "" || body.Description == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Name and description are required",
		})
	}

	claims := token.Claims.(jwt.MapClaims)

	userID, err := utils.GetUserIDFromClaims(claims)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	companyID, err := utils.GetCompanyIDFromClaims(claims)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	project := tables.Projects{
		Name:          body.Name,
		Description:   body.Description,
		Status:        body.Status,
		EstimatedCost: body.EstimatedCost,
		CreatedByID:   userID,
		ModifiedByID:  userID,
	}
	result := h.db.Create(&project)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create project",
			"msg":   result.Error.Error(),
		})

	}

	companyProject := tables.CompanyProjects{
		CompanyID: companyID,
		ProjectID: project.ID,
	}
	companyResult := h.db.Create(&companyProject)

	if companyResult.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create project company link",
			"msg":   companyResult.Error.Error(),
		})

	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":   "Project created successfully",
		"projectid": project.ID,
	})
}

func (h *Handlers) GetProjects(c *fiber.Ctx) error {
	token := c.Locals("user").(*jwt.Token)
	if token == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized - Missing JWT token",
		})
	}

	claims := token.Claims.(jwt.MapClaims)
	companyID, err := utils.GetCompanyIDFromClaims(claims)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	status := c.Query("status")
	name := c.Query("name")
	createdBy := c.Query("created_by")
	associated := c.Query("associated")

	var projects []tables.Projects
	query := h.db.Preload("CreatedBy").
		Preload("ModifiedBy").
		Preload("Budgets").
		Preload("AssociatedUsers").
		Preload("AssociatedUsers.User").
		Joins("JOIN company_projects ON projects.id = company_projects.project_id").
		Where("company_projects.company_id = ?", companyID)

	if status != "" {
		query = query.Where("projects.status = ?", status)
	}
	if name != "" {
		query = query.Where("projects.name ILIKE ?", "%"+name+"%")
	}
	if createdBy != "" {
		query = query.Where("projects.created_by_id = ?", createdBy)
	}
	if associated != "" {
		query = query.Joins("JOIN project_users ON projects.id = project_users.project_id").
			Where("project_users.user_id = ?", associated)
	}

	result := query.Find(&projects)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch projects",
			"msg":   result.Error.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"projects": projects,
	})
}

func (h *Handlers) AssignUserToProject(c *fiber.Ctx) error {
	var body types.AssignUserToProjectDTO

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if body.Role == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Role is required",
		})
	}

	projectIDStr := c.Params("projectid")
	if projectIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ProjectID is required in the URL",
		})
	}

	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid ProjectID format",
		})
	}

	projectUser := tables.ProjectUsers{
		ProjectID: projectID,
		UserID:    body.UserID,
		Role:      body.Role,
	}
	result := h.db.Create(&projectUser)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to assign user to project",
			"msg":   result.Error.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Assigned user to project successfully",
	})
}
