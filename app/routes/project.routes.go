package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/pedersandvoll/project-insight-be/app/handlers"
	"github.com/pedersandvoll/project-insight-be/config/middleware"
)

func ProjectRoutes(app *fiber.App, h *handlers.Handlers) {
	app.Use(cors.New())
	api := app.Group("/project")
	api.Use(middleware.AuthRequired(h.JWTSecret))

	api.Get("/", h.GetProjects)
	api.Get("/dashboard", h.GetProjectsDashboard)
	api.Get("/:projectid", h.GetProjectById)
	api.Post("/create", h.CreateProject)
	api.Post("/assign/:projectid", h.AssignUserToProject)
	api.Patch("/status/:projectid", h.UpdateProjectStatus)
}
