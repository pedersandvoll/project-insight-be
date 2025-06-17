# Testing

This directory contains test helpers and utilities for the project.

## Setup

### Test Database

The tests use a separate Postgres database for testing. Start it with:

```bash
docker compose up test-db -d
```

The test database runs on port `5433` with:
- Database: `project_insight_test`
- User: `test`
- Password: `test`

### Running Tests

Run all tests:
```bash
go test ./...
```

Run specific package tests:
```bash
go test ./app/routes -v
```

## Test Structure

### `helpers/database.go`
- `SetupTestDB()` - Creates test database connection
- `CleanDatabase()` - Removes all test data
- `SeedTestData()` - Creates minimal test data
- `GenerateTestJWT()` - Creates valid JWT tokens for testing

### Test Data

The test suite creates minimal seed data:
- 3 test users (Admin, Developer, Designer)
- 1 test company
- 1 test project
- Associated roles and relationships

### Test Coverage

Current tests cover:
- **Authentication**: Unauthorized access returns 401
- **User Routes**: GET /user, GET /user/role/:role, POST /user/assign/role
- **Data Validation**: Response structure and content verification
- **Database Integration**: Real database operations with cleanup

## Adding New Tests

1. Use the `setupTestApp(t)` helper to get app, database, and test data
2. Generate JWT tokens with `testDB.GenerateTestJWT()` for authenticated tests
3. Use `validateFunc` to verify response data structure and content
4. Always defer `testDB.TearDown()` to clean up connections