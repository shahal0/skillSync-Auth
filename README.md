# Auth Service

## Overview
The Auth Service is a gRPC-based authentication and user management service for the SkillSync platform. It handles user registration, authentication, profile management, and token validation for both employers and candidates.

## Features
- User registration and authentication (employers and candidates)
- JWT token generation and validation
- Profile management
- Resume upload and storage using Google Cloud Storage
- User verification

## Technical Stack
- Go (Golang)
- gRPC/Protocol Buffers
- PostgreSQL with GORM ORM
- JWT for authentication
- Google Cloud Storage for resume storage

## Service Architecture
The service follows a clean architecture pattern:
- **Domain**: Contains business models and repository interfaces
- **Repository**: Implements data access logic
- **Usecase**: Implements business logic
- **Delivery**: Handles gRPC service implementation

## API Endpoints
The Auth Service exposes the following gRPC endpoints:
- Candidate registration and login
- Employer registration and login
- Profile management (get/update)
- Token validation
- Resume upload

## Configuration
Configuration is loaded from environment variables or a `.env` file. A sample environment file (`.env.sample`) is provided as a template.

### Environment Setup

```bash
# Copy the sample env file
cp .env.sample .env

# Edit the .env file with your specific values
```

### Required Environment Variables

Key environment variables include:

- **Server Configuration**
  - `PORT`: The port on which the service will listen (default: 50051)
  - `ENV`: Environment mode (`development` or `production`)

- **Database Configuration**
  - `DB_HOST`: Database host address
  - `DB_PORT`: Database port
  - `DB_USER`: Database username
  - `DB_PASSWORD`: Database password
  - `DB_NAME`: Database name
  - `DB_SSL_MODE`: SSL mode for database connection

- **JWT Configuration**
  - `JWT_SECRET`: Secret key for JWT token generation and validation
  - `JWT_EXPIRATION_HOURS`: JWT token expiration time in hours

- **Email Configuration**
  - `SMTP_HOST`: SMTP server host
  - `SMTP_PORT`: SMTP server port
  - `SMTP_USER`: SMTP username
  - `SMTP_PASSWORD`: SMTP password
  - `EMAIL_FROM`: Sender email address

- **Logging**
  - `LOG_LEVEL`: Logging level (`debug`, `info`, `warn`, `error`)

## Running the Service
1. Ensure PostgreSQL is running
2. Set up environment variables or `.env` file
3. Run the service:
   ```
   go run main.go
   ```

## Profiling
The Auth Service includes built-in profiling capabilities using Go's `pprof` package. The profiling server runs on port 6060.

### Accessing Profiling Data
1. While the service is running, access the profiling interface at: http://localhost:6060/debug/pprof/
2. Available profiles include:
   - CPU profiling: http://localhost:6060/debug/pprof/profile
   - Heap profiling: http://localhost:6060/debug/pprof/heap
   - Goroutine profiling: http://localhost:6060/debug/pprof/goroutine
   - Block profiling: http://localhost:6060/debug/pprof/block
   - Thread creation profiling: http://localhost:6060/debug/pprof/threadcreate

### Using the Go Tool
You can also use the Go tool to analyze profiles:

```bash
# CPU profile (30-second sample)
go tool pprof http://localhost:6060/debug/pprof/profile

# Memory profile
go tool pprof http://localhost:6060/debug/pprof/heap

# Goroutine profile
go tool pprof http://localhost:6060/debug/pprof/goroutine
```

Once in the pprof interactive mode, you can use commands like `top`, `web`, `list`, etc. to analyze the profile.

## Service Communication
- The Auth Service runs on port 50051
- Other services connect to the Auth Service for token validation and user information

## Important Notes
- JWT tokens must be in the format "Bearer {token}" for proper validation
- The service handles stripping the "Bearer " prefix from tokens when validating
