package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/pedersandvoll/project-insight-be/app/handlers"
)

func AuthRoutes(app *fiber.App, h *handlers.Handlers) {
	app.Use(cors.New())
	api := app.Group("/auth")

	api.Post("/register", h.RegisterUser)
	api.Post("/login", h.LoginUser)
}
