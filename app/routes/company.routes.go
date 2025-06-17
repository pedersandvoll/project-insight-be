package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/pedersandvoll/project-insight-be/app/handlers"
	"github.com/pedersandvoll/project-insight-be/config/middleware"
)

// CompanyRoutes sets up company management routes (all protected)
func CompanyRoutes(app *fiber.App, h *handlers.Handlers) {
	app.Use(cors.New())
	api := app.Group("/company")
	api.Use(middleware.AuthRequired(h.JWTSecret))

	// Create a new company
	api.Post("/create", h.CreateCompany)
	// Join an existing company by ID
	api.Post("/join/:companyid", h.JoinCompany)
}
