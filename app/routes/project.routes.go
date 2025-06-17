package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/pedersandvoll/project-insight-be/app/handlers"
	"github.com/pedersandvoll/project-insight-be/config/middleware"
)

// ProjectRoutes sets up project management routes (all protected)
func ProjectRoutes(app *fiber.App, h *handlers.Handlers) {
	app.Use(cors.New())
	api := app.Group("/project")
	api.Use(middleware.AuthRequired(h.JWTSecret))

	// Get all projects
	api.Get("/", h.GetProjects)
	// Get projects dashboard data
	api.Get("/dashboard", h.GetProjectsDashboard)
	// Get specific project by ID
	api.Get("/:projectid", h.GetProjectById)
	// Create a new project
	api.Post("/create", h.CreateProject)
	// Assign a user to a project
	api.Post("/assign/:projectid", h.AssignUserToProject)
	// Update project status
	api.Patch("/status/:projectid", h.UpdateProjectStatus)
}
