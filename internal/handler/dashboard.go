package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/ronaldocristover/app-monitoring/internal/service"
	"github.com/ronaldocristover/app-monitoring/pkg/apierror"
	"github.com/ronaldocristover/app-monitoring/pkg/response"
)

type DashboardHandler struct {
	dashboardSvc service.DashboardService
}

func NewDashboardHandler(dashboardSvc service.DashboardService) *DashboardHandler {
	return &DashboardHandler{dashboardSvc: dashboardSvc}
}

// GetStats godoc
// @Summary      Get dashboard stats
// @Description  Returns dashboard stats with total apps, services up/down, recent incidents, and environment breakdown
// @Tags         dashboard
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.SuccessResponse{data=service.DashboardStats}
// @Failure      500  {object}  response.ErrorResponse
// @Router       /api/v1/dashboard [get]
func (h *DashboardHandler) GetStats(c *gin.Context) {
	stats, err := h.dashboardSvc.GetDashboardStats(c.Request.Context())
	if err != nil {
		response.Error(c, apierror.Internal("Failed to fetch dashboard stats"))
		return
	}

	response.Success(c, stats)
}
