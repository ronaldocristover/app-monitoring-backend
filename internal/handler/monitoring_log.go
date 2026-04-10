package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/ronaldocristover/app-monitoring/internal/model"
	"github.com/ronaldocristover/app-monitoring/internal/service"
	"github.com/ronaldocristover/app-monitoring/pkg/apierror"
	"github.com/ronaldocristover/app-monitoring/pkg/pagination"
	"github.com/ronaldocristover/app-monitoring/pkg/response"
)

type MonitoringLogHandler struct {
	svc service.MonitoringLogService
}

func NewMonitoringLogHandler(svc service.MonitoringLogService) *MonitoringLogHandler {
	return &MonitoringLogHandler{svc: svc}
}

// ListByService godoc
// @Summary      List monitoring logs
// @Description  Get paginated monitoring logs for a service
// @Tags         monitoring-logs
// @Produce      json
// @Param        id         path   string  true   "Service ID"
// @Param        page       query  int     false  "Page number"
// @Param        page_size  query  int     false  "Items per page"
// @Param        status     query  string  false  "Filter by status (up/down)"
// @Success      200  {object}  response.PaginatedResponse
// @Failure      400  {object}  response.ErrorResponse
// @Failure      404  {object}  response.ErrorResponse
// @Router       /services/{id}/logs [get]
func (h *MonitoringLogHandler) ListByService(c *gin.Context) {
	serviceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apierror.BadRequest("Invalid service ID"))
		return
	}

	var req model.ListMonitoringLogsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, apierror.BadRequest(err.Error()))
		return
	}

	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}

	logs, total, err := h.svc.ListByService(c.Request.Context(), serviceID, &req)
	if err != nil {
		if err == service.ErrServiceNotFound {
			response.Error(c, apierror.NotFound("Service not found"))
			return
		}
		response.Error(c, apierror.Internal("Failed to fetch monitoring logs"))
		return
	}

	meta := pagination.NewMeta(req.Page, req.PageSize, total)
	response.Paginated(c, logs, &response.PaginationMeta{
		Page:       meta.Page,
		PageSize:   meta.PageSize,
		TotalItems: meta.TotalItems,
		TotalPages: meta.TotalPages,
	})
}
