package handlers

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pedersandvoll/project-insight-be/app/types"
	"github.com/pedersandvoll/project-insight-be/config/tables"
	"github.com/pedersandvoll/project-insight-be/utils"
	"gorm.io/gorm"
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

	queries := c.Queries()

	var projects []tables.Projects
	query := h.db.Preload("CreatedBy").
		Preload("ModifiedBy").
		Preload("Budgets").
		Preload("AssociatedUsers").
		Preload("AssociatedUsers.User").
		Order("created_at DESC").
		Joins("JOIN company_projects ON projects.id = company_projects.project_id").
		Where("company_projects.company_id = ?", companyID)

	if statusQuery := utils.GetFilterValue(queries, "status"); statusQuery != "" {
		statuses := strings.Split(statusQuery, ",")
		for i, status := range statuses {
			statuses[i] = strings.TrimSpace(status)
		}
		query = query.Where("projects.status IN ?", statuses)
	}

	if name := utils.GetFilterValue(queries, "name"); name != "" {
		query = query.Where("projects.name ILIKE ?", "%"+name+"%")
	}

	if createdBy := utils.GetFilterValue(queries, "createdBy"); createdBy != "" {
		query = query.Where("projects.created_by_id = ?", createdBy)
	}

	if associated := utils.GetFilterValue(queries, "associated"); associated != "" {
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

	return c.Status(fiber.StatusOK).JSON(projects)
}

func (h *Handlers) GetProjectById(c *fiber.Ctx) error {
	projectID, err := utils.ParseUUIDParam(c, "projectid")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	var project tables.Projects
	result := h.db.Preload("CreatedBy").
		Preload("ModifiedBy").
		Preload("Budgets").
		Preload("AssociatedUsers").
		Preload("AssociatedUsers.User").
		Where("id = ?", projectID).
		First(&project)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Project not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Database error",
			"details": result.Error.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(project)
}

func (h *Handlers) AssignUserToProject(c *fiber.Ctx) error {
	var body types.AssignUserToProjectDTO

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	projectID, err := utils.ParseUUIDParam(c, "projectid")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	var user tables.Users
	userExist := h.db.Where("id = ?", body.UserID).First(&user)

	if userExist.Error != nil {
		if errors.Is(userExist.Error, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "User not found",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Database error",
			"details": userExist.Error.Error(),
		})
	}

	projectUser := tables.ProjectUsers{
		ProjectID: projectID,
		UserID:    body.UserID,
		Role:      body.Role,
	}
	userRole := tables.UserRoles{
		UserID: body.UserID,
		Role:   body.Role,
	}

	resultProjectUser := h.db.Create(&projectUser)

	if resultProjectUser.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to assign user to project",
			"msg":   resultProjectUser.Error.Error(),
		})
	}

	var existingUserRole tables.UserRoles
	roleCheckResult := h.db.
		Where("user_id = ?", body.UserID).
		Where("role = ?", body.Role).
		First(&existingUserRole)

	//Only create the role if it does not already exist
	if roleCheckResult.Error != nil {
		resultUserRole := h.db.Create(&userRole)

		if resultUserRole.Error != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create user role",
				"msg":   resultUserRole.Error.Error(),
			})
		}
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Assigned user to project successfully",
	})
}

func (h *Handlers) GetProjectsDashboard(c *fiber.Ctx) error {
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

	var totalProjects int64
	h.db.Table("projects").
		Joins("JOIN company_projects ON projects.id = company_projects.project_id").
		Where("company_projects.company_id = ?", companyID).
		Count(&totalProjects)

	var statusCounts []struct {
		Status tables.Status `json:"status"`
		Count  int64         `json:"count"`
	}
	h.db.Table("projects").
		Select("status, COUNT(*) as count").
		Joins("JOIN company_projects ON projects.id = company_projects.project_id").
		Where("company_projects.company_id = ?", companyID).
		Group("status").
		Find(&statusCounts)

	var totalEstimated struct {
		Total uint `json:"total"`
	}
	h.db.Table("projects").
		Select("SUM(estimated_cost) as total").
		Joins("JOIN company_projects ON projects.id = company_projects.project_id").
		Where("company_projects.company_id = ?", companyID).
		Find(&totalEstimated)

	var totalUsed struct {
		Total uint `json:"total"`
	}
	h.db.Table("budgets").
		Select("SUM(budget_used) as total").
		Joins("JOIN projects ON budgets.project_id = projects.id").
		Joins("JOIN company_projects ON projects.id = company_projects.project_id").
		Where("company_projects.company_id = ?", companyID).
		Find(&totalUsed)

	var recentProjects []tables.Projects
	h.db.Preload("Budgets").
		Order("created_at DESC").
		Limit(5).
		Find(&recentProjects)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"TotalProjects":  totalProjects,
		"ByStatus":       statusCounts,
		"TotalEstimated": totalEstimated.Total,
		"TotalUsed":      totalUsed.Total,
		"RecentProjects": recentProjects,
	})
}

func (h *Handlers) UpdateProjectStatus(c *fiber.Ctx) error {
	projectID, err := utils.ParseUUIDParam(c, "projectid")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	var body types.UpdateProjectStatusDTO

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	var project tables.Projects
	result := h.db.Where("id = ?", projectID).First(&project)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Project not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Database error",
			"details": result.Error.Error(),
		})
	}

	project.Status = body.Status

	saveResult := h.db.Save(&project)
	if saveResult.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to update project status in database",
			"details": saveResult.Error.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(project)
}
