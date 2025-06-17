package helpers

import (
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/pedersandvoll/project-insight-be/config/database"
	"github.com/pedersandvoll/project-insight-be/config/tables"
	"github.com/pedersandvoll/project-insight-be/utils"
)

// TestDatabase holds the test database instance and configuration
type TestDatabase struct {
	DB        *database.Database
	Config    *database.Config
	JWTSecret []byte
}

// SetupTestDB creates a test database connection
func SetupTestDB(t *testing.T) *TestDatabase {
	config := &database.Config{
		Host:      getTestEnv("TEST_DB_HOST", "localhost"),
		Port:      getTestEnv("TEST_DB_PORT", "5433"), // Different port for test DB
		User:      getTestEnv("TEST_DB_USER", "test"),
		Password:  getTestEnv("TEST_DB_PASSWORD", "test"),
		DBName:    getTestEnv("TEST_DB_NAME", "project_insight_test"),
		SSLMode:   getTestEnv("TEST_DB_SSLMODE", "disable"),
		JWTSecret: getTestEnv("TEST_JWT_SECRET", "test-secret-key-for-testing"),
	}

	db, err := database.NewDatabase(config)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Run migrations
	tables.RunMigrations(db.DB)

	return &TestDatabase{
		DB:        db,
		Config:    config,
		JWTSecret: []byte(config.JWTSecret),
	}
}

// CleanDatabase removes all data from test database tables
func (td *TestDatabase) CleanDatabase() {
	// Clean in reverse order of dependencies
	td.DB.Exec("TRUNCATE TABLE company_projects CASCADE")
	td.DB.Exec("TRUNCATE TABLE project_users CASCADE")
	td.DB.Exec("TRUNCATE TABLE company_users CASCADE")
	td.DB.Exec("TRUNCATE TABLE budgets CASCADE")
	td.DB.Exec("TRUNCATE TABLE projects CASCADE")
	td.DB.Exec("TRUNCATE TABLE companies CASCADE")
	td.DB.Exec("TRUNCATE TABLE user_roles CASCADE")
	td.DB.Exec("TRUNCATE TABLE users CASCADE")
}

// SeedTestData creates minimal test data
func (td *TestDatabase) SeedTestData() *TestData {
	// Create test users
	hashedPassword, _ := utils.HashPassword("testpassword123")
	
	users := []tables.Users{
		{
			BaseModel: tables.BaseModel{ID: uuid.New()},
			FirstName: "John",
			LastName:  "Admin",
			Email:     "admin@test.com",
			Password:  hashedPassword,
		},
		{
			BaseModel: tables.BaseModel{ID: uuid.New()},
			FirstName: "Jane",
			LastName:  "Developer",
			Email:     "dev@test.com",
			Password:  hashedPassword,
		},
		{
			BaseModel: tables.BaseModel{ID: uuid.New()},
			FirstName: "Bob",
			LastName:  "Designer",
			Email:     "designer@test.com",
			Password:  hashedPassword,
		},
	}
	
	for i := range users {
		td.DB.Create(&users[i])
	}

	// Create user roles
	userRoles := []tables.UserRoles{
		{
			BaseModel: tables.BaseModel{ID: uuid.New()},
			Role:      tables.RoleProjectlead,
			UserID:    users[0].ID,
		},
		{
			BaseModel: tables.BaseModel{ID: uuid.New()},
			Role:      tables.RoleDeveloper,
			UserID:    users[1].ID,
		},
		{
			BaseModel: tables.BaseModel{ID: uuid.New()},
			Role:      tables.RoleDesigner,
			UserID:    users[2].ID,
		},
	}

	for i := range userRoles {
		td.DB.Create(&userRoles[i])
	}

	// Create test company
	company := tables.Companies{
		BaseModel:    tables.BaseModel{ID: uuid.New()},
		Name:         "Test Company",
		CreatedByID:  users[0].ID,
		ModifiedByID: users[0].ID,
	}
	td.DB.Create(&company)

	// Create company users
	companyUsers := []tables.CompanyUsers{
		{
			BaseModel: tables.BaseModel{ID: uuid.New()},
			CompanyID: company.ID,
			UserID:    users[0].ID,
		},
		{
			BaseModel: tables.BaseModel{ID: uuid.New()},
			CompanyID: company.ID,
			UserID:    users[1].ID,
		},
		{
			BaseModel: tables.BaseModel{ID: uuid.New()},
			CompanyID: company.ID,
			UserID:    users[2].ID,
		},
	}

	for i := range companyUsers {
		td.DB.Create(&companyUsers[i])
	}

	// Create test project
	project := tables.Projects{
		BaseModel:     tables.BaseModel{ID: uuid.New()},
		Name:          "Test Project",
		Description:   "A test project for unit testing",
		Status:        tables.StatusInProgress,
		EstimatedCost: 50000,
		CreatedByID:   users[0].ID,
		ModifiedByID:  users[0].ID,
	}
	td.DB.Create(&project)

	// Link company to project
	companyProject := tables.CompanyProjects{
		BaseModel: tables.BaseModel{ID: uuid.New()},
		CompanyID: company.ID,
		ProjectID: project.ID,
	}
	td.DB.Create(&companyProject)

	// Create project users
	projectUsers := []tables.ProjectUsers{
		{
			BaseModel: tables.BaseModel{ID: uuid.New()},
			ProjectID: project.ID,
			UserID:    users[0].ID,
			Role:      tables.RoleProjectlead,
		},
		{
			BaseModel: tables.BaseModel{ID: uuid.New()},
			ProjectID: project.ID,
			UserID:    users[1].ID,
			Role:      tables.RoleDeveloper,
		},
	}

	for i := range projectUsers {
		td.DB.Create(&projectUsers[i])
	}

	return &TestData{
		Users:           users,
		UserRoles:       userRoles,
		Company:         company,
		CompanyUsers:    companyUsers,
		Project:         project,
		CompanyProject:  companyProject,
		ProjectUsers:    projectUsers,
	}
}

// TestData holds references to seeded test data
type TestData struct {
	Users           []tables.Users
	UserRoles       []tables.UserRoles
	Company         tables.Companies
	CompanyUsers    []tables.CompanyUsers
	Project         tables.Projects
	CompanyProject  tables.CompanyProjects
	ProjectUsers    []tables.ProjectUsers
}

// GenerateTestJWT creates a valid JWT token for testing
func (td *TestDatabase) GenerateTestJWT(userID uuid.UUID, email, companyID string) (string, error) {
	claims := jwt.MapClaims{
		"userid":    userID.String(),
		"email":     email,
		"companyid": companyID,
		"exp":       time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(td.JWTSecret)
}

// TearDown closes the database connection
func (td *TestDatabase) TearDown() {
	if td.DB != nil {
		if sqlDB, err := td.DB.DB.DB(); err == nil {
			sqlDB.Close()
		}
	}
}

// getTestEnv gets environment variable with fallback for testing
func getTestEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}