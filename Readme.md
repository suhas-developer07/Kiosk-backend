# Kiosk Backend

Kiosk Backend is the server-side foundation for a campus-facing document and printing platform. It powers secure access for faculty, admins, colleges, and machine users, while handling file storage, print-job processing, and recharge workflows.

This service is designed as a maintainable internal backend for a production environment rather than a quick prototype. The codebase favors clear separation of concerns, dependency wiring at the composition root, and domain-oriented packages that make it easier to extend safely.

## Serverless architecture

This is a serverless backend built for AWS. It runs as an AWS Lambda function behind API Gateway and is deployed using AWS SAM. The application is intentionally written to fit the serverless model: stateless request handling, infrastructure defined as code, and runtime dependencies provided through environment variables and AWS services.

## What this system does

At a high level, the backend provides:

- File and document management for academic content
- Faculty, admin, and college authentication and authorization
- Print-job creation and related billing flows
- Machine and RFID-based recharge operations
- Operational endpoints for dashboards, activity tracking, and admin controls

## Project goals

The architecture aims to make the system easy to reason about by keeping the flow simple:

1. HTTP requests enter through Echo handlers
2. Middleware validates identity and access rules
3. Services contain the business logic
4. Repositories interact with MongoDB and storage backends
5. Responses are returned in a consistent way

That layered approach helps keep business rules isolated from transport concerns and persistence details.

## Architecture at a glance

The application is structured around a clean internal layering model:

- Entry point: main.go
  - Bootstraps the Lambda-based runtime and initializes the application
- Composition root: src/cmd
  - Wires up the Echo server, middleware, dependencies, and route registration
- Domain layer: src/internals/domain
  - Contains core entities, request/response models, and error definitions
- Handlers: src/internals/handlers
  - Converts HTTP requests into application actions and returns responses
- Services: src/internals/service
  - Implements business rules and orchestration logic
- Repositories: src/internals/repository
  - Encapsulates database access and persistence behavior
- Middleware: src/internals/middleware
  - Enforces authentication and authorization rules for different user roles
- Shared infrastructure: src/pkg
  - Provides storage abstractions, JWT helpers, hashing, validation, and utility code

## Request flow

A typical request follows this path:

- Route is registered in the router layer
- Middleware checks authentication/role requirements
- Handler parses the incoming request
- Service performs validation and business logic
- Repository reads or writes data from MongoDB
- File storage integration is used when documents are uploaded or accessed
- Response is returned to the client

This makes the codebase easier to test and maintain because each layer has a focused responsibility.

## Folder structure

```text
.
├── main.go
├── Dockerfile
├── Makefile
├── template.yaml
├── env.json
├── go.mod
├── src/
│   ├── cmd/
│   │   ├── app.go
│   │   └── router.go
│   ├── internals/
│   │   ├── domain/
│   │   ├── handlers/
│   │   ├── middleware/
│   │   ├── repository/
│   │   └── service/
│   └── pkg/
│       ├── filestore/
│       └── utils/
```

### Key areas

- src/cmd
  - Application bootstrap and router wiring
- src/internals/handlers
  - HTTP entry points for each business module
- src/internals/service
  - Core business logic and cross-module orchestration
- src/internals/repository
  - Persistence layer for MongoDB-backed data access
- src/pkg
  - Reusable infrastructure shared across the system

## Main modules

The backend is organized around business capabilities rather than technical layers alone:

- File service
  - Handles document retrieval, access, upload orchestration, and print-job operations
- Faculty module
  - Manages faculty identity, subject associations, and faculty-related actions
- Admin module
  - Supports admin profile management, moderation, faculty oversight, and administrative controls
- Payment and machine system
  - Covers recharge flows, machine users, cards, billing history, and related operations
- Orchestrator layer
  - Coordinates multi-step flows such as upload lifecycle and file-handling operations

## Runtime and deployment

This project is built to run as a serverless AWS application using:

- AWS Lambda for compute
- API Gateway for HTTP entry points
- AWS SAM for infrastructure and deployment
- Go for the application runtime
- Echo as the HTTP framework
- MongoDB for persistent storage
- S3-compatible storage for file assets

The deployment setup is defined through the provided SAM template and build configuration, which package the Go binary as the Lambda handler.

## Development setup

### Prerequisites

- Go 1.24 or newer
- Docker (optional, useful for containerized builds)
- Access to MongoDB and an S3-compatible storage bucket

### Environment variables

The service expects configuration such as:

- MONGO_URI
- FILES_BUCKET

These values are typically supplied through environment variables or deployment configuration.

### Local commands

```bash
go mod download
make build
make sam-api
```

### Build and containerization

The repository includes:

- A Makefile for local build automation
- A Dockerfile for building a container image
- A SAM template for serverless deployment

## Design principles

This codebase is structured with a few maintainability principles in mind:

- Clear module boundaries for each business domain
- Thin handlers focused on transport concerns
- Service layer for business rules and orchestration
- Repository layer for persistence details
- Shared utilities and storage abstractions for reuse
- Structured logging and explicit error handling for easier debugging

## Notes for future maintainers

When extending this project, the expected pattern is:

- Add or update the relevant domain model if new business data is needed
- Implement behavior in the service layer
- Connect persistence through the repository layer
- Expose the capability through the handler layer
- Keep middleware and cross-cutting concerns separate

This keeps the backend predictable and easier to evolve as new features are introduced.
