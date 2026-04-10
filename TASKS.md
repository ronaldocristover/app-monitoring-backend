# App Monitoring & Management - Task Breakdown

## Agent 1: Foundation (Scaffold + Config + Models + Pkg)
- [ ] Initialize Go module `github.com/ronaldocristover/app-monitoring`
- [ ] Create directory structure (cmd/server, internal/{config,handler,middleware,model,repository,service,scheduler}, pkg/{apierror,pagination,response,validator})
- [ ] Copy & adapt pkg/ files from LMS (apierror, pagination, response, validator)
- [ ] Create config.go (Server, Database, JWT, CORS configs using Viper)
- [ ] Create .env.example
- [ ] Create all models: User, App, Environment, Server, Service, MonitoringConfig, Deployment, Backup, MonitoringLog
- [ ] Create middleware (auth.go with JWT, cors.go, logger.go, recovery.go, request_id.go, rate_limit.go)
- [ ] Create Dockerfile + docker-compose.yml

## Agent 2: Auth + Users
- [ ] Auth repository (Create, GetByEmail)
- [ ] Auth service (Register, Login, RefreshToken, Me)
- [ ] Auth handler (Register, Login, RefreshToken, Me)
- [ ] User repository (CRUD + List with pagination/search)
- [ ] User service (CRUD)
- [ ] User handler (CRUD endpoints)
- [ ] Unit tests for auth & user service

## Agent 3: Apps + Environments + Servers
- [ ] App repository (CRUD + List with pagination/search/filter by tags)
- [ ] App service (CRUD + Full Create/Update with nested environments+services)
- [ ] App handler (CRUD + Full Create/Update + Get Detail)
- [ ] Environment repository (CRUD + List by app)
- [ ] Environment service (CRUD)
- [ ] Environment handler (CRUD endpoints)
- [ ] Server repository (CRUD + List)
- [ ] Server service (CRUD)
- [ ] Server handler (CRUD endpoints)
- [ ] Unit tests

## Agent 4: Services + Monitoring + Backups + Deployments
- [ ] Service repository (CRUD + List by environment/server)
- [ ] Service service (CRUD)
- [ ] Service handler (CRUD + Manual Ping)
- [ ] MonitoringConfig repository + service + handler
- [ ] MonitoringLog repository + service + handler (list logs, get latest status)
- [ ] Backup repository + service + handler
- [ ] Deployment repository + service + handler
- [ ] Unit tests

## Agent 5: Monitoring Engine + Dashboard + Main Wiring
- [ ] Health check worker (ping services based on monitoring config, log results)
- [ ] Status change detection + notification trigger (interface for Telegram/Slack/Email)
- [ ] Scheduler for background monitoring jobs
- [ ] Dashboard handler (GET /api/v1/dashboard - summary stats)
- [ ] main.go - wire everything together (DB, repos, services, handlers, routes)
- [ ] routes.go - all route registration
- [ ] Swagger annotations
- [ ] Makefile
