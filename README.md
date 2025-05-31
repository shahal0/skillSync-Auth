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
Configuration is loaded from environment variables or a `.env` file. Required configuration:
- Database connection details
- JWT secret
- Google Cloud Storage credentials

## Running the Service
1. Ensure PostgreSQL is running
2. Set up environment variables or `.env` file
3. Run the service:
   ```
   go run main.go
   ```

## Service Communication
- The Auth Service runs on port 50051
- Other services connect to the Auth Service for token validation and user information

## Important Notes
- JWT tokens must be in the format "Bearer {token}" for proper validation
- The service handles stripping the "Bearer " prefix from tokens when validating
