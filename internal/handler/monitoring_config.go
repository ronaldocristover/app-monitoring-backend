package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/ronaldocristover/app-monitoring/internal/model"
	"github.com/ronaldocristover/app-monitoring/internal/service"
	"github.com/ronaldocristover/app-monitoring/pkg/apierror"
	"github.com/ronaldocristover/app-monitoring/pkg/response"
)

type MonitoringConfigHandler struct {
	svc service.MonitoringConfigService
}

func NewMonitoringConfigHandler(svc service.MonitoringConfigService) *MonitoringConfigHandler {
	return &MonitoringConfigHandler{svc: svc}
}

// Get godoc
// @Summary      Get monitoring config
// @Description  Get monitoring configuration for a service
// @Tags         monitoring
// @Produce      json
// @Param        id   path      string  true  "Service ID"
// @Success      200  {object}  response.SuccessResponse
// @Failure      400  {object}  response.ErrorResponse
// @Failure      404  {object}  response.ErrorResponse
// @Router       /services/{id}/monitoring [get]
func (h *MonitoringConfigHandler) Get(c *gin.Context) {
	serviceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apierror.BadRequest("Invalid service ID"))
		return
	}

	config, err := h.svc.GetByService(c.Request.Context(), serviceID)
	if err != nil {
		if err == service.ErrMonitoringConfigNotFound {
			response.Error(c, apierror.NotFound("Monitoring config not found"))
			return
		}
		response.Error(c, apierror.Internal("Failed to fetch monitoring config"))
		return
	}

	response.Success(c, config)
}

// Upsert godoc
// @Summary      Upsert monitoring config
// @Description  Create or update monitoring configuration for a service
// @Tags         monitoring
// @Accept       json
// @Produce      json
// @Param        id       path      string                            true  "Service ID"
// @Param        request  body      model.UpdateMonitoringConfigRequest  true  "Config data"
// @Success      200      {object}  response.SuccessResponse
// @Failure      400      {object}  response.ErrorResponse
// @Failure      404      {object}  response.ErrorResponse
// @Router       /services/{id}/monitoring [put]
func (h *MonitoringConfigHandler) Upsert(c *gin.Context) {
	serviceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apierror.BadRequest("Invalid service ID"))
		return
	}

	var req model.UpdateMonitoringConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apierror.BadRequest(err.Error()))
		return
	}

	config, err := h.svc.Upsert(c.Request.Context(), serviceID, &req)
	if err != nil {
		if err == service.ErrServiceNotFound {
			response.Error(c, apierror.NotFound("Service not found"))
			return
		}
		response.Error(c, apierror.Internal("Failed to save monitoring config"))
		return
	}

	response.Success(c, config)
}
