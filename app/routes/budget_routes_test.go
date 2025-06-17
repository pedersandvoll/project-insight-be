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

func setupBudgetTestApp(t *testing.T) (*fiber.App, *helpers.TestDatabase, *helpers.TestData) {
	testDB := helpers.SetupTestDB(t)
	testDB.CleanDatabase()
	testData := testDB.SeedTestData()

	h := handlers.NewHandlers(testDB.DB, testDB.Config.JWTSecret)

	app := fiber.New()
	BudgetRoutes(app, h)

	return app, testDB, testData
}

func TestBudgetRoutes(t *testing.T) {
	app, testDB, testData := setupBudgetTestApp(t)
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
			description: "POST /budget/create/:projectid with valid data - should create budget",
			route:       "/budget/create/" + testData.Project.ID.String(),
			method:      "POST",
			token:       adminToken,
			body: map[string]interface{}{
				"budgetused": 50000,
			},
			expectedCode: 201,
			validateFunc: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Contains(t, response, "message")
				assert.Contains(t, response, "budgetid")

				// Verify budget was created in database
				var budget tables.Budgets
				budgetID := response["budgetid"].(string)
				result := testDB.DB.Where("id = ?", budgetID).First(&budget)
				assert.NoError(t, result.Error)
				assert.Equal(t, uint(50000), budget.BudgetUsed)
				assert.Equal(t, testData.Project.ID, budget.ProjectID)
				assert.Equal(t, testData.Users[0].ID, budget.CreatedByID)
			},
		},
		{
			description: "POST /budget/create/:projectid with invalid project ID - should return error",
			route:       "/budget/create/invalid-uuid",
			method:      "POST",
			token:       adminToken,
			body: map[string]interface{}{
				"budgetused": 50000,
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
			description:  "POST /budget/create/:projectid with invalid body - should return error",
			route:        "/budget/create/" + testData.Project.ID.String(),
			method:       "POST",
			token:        adminToken,
			body:         "invalid json",
			expectedCode: 400,
			validateFunc: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Contains(t, response, "error")
				assert.Equal(t, "Invalid request body", response["error"])
			},
		},
		{
			description: "POST /budget/create/:projectid with zero budget - should create budget",
			route:       "/budget/create/" + testData.Project.ID.String(),
			method:      "POST",
			token:       adminToken,
			body: map[string]interface{}{
				"budgetused": 0,
			},
			expectedCode: 201,
			validateFunc: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Contains(t, response, "message")
				assert.Contains(t, response, "budgetid")

				// Verify budget was created with zero value
				var budget tables.Budgets
				budgetID := response["budgetid"].(string)
				result := testDB.DB.Where("id = ?", budgetID).First(&budget)
				assert.NoError(t, result.Error)
				assert.Equal(t, uint(0), budget.BudgetUsed)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			var req *http.Request

			if test.body != nil {
				if bodyStr, ok := test.body.(string); ok {
					// For invalid JSON test
					req, _ = http.NewRequest(test.method, test.route, bytes.NewBufferString(bodyStr))
				} else {
					bodyBytes, _ := json.Marshal(test.body)
					req, _ = http.NewRequest(test.method, test.route, bytes.NewBuffer(bodyBytes))
				}
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

