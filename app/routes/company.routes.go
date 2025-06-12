package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/pedersandvoll/project-insight-be/app/handlers"
	"github.com/pedersandvoll/project-insight-be/config/middleware"
)

func CompanyRoutes(app *fiber.App, h *handlers.Handlers) {
	app.Use(cors.New())
	api := app.Group("/company")
	api.Use(middleware.AuthRequired(h.JWTSecret))

	api.Post("/create", h.CreateCompany)
	api.Post("/join/:id", h.JoinCompany)
}
