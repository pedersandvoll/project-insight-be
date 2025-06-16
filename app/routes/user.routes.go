package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/pedersandvoll/project-insight-be/app/handlers"
	"github.com/pedersandvoll/project-insight-be/config/middleware"
)

func UserRoutes(app *fiber.App, h *handlers.Handlers) {
	app.Use(cors.New())
	api := app.Group("/user")
	api.Use(middleware.AuthRequired(h.JWTSecret))

	api.Get("/", h.GetUsers)
	api.Get("/role/:role", h.GetUsersByRole)
	api.Post("/assign/role", h.AssignRoleToUser)
}
