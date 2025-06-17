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

func setupCompanyTestApp(t *testing.T) (*fiber.App, *helpers.TestDatabase, *helpers.TestData) {
	testDB := helpers.SetupTestDB(t)
	testDB.CleanDatabase()
	testData := testDB.SeedTestData()

	h := handlers.NewHandlers(testDB.DB, testDB.Config.JWTSecret)

	app := fiber.New()
	CompanyRoutes(app, h)

	return app, testDB, testData
}

func TestCompanyRoutes(t *testing.T) {
	app, testDB, testData := setupCompanyTestApp(t)
	defer testDB.TearDown()

	// Generate JWT token for the admin user
	adminToken, err := testDB.GenerateTestJWT(
		testData.Users[0].ID,
		testData.Users[0].Email,
		testData.Company.ID.String(),
	)
	require.NoError(t, err)

	// Generate JWT token for developer user (for join company test)
	devToken, err := testDB.GenerateTestJWT(
		testData.Users[1].ID,
		testData.Users[1].Email,
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
			description: "POST /company/create with valid data - should create company",
			route:       "/company/create",
			method:      "POST",
			token:       adminToken,
			body: map[string]interface{}{
				"name": "New Test Company",
			},
			expectedCode: 201,
			validateFunc: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Contains(t, response, "message")
				assert.Contains(t, response, "companyid")

				// Verify company was created in database
				var company tables.Companies
				companyID := response["companyid"].(string)
				result := testDB.DB.Where("id = ?", companyID).First(&company)
				assert.NoError(t, result.Error)
				assert.Equal(t, "New Test Company", company.Name)
				assert.Equal(t, testData.Users[0].ID, company.CreatedByID)
			},
		},
		{
			description:  "POST /company/join/:companyid with valid company ID - should join company",
			route:        "/company/join/" + testData.Company.ID.String(),
			method:       "POST",
			token:        devToken,
			expectedCode: 201,
			validateFunc: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Contains(t, response, "message")
				assert.Equal(t, "Successfully joined company", response["message"])

				// Verify company user relationship was created
				var companyUser tables.CompanyUsers
				result := testDB.DB.Where("company_id = ? AND user_id = ?", testData.Company.ID, testData.Users[1].ID).First(&companyUser)
				assert.NoError(t, result.Error)
				assert.Equal(t, testData.Company.ID, companyUser.CompanyID)
				assert.Equal(t, testData.Users[1].ID, companyUser.UserID)
			},
		},
		{
			description:  "POST /company/join/:companyid with invalid company ID - should return error",
			route:        "/company/join/invalid-uuid",
			method:       "POST",
			token:        devToken,
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

