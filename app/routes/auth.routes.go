package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/pedersandvoll/project-insight-be/app/handlers"
)

func AuthRoutes(app *fiber.App, h *handlers.Handlers) {
	app.Use(cors.New())
	r := app.Group("/auth")

	r.Post("/register", h.RegisterUser)
	r.Post("/login", h.LoginUser)
}
