# App Monitoring & Management Backend

A backend service for monitoring and managing applications, built with **Go**, **Gin**, **GORM**, and **PostgreSQL**.

## Features

- **Authentication** — JWT-based auth (register, login, refresh token)
- **App Management** — CRUD with nested environments & services
- **Environment & Server Management** — Organize services by environment and server
- **Service Monitoring** — Configurable health checks with auto-ping
- **Monitoring Logs** — Response time, status code, and error tracking
- **Deployments** — Track deployment configs per service
- **Backups** — Configure and track backup schedules
- **Dashboard** — Summary stats (apps, services up/down, recent incidents)
- **Health Checks** — 5 endpoints (basic, status, detailed, live, ready)

## Architecture

Clean Architecture with strict layer separation:

```
Handler (HTTP) → Service (Business Logic) → Repository (Data Access) → Model (Domain)
```

## Tech Stack

| Component    | Technology            |
|-------------|----------------------|
| Language     | Go 1.25              |
| Router       | Gin                  |
| ORM          | GORM                 |
| Database     | PostgreSQL 16        |
| Auth         | JWT (golang-jwt/v5)  |
| Config       | Viper                |
| Logging      | Zap                  |
| Container    | Docker + Compose     |
| CI/CD        | GitHub Actions       |

## Quick Start

### Using Docker Compose

```bash
git clone git@github.com:ronaldocristover/app-monitoring-backend.git
cd app-monitoring-backend
docker compose up -d
```

Server runs on `http://localhost:8080`

### Manual Setup

```bash
# Prerequisites: Go 1.25+, PostgreSQL 16+

cp .env.example .env
# Edit .env with your database credentials

go mod download
go run ./cmd/server
```

### Build

```bash
make build
./bin/server
```

## API Endpoints

### Public

| Method | Endpoint                  | Description        |
|--------|---------------------------|--------------------|
| GET    | /health                   | Basic health check |
| GET    | /health/status            | DB status check    |
| GET    | /health/detailed          | Detailed DB stats  |
| GET    | /health/live              | Liveness probe     |
| GET    | /health/ready             | Readiness probe    |
| POST   | /api/v1/auth/register     | Register user      |
| POST   | /api/v1/auth/login        | Login              |
| POST   | /api/v1/auth/refresh      | Refresh token      |

### Protected (JWT Required)

| Method | Endpoint                                  | Description              |
|--------|-------------------------------------------|--------------------------|
| GET    | /api/v1/auth/me                           | Current user             |
| GET    | /api/v1/dashboard                         | Dashboard summary        |
| GET    | /api/v1/users                             | List users               |
| GET    | /api/v1/users/:id                         | Get user                 |
| PUT    | /api/v1/users/:id                         | Update user              |
| DELETE | /api/v1/users/:id                         | Delete user              |
| POST   | /api/v1/apps                              | Create app               |
| GET    | /api/v1/apps                              | List apps                |
| GET    | /api/v1/apps/:id                          | Get app                  |
| GET    | /api/v1/apps/:id/detail                   | Get app detail (nested)  |
| PUT    | /api/v1/apps/:id                          | Update app               |
| DELETE | /api/v1/apps/:id                          | Delete app               |
| POST   | /api/v1/apps/full                         | Create app (nested)      |
| PUT    | /api/v1/apps/:id/full                     | Update app (nested)      |
| POST   | /api/v1/environments                      | Create environment       |
| GET    | /api/v1/environments                      | List environments        |
| GET    | /api/v1/environments/:id                  | Get environment          |
| PUT    | /api/v1/environments/:id                  | Update environment       |
| DELETE | /api/v1/environments/:id                  | Delete environment       |
| POST   | /api/v1/servers                           | Create server            |
| GET    | /api/v1/servers                           | List servers             |
| GET    | /api/v1/servers/:id                       | Get server               |
| PUT    | /api/v1/servers/:id                       | Update server            |
| DELETE | /api/v1/servers/:id                       | Delete server            |
| POST   | /api/v1/services                          | Create service           |
| GET    | /api/v1/services                          | List services            |
| GET    | /api/v1/services/:id                      | Get service              |
| PUT    | /api/v1/services/:id                      | Update service           |
| DELETE | /api/v1/services/:id                      | Delete service           |
| POST   | /api/v1/services/:id/ping                 | Manual ping              |
| GET    | /api/v1/services/:id/monitoring           | Get monitoring config    |
| PUT    | /api/v1/services/:id/monitoring           | Update monitoring config |
| GET    | /api/v1/services/:id/logs                 | List monitoring logs     |
| POST   | /api/v1/services/:id/deployments          | Create deployment        |
| GET    | /api/v1/services/:id/deployments          | List deployments         |
| GET    | /api/v1/services/:id/deployments/:did     | Get deployment           |
| PUT    | /api/v1/services/:id/deployments/:did     | Update deployment        |
| DELETE | /api/v1/services/:id/deployments/:did     | Delete deployment        |
| POST   | /api/v1/services/:id/backups              | Create backup            |
| GET    | /api/v1/services/:id/backups              | List backups             |
| GET    | /api/v1/services/:id/backups/:bid         | Get backup               |
| PUT    | /api/v1/services/:id/backups/:bid         | Update backup            |
| DELETE | /api/v1/services/:id/backups/:bid         | Delete backup            |

## Project Structure

```
├── cmd/server/
│   ├── main.go              # Entry point, wiring
│   └── routes.go            # Route registration
├── internal/
│   ├── config/              # Configuration (Viper)
│   ├── handler/             # HTTP handlers
│   ├── middleware/           # Auth, CORS, logging, etc.
│   ├── model/               # Domain models (GORM)
│   ├── repository/          # Data access layer
│   ├── scheduler/           # Background monitoring worker
│   └── service/             # Business logic
├── pkg/
│   ├── apierror/            # Custom error types
│   ├── pagination/          # Pagination helpers
│   ├── response/            # HTTP response utilities
│   └── validator/           # Validation utilities
├── Dockerfile
├── docker-compose.yml
├── Makefile
└── .github/workflows/       # CI/CD
```

## Environment Variables

| Variable                | Default              | Description           |
|------------------------|----------------------|-----------------------|
| PORT                    | 8080                 | Server port           |
| ENV                     | development          | Environment           |
| SHUTDOWN_TIMEOUT        | 30s                  | Graceful shutdown     |
| POSTGRES_HOST           | localhost            | DB host               |
| POSTGRES_PORT           | 5432                 | DB port               |
| POSTGRES_USER           | app                  | DB user               |
| POSTGRES_PASSWORD       | secret               | DB password           |
| POSTGRES_DB             | app_monitoring       | DB name               |
| POSTGRES_MAX_OPEN       | 25                   | Max open connections  |
| POSTGRES_MAX_IDLE       | 5                    | Max idle connections  |
| JWT_SECRET              | -                    | JWT signing key       |
| JWT_EXPIRY              | 15m                  | Token expiry          |
| JWT_REFRESH_EXPIRY      | 168h                 | Refresh token expiry  |

## ERD

```
USERS ───────────────────────────────────────
APPS ──┬── ENVIRONMENTS ── SERVICES          │
       │                     ├── MONITORING_CONFIGS
       │                     ├── DEPLOYMENTS
       │                     ├── BACKUPS
SERVERS┤                     └── MONITORING_LOGS
       └─────────────────────────┘
```

## License

MIT
