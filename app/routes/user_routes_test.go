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

func setupTestApp(t *testing.T) (*fiber.App, *helpers.TestDatabase, *helpers.TestData) {
	testDB := helpers.SetupTestDB(t)
	testDB.CleanDatabase()
	testData := testDB.SeedTestData()

	h := handlers.NewHandlers(testDB.DB, testDB.Config.JWTSecret)

	app := fiber.New()
	UserRoutes(app, h)

	return app, testDB, testData
}

func TestUserRoutesWithAuth(t *testing.T) {
	app, testDB, testData := setupTestApp(t)
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
			description:  "GET /user with valid auth - should return users",
			route:        "/user",
			method:       "GET",
			token:        adminToken,
			expectedCode: 200,
			validateFunc: func(t *testing.T, body []byte) {
				var users []tables.Users
				err := json.Unmarshal(body, &users)
				require.NoError(t, err)
				assert.GreaterOrEqual(t, len(users), 3) // We seeded 3 users

				// Check that passwords are not included
				for _, user := range users {
					assert.Empty(t, user.Password, "Password should not be returned in API response")
				}
			},
		},
		{
			description:  "GET /user/role/0 with valid auth - should return project leads",
			route:        "/user/role/0", // RoleProjectlead = 0
			method:       "GET",
			token:        adminToken,
			expectedCode: 200,
			validateFunc: func(t *testing.T, body []byte) {
				var users []tables.Users
				err := json.Unmarshal(body, &users)
				require.NoError(t, err)
				assert.GreaterOrEqual(t, len(users), 1) // We seeded 1 project lead

				// Verify the user has the correct role
				if len(users) > 0 {
					assert.Equal(t, "John", users[0].FirstName)
					assert.Equal(t, "admin@test.com", users[0].Email)
				}
			},
		},
		{
			description:  "GET /user/role/3 with valid auth - should return developers",
			route:        "/user/role/3", // RoleDeveloper = 3
			method:       "GET",
			token:        adminToken,
			expectedCode: 200,
			validateFunc: func(t *testing.T, body []byte) {
				var users []tables.Users
				err := json.Unmarshal(body, &users)
				require.NoError(t, err)
				assert.GreaterOrEqual(t, len(users), 1) // We seeded 1 developer

				if len(users) > 0 {
					assert.Equal(t, "Jane", users[0].FirstName)
					assert.Equal(t, "dev@test.com", users[0].Email)
				}
			},
		},
		{
			description:  "GET /user/role/999 with valid auth - should return empty array",
			route:        "/user/role/999", // Non-existent role
			method:       "GET",
			token:        adminToken,
			expectedCode: 200,
			validateFunc: func(t *testing.T, body []byte) {
				var users []tables.Users
				err := json.Unmarshal(body, &users)
				require.NoError(t, err)
				assert.Equal(t, 0, len(users))
			},
		},
		{
			description: "POST /user/assign/role with valid auth - should assign role",
			route:       "/user/assign/role",
			method:      "POST",
			token:       adminToken,
			body: map[string]interface{}{
				"userid": testData.Users[2].ID.String(), // Designer user
				"role":   1,                             // RoleContactperson
			},
			expectedCode: 201,
			validateFunc: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Contains(t, response, "message")

				// Verify the role was actually assigned in the database
				var userRole tables.UserRoles
				result := testDB.DB.Where("user_id = ? AND role = ?", testData.Users[2].ID, 1).First(&userRole)
				assert.NoError(t, result.Error)
				assert.Equal(t, tables.Role(1), userRole.Role)
			},
		},
		{
			description: "POST /user/assign/role with invalid user ID - should return error",
			route:       "/user/assign/role",
			method:      "POST",
			token:       adminToken,
			body: map[string]interface{}{
				"userid": "invalid-uuid",
				"role":   1,
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

			// Add Authorization header
			req.Header.Set("Authorization", "Bearer "+test.token)

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
