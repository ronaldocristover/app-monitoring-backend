// @title           App Monitoring & Management API
// @version         1.0.0
// @description     Application Monitoring and Management System Backend API
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.email  admin@app-monitoring.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /
// @schemes   http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"github.com/ronaldocristover/app-monitoring/internal/config"
	"github.com/ronaldocristover/app-monitoring/internal/handler"
	"github.com/ronaldocristover/app-monitoring/internal/model"
	"github.com/ronaldocristover/app-monitoring/internal/repository"
	"github.com/ronaldocristover/app-monitoring/internal/scheduler"
	"github.com/ronaldocristover/app-monitoring/internal/service"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	cfg, err := config.Load(".env")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	var sugar *zap.SugaredLogger
	if cfg.Server.Env == "production" {
		prodLog, _ := zap.NewProduction()
		sugar = prodLog.Sugar()
	} else {
		devLog, _ := zap.NewDevelopment()
		sugar = devLog.Sugar()
	}
	defer sugar.Sync()

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
	)

	gormConfig := &gorm.Config{}
	if cfg.Server.Env != "production" {
		gormConfig.Logger = logger.Default.LogMode(logger.Info)
	}

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		sugar.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		sugar.Fatalf("Failed to get underlying sql.DB: %v", err)
	}
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpen)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdle)

	// Auto-migrate all models
	if err := db.AutoMigrate(
		&model.User{},
		&model.App{},
		&model.Environment{},
		&model.Server{},
		&model.Service{},
		&model.MonitoringConfig{},
		&model.MonitoringLog{},
		&model.Deployment{},
		&model.Backup{},
	); err != nil {
		sugar.Fatalf("Failed to migrate database: %v", err)
	}

	// Create repositories
	userRepo := repository.NewUserRepository(db)
	appRepo := repository.NewAppRepository(db)
	envRepo := repository.NewEnvironmentRepository(db)
	serverRepo := repository.NewServerRepository(db)
	serviceRepo := repository.NewServiceRepository(db)
	monitoringConfigRepo := repository.NewMonitoringConfigRepository(db)
	monitoringLogRepo := repository.NewMonitoringLogRepository(db)
	deploymentRepo := repository.NewDeploymentRepository(db)
	backupRepo := repository.NewBackupRepository(db)

	// Create services
	authSvc := service.NewAuthService(userRepo, cfg.JWT.Secret, cfg.JWT.Expiry, cfg.JWT.RefreshExpiry)
	userSvc := service.NewUserService(userRepo)
	appSvc := service.NewAppService(appRepo, db, sugar)
	envSvc := service.NewEnvironmentService(envRepo, sugar)
	serverSvc := service.NewServerService(serverRepo, sugar)
	serviceSvc := service.NewServiceService(serviceRepo, monitoringLogRepo, sugar)
	monitoringConfigSvc := service.NewMonitoringConfigService(monitoringConfigRepo, serviceRepo, sugar)
	monitoringLogSvc := service.NewMonitoringLogService(monitoringLogRepo, serviceRepo, sugar)
	deploymentSvc := service.NewDeploymentService(deploymentRepo, serviceRepo, sugar)
	backupSvc := service.NewBackupService(backupRepo, serviceRepo, sugar)
	dashboardSvc := service.NewDashboardService(appRepo, serviceRepo, monitoringLogRepo, envRepo, db, sugar)

	// Create handlers
	healthHandler := handler.NewHealthHandler(db, "1.0.0")
	authHandler := handler.NewAuthHandler(authSvc, userSvc)
	userHandler := handler.NewUserHandler(userSvc)
	appHandler := handler.NewAppHandler(appSvc)
	envHandler := handler.NewEnvironmentHandler(envSvc)
	serverHandler := handler.NewServerHandler(serverSvc)
	serviceHandler := handler.NewServiceHandler(serviceSvc)
	monitoringConfigHandler := handler.NewMonitoringConfigHandler(monitoringConfigSvc)
	monitoringLogHandler := handler.NewMonitoringLogHandler(monitoringLogSvc)
	deploymentHandler := handler.NewDeploymentHandler(deploymentSvc)
	backupHandler := handler.NewBackupHandler(backupSvc)
	dashboardHandler := handler.NewDashboardHandler(dashboardSvc)

	// Create handlers struct for routes
	h := &handlers{
		Health:             healthHandler,
		Auth:               authHandler,
		User:               userHandler,
		App:                appHandler,
		Environment:        envHandler,
		Server:             serverHandler,
		Service:            serviceHandler,
		MonitoringConfig:   monitoringConfigHandler,
		MonitoringLog:      monitoringLogHandler,
		Deployment:         deploymentHandler,
		Backup:             backupHandler,
		Dashboard:          dashboardHandler,
	}

	// Start monitoring scheduler
	sched := scheduler.NewScheduler(4, 100)
	sched.Start()

	// Create and start monitoring worker
	worker := scheduler.NewMonitoringWorker(
		monitoringConfigRepo,
		serviceRepo,
		monitoringLogRepo,
		sched,
		sugar,
		30*time.Second,
	)
	worker.Start()

	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()

	// Setup routes
	setupRoutes(router, h, cfg, sugar)

	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: router,
	}

	go func() {
		sugar.Infof("Server starting on port %s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			sugar.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	sugar.Info("Shutting down server...")

	worker.Stop()
	sched.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		sugar.Fatalf("Server forced to shutdown: %v", err)
	}

	sugar.Info("Server exited gracefully")
}

type handlers struct {
	Health              *handler.HealthHandler
	Auth                *handler.AuthHandler
	User               *handler.UserHandler
	App                *handler.AppHandler
	Environment        *handler.EnvironmentHandler
	Server             *handler.ServerHandler
	Service            *handler.ServiceHandler
	MonitoringConfig   *handler.MonitoringConfigHandler
	MonitoringLog      *handler.MonitoringLogHandler
	Deployment         *handler.DeploymentHandler
	Backup             *handler.BackupHandler
	Dashboard          *handler.DashboardHandler
}
