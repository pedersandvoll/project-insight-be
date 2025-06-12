package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/pedersandvoll/project-insight-be/app/handlers"
	"github.com/pedersandvoll/project-insight-be/app/routes"
	"github.com/pedersandvoll/project-insight-be/config/database"
	"github.com/pedersandvoll/project-insight-be/config/tables"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	dbConfig := database.NewConfig()
	db, err := database.NewDatabase(dbConfig)
	if err != nil {
		log.Fatalf("Could not initialize database: %v", err)
	}

	tables.RunMigrations(db.DB)

	app := fiber.New()

	h := handlers.NewHandlers(db, dbConfig.JWTSecret)

	routes.AuthRoutes(app, h)
	routes.CompanyRoutes(app, h)

	app.Listen(":3000")
}
