# CompanyFlow

CompanyFlow is a robust Human Resource Management System (HRMS) and ERP backend built with Go. It provides a multi-tenant architecture to manage companies, departments, employees, roles, leave requests, memos, and approvals.

## Features

- Multi-tenant Architecture: Support for multiple companies (tenants) within a single instance.
- Employee Management: Full CRUD operations for employee records, including bulk upload capabilities.
- Role-Based Access Control (RBAC): Flexible role and permission system to manage user access.
- Organization Structure: Manage departments, designations, and employee levels.
- Leave Management: System for requesting and tracking employee leaves.
- Memo and Approvals: Internal communication via memos with integrated approval workflows.
- Audit Logging: Tracking of system activities for compliance and monitoring.
- RESTful API: Clean and documented API endpoints using Swagger.
- JWT Authentication: Secure authentication and authorization for all endpoints.

## Tech Stack

- Language: Go (1.24.0)
- Database: PostgreSQL (using pgx/v5)
- Router: Gorilla Mux
- Authentication: JWT (JSON Web Tokens)
- Documentation: Swagger (swaggo)
- Configuration: Environment variables (godotenv)
- Migrations: Custom SQL-based migration runner

## Prerequisites

- Go 1.24 or higher
- PostgreSQL 14 or higher
- Git

## Installation and Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/falasefemi2/companyflowlow.git
   cd companyflowlow
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Configure environment variables:
   Create a .env file in the root directory and add the following:
   ```env
   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=your_user
   DB_PASSWORD=your_password
   DB_NAME=companyflow
   DB_SSLMODE=disable
   PORT=8080
   CORS_ORIGIN=http://localhost:3000
   JWT_SECRET=your_jwt_secret_key
   ```

4. Run the application:
   ```bash
   go run main.go
   ```
   The application will automatically run database migrations on startup.

## API Documentation

The API is documented using Swagger. Once the server is running, you can access the documentation at:
http://localhost:8080/swagger/index.html

## Project Structure

- /config: Database connection and environment configuration.
- /database: Migration runner and SQL migration files.
- /docs: Generated Swagger documentation.
- /dto: Data Transfer Objects for API requests and responses.
- /handlers: HTTP request handlers and route definitions.
- /models: Database models and structural definitions.
- /repositories: Database access layer for each module.
- /services: Business logic layer.
- /utils: Common utility functions and helpers.

## Database Migrations

Migrations are stored as SQL files in `database/migration/`. They are executed in alphabetical order based on their prefix (000_, 001_, etc.). The system tracks executed migrations in a `schema_migrations` table to ensure they only run once.

## Running Tests

To run the test suite, ensure you have a test database configured and set the `TEST_DATABASE_URL` environment variable or the standard DB variables in your .env.

Run all tests:
```bash
go test ./...
```

Run specific repository tests:
```bash
go test ./repositories/...
```

## Security

- Authentication: All protected routes require a valid JWT token in the Authorization header.
- Password Hashing: Employee passwords are hashed using bcrypt.
- CORS: Configurable CORS middleware to restrict access to trusted origins.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
