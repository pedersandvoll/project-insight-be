package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/pedersandvoll/project-insight-be/app/handlers"
	"github.com/pedersandvoll/project-insight-be/config/middleware"
)

// UserRoutes sets up user management routes (all protected)
func UserRoutes(app *fiber.App, h *handlers.Handlers) {
	app.Use(cors.New())
	api := app.Group("/user")
	api.Use(middleware.AuthRequired(h.JWTSecret))

	// Get all users
	api.Get("/", h.GetUsers)
	// Get users filtered by specific role
	api.Get("/role/:role", h.GetUsersByRole)
	// Assign a role to a user
	api.Post("/assign/role", h.AssignRoleToUser)
}
