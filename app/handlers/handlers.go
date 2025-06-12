package handlers

import "github.com/pedersandvoll/project-insight-be/config/database"

type Handlers struct {
	db        *database.Database
	JWTSecret []byte
}

func NewHandlers(db *database.Database, jwtSecret string) *Handlers {
	return &Handlers{
		db:        db,
		JWTSecret: []byte(jwtSecret),
	}
}
