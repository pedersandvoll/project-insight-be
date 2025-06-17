package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/pedersandvoll/project-insight-be/app/handlers"
	"github.com/pedersandvoll/project-insight-be/config/middleware"
)

// BudgetRoutes sets up budget management routes (all protected)
func BudgetRoutes(app *fiber.App, h *handlers.Handlers) {
	app.Use(cors.New())
	api := app.Group("/budget")
	api.Use(middleware.AuthRequired(h.JWTSecret))

	// Create a new budget for a specific project
	api.Post("/create/:projectid", h.CreateBudget)
}
