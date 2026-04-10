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

type ServiceHandler struct {
	svc service.ServiceService
}

func NewServiceHandler(svc service.ServiceService) *ServiceHandler {
	return &ServiceHandler{svc: svc}
}

// Create godoc
// @Summary      Create a service
// @Description  Register a new service
// @Tags         services
// @Accept       json
// @Produce      json
// @Param        request  body      model.CreateServiceRequest  true  "Service data"
// @Success      201      {object}  response.SuccessResponse
// @Failure      400      {object}  response.ErrorResponse
// @Failure      500      {object}  response.ErrorResponse
// @Router       /services [post]
func (h *ServiceHandler) Create(c *gin.Context) {
	var req model.CreateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apierror.BadRequest(err.Error()))
		return
	}

	svc, err := h.svc.Create(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, apierror.Internal("Failed to create service"))
		return
	}

	response.Created(c, svc)
}

// Get godoc
// @Summary      Get service by ID
// @Description  Get service details
// @Tags         services
// @Produce      json
// @Param        id   path      string  true  "Service ID"
// @Success      200  {object}  response.SuccessResponse
// @Failure      400  {object}  response.ErrorResponse
// @Failure      404  {object}  response.ErrorResponse
// @Router       /services/{id} [get]
func (h *ServiceHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apierror.BadRequest("Invalid service ID"))
		return
	}

	svc, err := h.svc.GetByIDFull(c.Request.Context(), id)
	if err != nil {
		response.Error(c, apierror.NotFound("Service not found"))
		return
	}

	response.Success(c, svc)
}

// Update godoc
// @Summary      Update service
// @Description  Update an existing service
// @Tags         services
// @Accept       json
// @Produce      json
// @Param        id       path      string                     true  "Service ID"
// @Param        request  body      model.UpdateServiceRequest  true  "Service data"
// @Success      200      {object}  response.SuccessResponse
// @Failure      400      {object}  response.ErrorResponse
// @Failure      404      {object}  response.ErrorResponse
// @Router       /services/{id} [put]
func (h *ServiceHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apierror.BadRequest("Invalid service ID"))
		return
	}

	var req model.UpdateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apierror.BadRequest(err.Error()))
		return
	}

	svc, err := h.svc.Update(c.Request.Context(), id, &req)
	if err != nil {
		if err == service.ErrServiceNotFound {
			response.Error(c, apierror.NotFound("Service not found"))
			return
		}
		response.Error(c, apierror.Internal("Failed to update service"))
		return
	}

	response.Success(c, svc)
}

// Delete godoc
// @Summary      Delete service
// @Description  Delete a service by ID
// @Tags         services
// @Param        id  path  string  true  "Service ID"
// @Success      204
// @Failure      400  {object}  response.ErrorResponse
// @Failure      404  {object}  response.ErrorResponse
// @Router       /services/{id} [delete]
func (h *ServiceHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apierror.BadRequest("Invalid service ID"))
		return
	}

	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		if err == service.ErrServiceNotFound {
			response.Error(c, apierror.NotFound("Service not found"))
			return
		}
		response.Error(c, apierror.Internal("Failed to delete service"))
		return
	}

	response.NoContent(c)
}

// List godoc
// @Summary      List services
// @Description  List services with pagination and filters
// @Tags         services
// @Produce      json
// @Param        page            query  int     false  "Page number"
// @Param        page_size       query  int     false  "Items per page"
// @Param        environment_id  query  string  false  "Filter by environment ID"
// @Param        server_id       query  string  false  "Filter by server ID"
// @Param        search          query  string  false  "Search by name"
// @Success      200  {object}  response.PaginatedResponse
// @Failure      400  {object}  response.ErrorResponse
// @Router       /services [get]
func (h *ServiceHandler) List(c *gin.Context) {
	var req model.ListServicesRequest
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

	services, total, err := h.svc.List(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, apierror.Internal("Failed to fetch services"))
		return
	}

	meta := pagination.NewMeta(req.Page, req.PageSize, total)
	response.Paginated(c, services, &response.PaginationMeta{
		Page:       meta.Page,
		PageSize:   meta.PageSize,
		TotalItems: meta.TotalItems,
		TotalPages: meta.TotalPages,
	})
}

// ManualPing godoc
// @Summary      Manual ping
// @Description  Perform a manual HTTP GET ping to the service and record the result
// @Tags         services
// @Produce      json
// @Param        id   path      string  true  "Service ID"
// @Success      200  {object}  response.SuccessResponse
// @Failure      400  {object}  response.ErrorResponse
// @Failure      404  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /services/{id}/ping [post]
func (h *ServiceHandler) ManualPing(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apierror.BadRequest("Invalid service ID"))
		return
	}

	log, err := h.svc.ManualPing(c.Request.Context(), id)
	if err != nil {
		if err == service.ErrServiceNotFound {
			response.Error(c, apierror.NotFound("Service not found"))
			return
		}
		if apiErr, ok := err.(*apierror.Error); ok {
			response.Error(c, apiErr)
			return
		}
		response.Error(c, apierror.Internal("Failed to ping service"))
		return
	}

	response.Success(c, log)
}
