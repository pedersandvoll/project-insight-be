package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/pedersandvoll/project-insight-be/app/handlers"
	"github.com/pedersandvoll/project-insight-be/config/middleware"
)

func BudgetRoutes(app *fiber.App, h *handlers.Handlers) {
	app.Use(cors.New())
	api := app.Group("/budget")
	api.Use(middleware.AuthRequired(h.JWTSecret))

	api.Post("/create/:id", h.CreateBudget)
}
