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

type EnvironmentHandler struct {
	envSvc service.EnvironmentService
}

func NewEnvironmentHandler(envSvc service.EnvironmentService) *EnvironmentHandler {
	return &EnvironmentHandler{envSvc: envSvc}
}

// CreateEnvironment godoc
// @Summary      Create an environment
// @Description  Create a new environment for an app
// @Tags         environments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      model.CreateEnvironmentRequest  true  "Environment data"
// @Success      201  {object}  response.SuccessResponse{data=model.Environment}
// @Failure      400  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /environments [post]
func (h *EnvironmentHandler) Create(c *gin.Context) {
	var req model.CreateEnvironmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apierror.BadRequest(err.Error()))
		return
	}

	env, err := h.envSvc.Create(c.Request.Context(), &req)
	if err != nil {
		switch err {
		case service.ErrInvalidAppID:
			response.Error(c, apierror.BadRequest("Invalid app ID"))
		default:
			response.Error(c, apierror.Internal("Failed to create environment"))
		}
		return
	}

	response.Created(c, env)
}

// GetEnvironment godoc
// @Summary      Get an environment by ID
// @Description  Get a single environment by UUID
// @Tags         environments
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Environment ID (UUID)"
// @Success      200  {object}  response.SuccessResponse{data=model.Environment}
// @Failure      400  {object}  response.ErrorResponse
// @Failure      404  {object}  response.ErrorResponse
// @Router       /environments/{id} [get]
func (h *EnvironmentHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apierror.BadRequest("Invalid environment ID"))
		return
	}

	env, err := h.envSvc.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, apierror.NotFound("Environment not found"))
		return
	}

	response.Success(c, env)
}

// UpdateEnvironment godoc
// @Summary      Update an environment
// @Description  Update environment fields (name)
// @Tags         environments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      string                          true  "Environment ID (UUID)"
// @Param        request  body      model.UpdateEnvironmentRequest  true  "Update fields"
// @Success      200  {object}  response.SuccessResponse{data=model.Environment}
// @Failure      400  {object}  response.ErrorResponse
// @Failure      404  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /environments/{id} [put]
func (h *EnvironmentHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apierror.BadRequest("Invalid environment ID"))
		return
	}

	var req model.UpdateEnvironmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apierror.BadRequest(err.Error()))
		return
	}

	env, err := h.envSvc.Update(c.Request.Context(), id, &req)
	if err != nil {
		switch err {
		case service.ErrEnvironmentNotFound:
			response.Error(c, apierror.NotFound("Environment not found"))
		default:
			response.Error(c, apierror.Internal("Failed to update environment"))
		}
		return
	}

	response.Success(c, env)
}

// DeleteEnvironment godoc
// @Summary      Delete an environment
// @Description  Delete an environment by ID
// @Tags         environments
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Environment ID (UUID)"
// @Success      200  {object}  response.SuccessResponse
// @Failure      400  {object}  response.ErrorResponse
// @Failure      404  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /environments/{id} [delete]
func (h *EnvironmentHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apierror.BadRequest("Invalid environment ID"))
		return
	}

	if err := h.envSvc.Delete(c.Request.Context(), id); err != nil {
		switch err {
		case service.ErrEnvironmentNotFound:
			response.Error(c, apierror.NotFound("Environment not found"))
		default:
			response.Error(c, apierror.Internal("Failed to delete environment"))
		}
		return
	}

	response.NoContent(c)
}

// ListEnvironments godoc
// @Summary      List environments
// @Description  Get paginated list of environments with optional app filter
// @Tags         environments
// @Produce      json
// @Security     BearerAuth
// @Param        page       query  int     false  "Page number"    minimum(1)  default(1)
// @Param        page_size  query  int     false  "Items per page" minimum(1)  maximum(100)  default(20)
// @Param        app_id     query  string  false  "Filter by app ID (UUID)"
// @Success      200  {object}  response.PaginatedResponse{data=[]model.Environment}
// @Failure      400  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /environments [get]
func (h *EnvironmentHandler) List(c *gin.Context) {
	var req model.ListEnvironmentsRequest
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

	envs, total, err := h.envSvc.List(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, apierror.Internal("Failed to fetch environments"))
		return
	}

	meta := pagination.NewMeta(req.Page, req.PageSize, total)
	response.Paginated(c, envs, &response.PaginationMeta{
		Page:       meta.Page,
		PageSize:   meta.PageSize,
		TotalItems: meta.TotalItems,
		TotalPages: meta.TotalPages,
	})
}
