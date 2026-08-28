# Platform-Orchestrator

A production-focused Platform Orchestrator written in Go for automating platform engineering workflows across Azure DevOps, Azure, Kubernetes, and Backstage.

## Overview

Platform-Orchestrator provides an API-driven orchestration layer for creating and managing developer platform resources.

The project is designed to sit behind developer-facing tools such as Backstage and provide a consistent, controlled workflow for provisioning repositories, CI/CD pipelines, cloud resources, and Kubernetes workloads.

Rather than requiring developers to understand the implementation details of Azure DevOps, Azure, Kubernetes, or other platform services, Platform-Orchestrator exposes higher-level platform operations through a consistent API.

For example, a repository creation request can select a predefined repository template:

```json
{
  "project": "Platform Engineering",
  "repositoryName": "payments-api",
  "template": "dotnet-api",
  "values": {
    "SERVICE_NAME": "payments-api",
    "NAMESPACE": "Company.Payments"
  }
}
```

The orchestrator can then:

1. Select the requested repository template.
2. Validate the template metadata and required values.
3. Render the template files.
4. Create or locate the Azure DevOps repository.
5. Push the generated repository content.
6. Create the associated Azure DevOps YAML pipeline.
7. Return the resulting repository and pipeline information.

This provides a foundation for implementing repeatable platform engineering workflows and developer self-service capabilities.

## Architecture

The application follows a ports-and-adapters style architecture, keeping orchestration logic independent from external platform implementations.

```text
                         ┌─────────────────────┐
                         │      Backstage      │
                         │   Developer Portal  │
                         └──────────┬──────────┘
                                    │
                                    │ HTTP
                                    ▼
                         ┌─────────────────────┐
                         │ Platform-Orchestrator│
                         │       Go API        │
                         └──────────┬──────────┘
                                    │
                    ┌───────────────┼───────────────┐
                    │               │               │
                    ▼               ▼               ▼
             ┌────────────┐  ┌────────────┐  ┌────────────┐
             │ Azure      │  │   Azure    │  │ Kubernetes │
             │ DevOps     │  │            │  │            │
             └────────────┘  └────────────┘  └────────────┘
```

The current repository provisioning workflow is structured around the following interfaces:

```text
HTTP API
   │
   ▼
Repository Service
   │
   ├── TemplateService
   │
   ├── RepositoryProvider
   │
   └── PipelineProvider
          │
          ▼
     Azure DevOps
```

This separation allows additional providers and platform capabilities to be introduced without coupling the core orchestration service directly to a particular external API.

## Repository Templates

Platform-Orchestrator supports named repository templates.

Templates are stored beneath the `templates` directory:

```text
templates/
├── dotnet-api/
│   ├── template.yaml
│   ├── README.md
│   └── .azuredevops/
│       └── azure-pipelines.yml
│
└── react-app/
    ├── template.yaml
    ├── README.md
    └── .azuredevops/
        └── azure-pipelines.yml
```

Each template contains a `template.yaml` metadata file describing the template and the values required to render it.

Example:

```yaml
apiVersion: platform-orchestrator/v1
kind: RepositoryTemplate

metadata:
  name: dotnet-api
  displayName: .NET API
  description: Production-ready ASP.NET Core API service
  version: "1.0.0"

spec:
  requiredValues:
    - name: SERVICE_NAME
      description: Name of the service
      example: payments-api

    - name: NAMESPACE
      description: Root .NET namespace
      example: Company.Payments

  pipeline:
    path: /.azuredevops/azure-pipelines.yml
```

Template values are rendered using tokens such as:

```text
__SERVICE_NAME__
__NAMESPACE__
```

The `template.yaml` file is orchestration metadata and is not included in the generated repository.

## Features

Current and planned platform capabilities include:

* Go-based REST API
* Gin HTTP routing
* OpenAPI and Swagger documentation using Swaggo
* Azure DevOps repository provisioning
* Azure DevOps YAML pipeline provisioning
* Named repository templates
* Template metadata and validation
* Required template value validation
* Recursive template rendering
* Template token replacement in file content and paths
* Repository creation idempotency
* Azure DevOps authentication abstraction
* Production-quality implementation
* Automated unit testing
* Integration testing support
* CI/CD support
* Container support
* Azure integration
* Kubernetes integration
* Helm deployment support
* Backstage integration
* Infrastructure as Code where applicable
* Cross-platform development support

## API

### Create Repository

```http
POST /v1/repositories
```

Example request:

```json
{
  "project": "Platform Engineering",
  "repositoryName": "payments-api",
  "template": "dotnet-api",
  "values": {
    "SERVICE_NAME": "payments-api",
    "NAMESPACE": "Company.Payments"
  }
}
```

A successful request returns HTTP `201 Created`.

Example response:

```json
{
  "repository": {
    "id": "repository-id",
    "name": "payments-api",
    "url": "https://dev.azure.com/example/Platform%20Engineering/_git/payments-api"
  },
  "pipeline": {
    "id": 123,
    "name": "payments-api-ci"
  }
}
```

### Health

```http
GET /health
```

Example response:

```json
{
  "status": "healthy"
}
```

### Swagger

When running locally, interactive API documentation is available at:

```text
http://localhost:8080/swagger/index.html
```

Swagger documentation is generated from annotations in the Go HTTP handlers.

Regenerate the documentation with:

```bash
swag init \
  -g cmd/api/main.go \
  -o docs \
  --parseInternal
```

## Getting Started

### Prerequisites

The core development environment requires:

* Git
* Go
* Azure DevOps access
* Azure DevOps authentication credentials

Additional tools may be required depending on the development or deployment workflow:

* Docker or Podman
* Terraform
* Helm
* kubectl
* PowerShell
* Bash

Refer to `go.mod` and the project-specific documentation for required dependency versions.

## Clone the Repository

```bash
git clone <repository-url>
cd Platform-Orchestrator
```

Install Go dependencies:

```bash
go mod download
```

Verify the project:

```bash
go test ./...
```

## Configuration

Local configuration and secrets should never be committed to source control.

Where provided, copy the example environment file:

```bash
cp .env.example .env
```

Update `.env` with the configuration required for your local environment.

Azure DevOps authentication is abstracted behind a token provider so that different authentication mechanisms can be used depending on the runtime environment.

Development environments may use a static token provider, while production environments should use an appropriate identity-based authentication mechanism such as Microsoft Entra ID, Managed Identity, or Kubernetes Workload Identity.

Never commit:

* Personal Access Tokens
* Azure credentials
* API keys
* Client secrets
* Certificates
* Private keys
* Access tokens
* Other sensitive configuration

## Development

The application entry point is:

```text
cmd/api/main.go
```

Core packages are located beneath:

```text
internal/
├── api/
├── azuredevops/
├── repository/
└── template/
```

Run the API locally with:

```bash
go run ./cmd/api
```

Format the source:

```bash
go fmt ./...
```

Run Go static analysis:

```bash
go vet ./...
```

Run the complete test suite:

```bash
go test ./...
```

Run tests with verbose output:

```bash
go test ./... -v
```

Run a specific package:

```bash
go test ./internal/repository -v
```

Run a specific test:

```bash
go test ./internal/repository \
  -run TestService_Create_ReusesExistingRepository \
  -v
```

## Testing

All changes should include appropriate automated tests.

The project is developed using a test-driven approach:

```text
RED
 │
 ▼
Write a failing test
 │
 ▼
GREEN
 │
 ▼
Implement the smallest change required
 │
 ▼
REFACTOR
 │
 ▼
Improve the implementation while tests remain green
```

Tests currently cover areas including:

* Template loading
* Template rendering
* Template metadata
* Required template values
* Invalid template paths
* Duplicate rendered paths
* Repository orchestration
* Repository idempotency
* Azure DevOps HTTP interactions
* Azure DevOps repository operations
* Azure DevOps pipeline operations
* API request validation
* HTTP response handling

Before submitting changes, run:

```bash
go fmt ./...
go vet ./...
go test ./...
```

## Project Structure

```text
Platform-Orchestrator/
├── cmd/
│   └── api/
│       └── main.go
│
├── docs/
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
│
├── internal/
│   ├── api/
│   ├── azuredevops/
│   ├── repository/
│   └── template/
│
├── templates/
│   ├── dotnet-api/
│   └── react-app/
│
├── go.mod
├── go.sum
└── README.md
```

## Roadmap

Platform-Orchestrator is being developed incrementally, with planned capabilities including:

* Idempotent repository provisioning
* Idempotent repository content updates
* Idempotent pipeline provisioning
* Partial orchestration state handling
* Structured logging
* Correlation IDs
* Metrics and observability
* Microsoft Entra ID authentication
* Azure Managed Identity
* Kubernetes Workload Identity
* Container hardening
* Kubernetes deployment
* Helm charts
* Readiness and liveness probes
* CI/CD pipelines
* Security scanning
* Backstage software template integration
* Additional repository templates
* Azure resource orchestration
* Production readiness documentation

## Contributing

Contributions are welcome.

Please read [CONTRIBUTING.md](./CONTRIBUTING.md) before submitting changes.

All contributors are expected to follow the project's Code of Conduct.

Changes should:

* Include appropriate automated tests.
* Maintain existing architectural boundaries.
* Follow Go formatting and linting standards.
* Avoid introducing secrets or environment-specific configuration.
* Keep external platform dependencies behind appropriate interfaces.
* Keep the full test suite green.

## Security

Please do not report security vulnerabilities through public GitHub issues.

See [SECURITY.md](./SECURITY.md) for information about responsibly reporting security issues.

Credentials and other secrets must never be stored in the repository.

Production deployments should prefer workload identity and short-lived credentials over long-lived static credentials.

## License

This project is licensed under the MIT License.

See [LICENSE](./LICENSE) for details.

## Maintainer

### Steve Owens

* GitHub: @sowens81
* Email: stevejowens81@outlook.com
