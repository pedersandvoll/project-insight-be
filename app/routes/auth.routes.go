package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/pedersandvoll/project-insight-be/app/handlers"
	"github.com/pedersandvoll/project-insight-be/config/middleware"
)

// AuthRoutes sets up authentication and user management routes
func AuthRoutes(app *fiber.App, h *handlers.Handlers) {
	app.Use(cors.New())
	api := app.Group("/auth")

	// User registration and login (public routes)
	api.Post("/register", h.RegisterUser)
	api.Post("/login", h.LoginUser)

	// Protected routes requiring authentication
	api.Use(middleware.AuthRequired(h.JWTSecret))
	// Get current authenticated user information
	api.Get("/currentuser", h.GetCurrentUserInformation)
}
