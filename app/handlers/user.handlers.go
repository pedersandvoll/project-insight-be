package handlers

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/pedersandvoll/project-insight-be/app/types"
	"github.com/pedersandvoll/project-insight-be/config/tables"
	"gorm.io/gorm"
)

func (h *Handlers) GetUsers(c *fiber.Ctx) error {
	var users []tables.Users
	result := h.db.Preload("Roles").
		Find(&users)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to retrieve users",
			"details": result.Error.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(users)
}

func (h *Handlers) AssignRoleToUser(c *fiber.Ctx) error {
	var body types.AssignRoleToUserDTO

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
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

	var existingUserRole tables.UserRoles
	roleCheckResult := h.db.
		Where("user_id = ?", body.UserID).
		Where("role = ?", body.Role).
		First(&existingUserRole)

	if roleCheckResult.Error == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"message": "User already has this role",
			"user_id": body.UserID,
			"role":    body.Role,
		})
	} else if !errors.Is(roleCheckResult.Error, gorm.ErrRecordNotFound) {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Database error checking existing role",
			"details": roleCheckResult.Error.Error(),
		})
	}

	userRole := tables.UserRoles{
		UserID: body.UserID,
		Role:   body.Role,
	}
	result := h.db.Create(&userRole)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to assigne user to role",
			"msg":   result.Error.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User assigned to role successfully",
	})
}

func (h *Handlers) GetUsersByRole(c *fiber.Ctx) error {
	roleStr := c.Params("role")
	if roleStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Role is required in the URL",
		})
	}

	roleInt64, err := strconv.ParseInt(roleStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid role format",
		})
	}
	targetRole := tables.Role(roleInt64)

	var users []tables.Users
	result := h.db.
		Joins("JOIN user_roles ON users.id = user_roles.user_id").
		Where("user_roles.role = ?", targetRole).
		Find(&users)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Database error",
			"details": result.Error.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(users)
}
