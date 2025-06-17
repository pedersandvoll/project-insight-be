package routes

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/pedersandvoll/project-insight-be/app/handlers"
	"github.com/pedersandvoll/project-insight-be/config/tables"
	"github.com/pedersandvoll/project-insight-be/test/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupProjectTestApp(t *testing.T) (*fiber.App, *helpers.TestDatabase, *helpers.TestData) {
	testDB := helpers.SetupTestDB(t)
	testDB.CleanDatabase()
	testData := testDB.SeedTestData()

	h := handlers.NewHandlers(testDB.DB, testDB.Config.JWTSecret)

	app := fiber.New()
	ProjectRoutes(app, h)

	return app, testDB, testData
}

func TestProjectRoutes(t *testing.T) {
	app, testDB, testData := setupProjectTestApp(t)
	defer testDB.TearDown()

	// Generate JWT token for the admin user
	adminToken, err := testDB.GenerateTestJWT(
		testData.Users[0].ID,
		testData.Users[0].Email,
		testData.Company.ID.String(),
	)
	require.NoError(t, err)

	tests := []struct {
		description  string
		route        string
		method       string
		token        string
		body         interface{}
		expectedCode int
		validateFunc func(t *testing.T, body []byte)
	}{
		{
			description: "POST /project/create with valid data - should create project",
			route:       "/project/create",
			method:      "POST",
			token:       adminToken,
			body: map[string]interface{}{
				"name":          "New Test Project",
				"description":   "A test project description",
				"status":        0, // tables.StatusPlanning
				"estimatedcost": 100000,
			},
			expectedCode: 201,
			validateFunc: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Contains(t, response, "message")
				assert.Contains(t, response, "projectid")

				// Verify project was created in database
				var project tables.Projects
				projectID := response["projectid"].(string)
				result := testDB.DB.Where("id = ?", projectID).First(&project)
				assert.NoError(t, result.Error)
				assert.Equal(t, "New Test Project", project.Name)
				assert.Equal(t, "A test project description", project.Description)
				assert.Equal(t, tables.Status(0), project.Status)
				assert.Equal(t, uint(100000), project.EstimatedCost)
			},
		},
		{
			description:  "GET /project/ - should return projects",
			route:        "/project/",
			method:       "GET",
			token:        adminToken,
			expectedCode: 200,
			validateFunc: func(t *testing.T, body []byte) {
				var projects []tables.Projects
				err := json.Unmarshal(body, &projects)
				require.NoError(t, err)
				assert.GreaterOrEqual(t, len(projects), 1) // We seeded 1 project
			},
		},
		{
			description:  "GET /project/dashboard - should return dashboard data",
			route:        "/project/dashboard",
			method:       "GET",
			token:        adminToken,
			expectedCode: 200,
			validateFunc: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Contains(t, response, "TotalProjects")
				assert.Contains(t, response, "ByStatus")
				assert.Contains(t, response, "TotalEstimated")
				assert.Contains(t, response, "TotalUsed")
				assert.Contains(t, response, "RecentProjects")
			},
		},
		{
			description:  "GET /project/:projectid - should return specific project",
			route:        "/project/" + testData.Project.ID.String(),
			method:       "GET",
			token:        adminToken,
			expectedCode: 200,
			validateFunc: func(t *testing.T, body []byte) {
				var project tables.Projects
				err := json.Unmarshal(body, &project)
				require.NoError(t, err)
				assert.Equal(t, testData.Project.ID, project.ID)
				assert.Equal(t, testData.Project.Name, project.Name)
			},
		},
		{
			description:  "GET /project/:projectid with invalid ID - should return error",
			route:        "/project/invalid-uuid",
			method:       "GET",
			token:        adminToken,
			expectedCode: 400,
			validateFunc: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Contains(t, response, "error")
			},
		},
		{
			description: "POST /project/assign/:projectid - should assign user to project",
			route:       "/project/assign/" + testData.Project.ID.String(),
			method:      "POST",
			token:       adminToken,
			body: map[string]interface{}{
				"userid": testData.Users[1].ID.String(), // Developer user
				"role":   3,                             // tables.RoleDeveloper
			},
			expectedCode: 201,
			validateFunc: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Contains(t, response, "message")
				assert.Equal(t, "Assigned user to project successfully", response["message"])

				// Verify project user relationship was created
				var projectUser tables.ProjectUsers
				result := testDB.DB.Where("project_id = ? AND user_id = ?", testData.Project.ID, testData.Users[1].ID).First(&projectUser)
				assert.NoError(t, result.Error)
				assert.Equal(t, tables.Role(3), projectUser.Role)
			},
		},
		{
			description: "POST /project/assign/:projectid with invalid project ID - should return error",
			route:       "/project/assign/invalid-uuid",
			method:      "POST",
			token:       adminToken,
			body: map[string]interface{}{
				"userid": testData.Users[1].ID.String(),
				"role":   3,
			},
			expectedCode: 400,
			validateFunc: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Contains(t, response, "error")
			},
		},
		{
			description: "POST /project/assign/:projectid with invalid user ID - should return error",
			route:       "/project/assign/" + testData.Project.ID.String(),
			method:      "POST",
			token:       adminToken,
			body: map[string]interface{}{
				"userid": "00000000-0000-0000-0000-000000000000", // Non-existent user
				"role":   3,
			},
			expectedCode: 404,
			validateFunc: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Contains(t, response, "error")
				assert.Equal(t, "User not found", response["error"])
			},
		},
		{
			description: "PATCH /project/status/:projectid - should update project status",
			route:       "/project/status/" + testData.Project.ID.String(),
			method:      "PATCH",
			token:       adminToken,
			body: map[string]interface{}{
				"status": 1, // tables.StatusInProgress
			},
			expectedCode: 200,
			validateFunc: func(t *testing.T, body []byte) {
				var project tables.Projects
				err := json.Unmarshal(body, &project)
				require.NoError(t, err)
				assert.Equal(t, tables.Status(1), project.Status)

				// Verify status was updated in database
				var dbProject tables.Projects
				result := testDB.DB.Where("id = ?", testData.Project.ID).First(&dbProject)
				assert.NoError(t, result.Error)
				assert.Equal(t, tables.Status(1), dbProject.Status)
			},
		},
		{
			description: "PATCH /project/status/:projectid with invalid project ID - should return error",
			route:       "/project/status/invalid-uuid",
			method:      "PATCH",
			token:       adminToken,
			body: map[string]interface{}{
				"status": 1,
			},
			expectedCode: 400,
			validateFunc: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Contains(t, response, "error")
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			var req *http.Request

			if test.body != nil {
				bodyBytes, _ := json.Marshal(test.body)
				req, _ = http.NewRequest(test.method, test.route, bytes.NewBuffer(bodyBytes))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req, _ = http.NewRequest(test.method, test.route, nil)
			}

			// Add Authorization header if token is provided
			if test.token != "" {
				req.Header.Set("Authorization", "Bearer "+test.token)
			}

			res, err := app.Test(req, -1)
			require.NoError(t, err)

			assert.Equal(t, test.expectedCode, res.StatusCode)

			if test.validateFunc != nil {
				bodyBytes, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				test.validateFunc(t, bodyBytes)
			}
		})
	}
}

