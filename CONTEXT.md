# App Monitoring & Management - Build Context

## Reference Codebase
The LMS Backend project at `/root/.openclaw/workspace/projects/lms-backend/` is the reference for structure, patterns, and style.

## Architecture Pattern (from LMS)
- **Clean Architecture:** handler → service → repository → model
- **Router:** Gin
- **ORM:** GORM with PostgreSQL
- **Config:** Viper (.env based)
- **Logging:** Zap (sugared)
- **Auth:** JWT (golang-jwt/jwt/v5), Bcrypt
- **DI:** Constructor injection (no wire/dig)
- **Models:** Use `uuid.UUID` for PKs, GORM tags, request/response structs in model package
- **Repositories:** Interface + struct, context-aware, pagination helper
- **Services:** Business logic, return sentinel errors
- **Handlers:** Gin context, bind JSON/query, use pkg/response + pkg/apierror
- **Middleware:** Auth (JWT Bearer), CORS, Logger, Recovery, RequestID, RateLimit
- **Tests:** testify/assert + testify/mock
- **Swagger:** swag annotations on handlers
- **Module:** `github.com/ronaldocristover/app-monitoring`
- **Go version:** 1.25.0

## ERD / Models

```
USERS: id(uuid), name, email, password_hash, created_at
APPS: id(uuid), app_name, description, tags(comma-sep), created_at
ENVIRONMENTS: id(uuid), app_id(FK), name, created_at
SERVERS: id(uuid), name, ip, provider, created_at
SERVICES: id(uuid), environment_id(FK), server_id(FK), name, type, url, repository, stack_language, stack_framework, db_type, db_host, created_at
MONITORING_CONFIGS: id(uuid), service_id(FK-unique), enabled, ping_interval_seconds, timeout_seconds, retries
DEPLOYMENTS: id(uuid), service_id(FK), method, container_name, port, config(json), created_at
BACKUPS: id(uuid), service_id(FK), enabled, path, schedule, last_backup_time, status
MONITORING_LOGS: id(uuid), service_id(FK), status(up/down), response_time_ms, status_code, error_message, checked_at
```

## API Endpoints

### Auth
- POST /api/v1/auth/register
- POST /api/v1/auth/login → {access_token, user}
- POST /api/v1/auth/refresh
- GET /api/v1/auth/me (protected)

### Apps CRUD
- POST /api/v1/apps
- GET /api/v1/apps (list, paginated, filter by tags)
- GET /api/v1/apps/{id} (detail with environments, services, monitoring status, backup info)
- PUT /api/v1/apps/{id}
- DELETE /api/v1/apps/{id}

### Full App (nested create/update)
- POST /api/v1/apps/full
- PUT /api/v1/apps/{id}/full

### Environments
- POST /api/v1/environments
- GET /api/v1/environments?app_id= (list by app)
- GET /api/v1/environments/{id}
- PUT /api/v1/environments/{id}
- DELETE /api/v1/environments/{id}

### Servers
- POST /api/v1/servers
- GET /api/v1/servers (list)
- GET /api/v1/servers/{id}
- PUT /api/v1/servers/{id}
- DELETE /api/v1/servers/{id}

### Services
- POST /api/v1/services
- GET /api/v1/services?environment_id=&server_id= (list)
- GET /api/v1/services/{id}
- PUT /api/v1/services/{id}
- DELETE /api/v1/services/{id}
- POST /api/v1/services/{id}/ping (manual ping)

### Monitoring Configs
- GET /api/v1/services/{id}/monitoring
- PUT /api/v1/services/{id}/monitoring

### Monitoring Logs
- GET /api/v1/services/{id}/logs (paginated history)

### Deployments
- POST /api/v1/services/{id}/deployments
- GET /api/v1/services/{id}/deployments (list)
- GET /api/v1/deployments/{id}
- PUT /api/v1/deployments/{id}
- DELETE /api/v1/deployments/{id}

### Backups
- POST /api/v1/services/{id}/backups
- GET /api/v1/services/{id}/backups (list)
- GET /api/v1/backups/{id}
- PUT /api/v1/backups/{id}
- DELETE /api/v1/backups/{id}

### Dashboard
- GET /api/v1/dashboard (summary: total apps, services up/down, recent incidents)

## Important: Read LMS reference files first!
Before writing any code, read the LMS project files to match the exact patterns:
- `internal/model/user.go` - model pattern
- `internal/repository/user.go` - repository pattern
- `internal/service/auth.go` - service pattern
- `internal/handler/auth.go` - handler pattern
- `internal/config/config.go` - config pattern
- `internal/middleware/auth.go` - middleware pattern
- `pkg/response/response.go` - response helpers
- `pkg/apierror/error.go` - error helpers
- `pkg/pagination/paginate.go` + `pagination.go` - pagination helpers
- `cmd/server/main.go` - main wiring pattern
- `cmd/server/routes.go` - routes pattern
- `go.mod` - dependencies
