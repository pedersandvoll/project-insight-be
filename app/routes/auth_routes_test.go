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

func setupAuthTestApp(t *testing.T) (*fiber.App, *helpers.TestDatabase, *helpers.TestData) {
	testDB := helpers.SetupTestDB(t)
	testDB.CleanDatabase()
	testData := testDB.SeedTestData()

	h := handlers.NewHandlers(testDB.DB, testDB.Config.JWTSecret)

	app := fiber.New()
	AuthRoutes(app, h)

	return app, testDB, testData
}

func TestAuthRoutes(t *testing.T) {
	app, testDB, testData := setupAuthTestApp(t)
	defer testDB.TearDown()

	tests := []struct {
		description  string
		route        string
		method       string
		body         interface{}
		token        string
		expectedCode int
		validateFunc func(t *testing.T, body []byte)
	}{
		{
			description: "POST /auth/register - should create new user",
			route:       "/auth/register",
			method:      "POST",
			body: map[string]interface{}{
				"firstName": "Test",
				"lastName":  "User",
				"email":     "newuser@test.com",
				"password":  "testpassword123",
			},
			expectedCode: 201,
			validateFunc: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Contains(t, response, "message")
				assert.Contains(t, response, "userid")

				// Verify user was created in database
				var user tables.Users
				result := testDB.DB.Where("email = ?", "newuser@test.com").First(&user)
				assert.NoError(t, result.Error)
				assert.Equal(t, "Test", user.FirstName)
				assert.Equal(t, "User", user.LastName)
			},
		},
		{
			description: "POST /auth/register with existing email - should return error",
			route:       "/auth/register",
			method:      "POST",
			body: map[string]interface{}{
				"firstName": "Test",
				"lastName":  "User",
				"email":     testData.Users[0].Email, // Use existing user email
				"password":  "testpassword123",
			},
			expectedCode: 409,
			validateFunc: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Contains(t, response, "error")
			},
		},
		{
			description: "POST /auth/register with invalid data - should return error",
			route:       "/auth/register",
			method:      "POST",
			body: map[string]interface{}{
				"firstName": "",
				"lastName":  "",
				"email":     "invalid-email",
				"password":  "",
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
			description: "POST /auth/login with valid credentials - should return token",
			route:       "/auth/login",
			method:      "POST",
			body: map[string]interface{}{
				"email":    testData.Users[0].Email,
				"password": "testpassword123", // From test seed data
			},
			expectedCode: 200,
			validateFunc: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Contains(t, response, "token")
				
				// Verify token is a string and not empty
				token, ok := response["token"].(string)
				assert.True(t, ok)
				assert.NotEmpty(t, token)
			},
		},
		{
			description: "POST /auth/login with invalid email - should return error",
			route:       "/auth/login",
			method:      "POST",
			body: map[string]interface{}{
				"email":    "nonexistent@test.com",
				"password": "testpassword123",
			},
			expectedCode: 401,
			validateFunc: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Contains(t, response, "error")
			},
		},
		{
			description: "POST /auth/login with invalid password - should return error",
			route:       "/auth/login",
			method:      "POST",
			body: map[string]interface{}{
				"email":    testData.Users[0].Email,
				"password": "wrongpassword",
			},
			expectedCode: 401,
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

func TestAuthRoutesProtected(t *testing.T) {
	app, testDB, testData := setupAuthTestApp(t)
	defer testDB.TearDown()

	// Generate JWT token for testing protected routes
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
		expectedCode int
		validateFunc func(t *testing.T, body []byte)
	}{
		{
			description:  "GET /auth/currentuser with valid token - should return user info",
			route:        "/auth/currentuser",
			method:       "GET",
			token:        adminToken,
			expectedCode: 200,
			validateFunc: func(t *testing.T, body []byte) {
				var user tables.Users
				err := json.Unmarshal(body, &user)
				require.NoError(t, err)
				assert.Equal(t, testData.Users[0].Email, user.Email)
				assert.Equal(t, testData.Users[0].FirstName, user.FirstName)
				assert.Empty(t, user.Password, "Password should not be returned")
			},
		},
		{
			description:  "GET /auth/currentuser without token - should return unauthorized",
			route:        "/auth/currentuser",
			method:       "GET",
			token:        "",
			expectedCode: 401,
			validateFunc: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Contains(t, response, "error")
			},
		},
		{
			description:  "GET /auth/currentuser with invalid token - should return unauthorized",
			route:        "/auth/currentuser",
			method:       "GET",
			token:        "invalid-token",
			expectedCode: 401,
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
			req, _ := http.NewRequest(test.method, test.route, nil)

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