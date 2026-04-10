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

type DeploymentHandler struct {
	svc service.DeploymentService
}

func NewDeploymentHandler(svc service.DeploymentService) *DeploymentHandler {
	return &DeploymentHandler{svc: svc}
}

// Create godoc
// @Summary      Create deployment
// @Description  Create a new deployment for a service
// @Tags         deployments
// @Accept       json
// @Produce      json
// @Param        id       path      string                       true  "Service ID"
// @Param        request  body      model.CreateDeploymentRequest  true  "Deployment data"
// @Success      201      {object}  response.SuccessResponse
// @Failure      400      {object}  response.ErrorResponse
// @Failure      404      {object}  response.ErrorResponse
// @Failure      500      {object}  response.ErrorResponse
// @Router       /services/{id}/deployments [post]
func (h *DeploymentHandler) Create(c *gin.Context) {
	serviceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apierror.BadRequest("Invalid service ID"))
		return
	}

	var req model.CreateDeploymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apierror.BadRequest(err.Error()))
		return
	}
	req.ServiceID = serviceID

	deployment, err := h.svc.Create(c.Request.Context(), &req)
	if err != nil {
		if err == service.ErrServiceNotFound {
			response.Error(c, apierror.NotFound("Service not found"))
			return
		}
		response.Error(c, apierror.Internal("Failed to create deployment"))
		return
	}

	response.Created(c, deployment)
}

// Get godoc
// @Summary      Get deployment
// @Description  Get deployment by ID
// @Tags         deployments
// @Produce      json
// @Param        id   path      string  true  "Deployment ID"
// @Success      200  {object}  response.SuccessResponse
// @Failure      400  {object}  response.ErrorResponse
// @Failure      404  {object}  response.ErrorResponse
// @Router       /deployments/{id} [get]
func (h *DeploymentHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apierror.BadRequest("Invalid deployment ID"))
		return
	}

	deployment, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, apierror.NotFound("Deployment not found"))
		return
	}

	response.Success(c, deployment)
}

// Update godoc
// @Summary      Update deployment
// @Description  Update an existing deployment
// @Tags         deployments
// @Accept       json
// @Produce      json
// @Param        id       path      string                        true  "Deployment ID"
// @Param        request  body      model.UpdateDeploymentRequest  true  "Deployment data"
// @Success      200      {object}  response.SuccessResponse
// @Failure      400      {object}  response.ErrorResponse
// @Failure      404      {object}  response.ErrorResponse
// @Router       /deployments/{id} [put]
func (h *DeploymentHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apierror.BadRequest("Invalid deployment ID"))
		return
	}

	var req model.UpdateDeploymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apierror.BadRequest(err.Error()))
		return
	}

	deployment, err := h.svc.Update(c.Request.Context(), id, &req)
	if err != nil {
		if err == service.ErrDeploymentNotFound {
			response.Error(c, apierror.NotFound("Deployment not found"))
			return
		}
		response.Error(c, apierror.Internal("Failed to update deployment"))
		return
	}

	response.Success(c, deployment)
}

// Delete godoc
// @Summary      Delete deployment
// @Description  Delete a deployment by ID
// @Tags         deployments
// @Param        id  path  string  true  "Deployment ID"
// @Success      204
// @Failure      400  {object}  response.ErrorResponse
// @Failure      404  {object}  response.ErrorResponse
// @Router       /deployments/{id} [delete]
func (h *DeploymentHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apierror.BadRequest("Invalid deployment ID"))
		return
	}

	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		if err == service.ErrDeploymentNotFound {
			response.Error(c, apierror.NotFound("Deployment not found"))
			return
		}
		response.Error(c, apierror.Internal("Failed to delete deployment"))
		return
	}

	response.NoContent(c)
}

// ListByService godoc
// @Summary      List deployments
// @Description  List deployments for a service with pagination
// @Tags         deployments
// @Produce      json
// @Param        id         path   string  true   "Service ID"
// @Param        page       query  int     false  "Page number"
// @Param        page_size  query  int     false  "Items per page"
// @Success      200  {object}  response.PaginatedResponse
// @Failure      400  {object}  response.ErrorResponse
// @Failure      404  {object}  response.ErrorResponse
// @Router       /services/{id}/deployments [get]
func (h *DeploymentHandler) ListByService(c *gin.Context) {
	serviceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apierror.BadRequest("Invalid service ID"))
		return
	}

	var req model.ListDeploymentsRequest
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

	deployments, total, err := h.svc.ListByService(c.Request.Context(), serviceID, &req)
	if err != nil {
		if err == service.ErrServiceNotFound {
			response.Error(c, apierror.NotFound("Service not found"))
			return
		}
		response.Error(c, apierror.Internal("Failed to fetch deployments"))
		return
	}

	meta := pagination.NewMeta(req.Page, req.PageSize, total)
	response.Paginated(c, deployments, &response.PaginationMeta{
		Page:       meta.Page,
		PageSize:   meta.PageSize,
		TotalItems: meta.TotalItems,
		TotalPages: meta.TotalPages,
	})
}
