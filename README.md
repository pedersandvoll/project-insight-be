# Project Insight Backend

A Go-based REST API backend for project management and insights, built with Fiber framework and PostgreSQL.

## Features

- User authentication and authorization with JWT
- Company and project management
- Budget tracking and reporting
- User assignment to projects
- Project filtering and dashboard analytics

## Tech Stack

- **Go** (1.24.2)
- **Fiber** - Web framework
- **GORM** - ORM for database operations
- **PostgreSQL** - Database
- **JWT** - Authentication
- **Docker** - Containerization

## Prerequisites

- Go 1.24.2 or higher
- Docker & Docker Compose (recommended)
- Git

## Installation & Setup

### 1. Clone the Repository

```bash
git clone https://github.com/pedersandvoll/project-insight-be.git
cd project-insight-be
```

### 2. Environment Configuration

Create a `.env` file in the root directory:

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=your_db_user
DB_PASS=your_db_password
DB_NAME=project_insight
JWT_SECRET=your_jwt_secret_key
```

## Running the Project

### Option 1: Using Docker Compose + Go (Recommended)

1. Start the database:
```bash
docker-compose up -d
```

2. Install Go dependencies:
```bash
go mod download
```

3. Run the API server:
```bash
go run main.go
```

The server will start on `http://localhost:3000`

### Option 2: Using Go Only (Development)

Requires a local PostgreSQL installation.

1. Install dependencies:
```bash
go mod download
```

2. Run the application:
```bash
go run main.go
```

### Option 3: Using Make (if available)

```bash
make run
```

## API Endpoints

The API provides the following main endpoints:

- **Authentication**: `/auth/*`
- **Companies**: `/company/*`
- **Projects**: `/project/*`
- **Budgets**: `/budget/*`
- **Users**: `/user/*`

## Development

### Running Tests

```bash
go test ./...
```

### Building for Production

```bash
go build -o bin/main main.go
```

### Database Migrations

Migrations run automatically on application startup via GORM AutoMigrate.

## Project Structure

```
.
├── app/
│   ├── handlers/     # HTTP request handlers
│   ├── routes/       # Route definitions
│   └── types/        # Data transfer objects
├── config/
│   ├── database/     # Database configuration
│   ├── middleware/   # Custom middleware
│   └── tables/       # Database models
├── utils/            # Utility functions
├── http/             # HTTP test files
└── main.go          # Application entry point
```

