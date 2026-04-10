package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ronaldocristover/app-monitoring/pkg/response"
	"gorm.io/gorm"
)

type HealthHandler struct {
	db      *gorm.DB
	version string
}

func NewHealthHandler(db *gorm.DB, version string) *HealthHandler {
	return &HealthHandler{db: db, version: version}
}

// HealthCheck godoc
// @Summary      Health check
// @Description  Basic health check
// @Tags         health
// @Produce      json
// @Success      200  {object}  response.SuccessResponse
// @Router       /health [get]
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	response.Success(c, gin.H{
		"status":  "ok",
		"version": h.version,
	})
}

// HealthStatusCheck godoc
// @Summary      Health status
// @Description  Check service and database status
// @Tags         health
// @Produce      json
// @Success      200  {object}  response.SuccessResponse
// @Failure      503  {object}  response.ErrorResponse
// @Router       /health/status [get]
func (h *HealthHandler) HealthStatusCheck(c *gin.Context) {
	status := "ok"
	dbStatus := "up"

	sqlDB, err := h.db.DB()
	if err != nil {
		dbStatus = "down"
		status = "degraded"
	} else if err := sqlDB.Ping(); err != nil {
		dbStatus = "down"
		status = "degraded"
	}

	code := http.StatusOK
	if status != "ok" {
		code = http.StatusServiceUnavailable
	}

	c.JSON(code, gin.H{
		"success": true,
		"data": gin.H{
			"status":    status,
			"version":   h.version,
			"timestamp": time.Now().UTC(),
			"services": gin.H{
				"database": dbStatus,
			},
		},
	})
}

// HealthDetailedCheck godoc
// @Summary      Detailed health check
// @Description  Detailed health with DB stats
// @Tags         health
// @Produce      json
// @Success      200  {object}  response.SuccessResponse
// @Router       /health/detailed [get]
func (h *HealthHandler) HealthDetailedCheck(c *gin.Context) {
	sqlDB, err := h.db.DB()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"error":   gin.H{"code": 503, "message": "database connection failed"},
		})
		return
	}

	stats := sqlDB.Stats()
	response.Success(c, gin.H{
		"status":    "ok",
		"version":   h.version,
		"timestamp": time.Now().UTC(),
		"database": gin.H{
			"status":         "up",
			"max_open_conns": stats.MaxOpenConnections,
			"open_conns":     stats.OpenConnections,
			"in_use":         stats.InUse,
			"idle":           stats.Idle,
			"wait_count":     stats.WaitCount,
			"wait_duration":  stats.WaitDuration.String(),
		},
	})
}

// HealthLiveCheck godoc
// @Summary      Liveness probe
// @Description  Check if the application is alive
// @Tags         health
// @Produce      json
// @Success      200  {object}  response.SuccessResponse
// @Router       /health/live [get]
func (h *HealthHandler) HealthLiveCheck(c *gin.Context) {
	response.Success(c, gin.H{"status": "alive"})
}

// HealthReadyCheck godoc
// @Summary      Readiness probe
// @Description  Check if the application is ready to serve traffic
// @Tags         health
// @Produce      json
// @Success      200  {object}  response.SuccessResponse
// @Failure      503  {object}  response.ErrorResponse
// @Router       /health/ready [get]
func (h *HealthHandler) HealthReadyCheck(c *gin.Context) {
	sqlDB, err := h.db.DB()
	if err != nil || sqlDB.Ping() != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"error":   gin.H{"code": 503, "message": "not ready"},
		})
		return
	}
	response.Success(c, gin.H{"status": "ready"})
}
