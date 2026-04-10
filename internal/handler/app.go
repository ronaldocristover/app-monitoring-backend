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

type AppHandler struct {
	appSvc service.AppService
}

func NewAppHandler(appSvc service.AppService) *AppHandler {
	return &AppHandler{appSvc: appSvc}
}

// CreateApp godoc
// @Summary      Create an app
// @Description  Create a new application
// @Tags         apps
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      model.CreateAppRequest  true  "App data"
// @Success      201  {object}  response.SuccessResponse{data=model.App}
// @Failure      400  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /apps [post]
func (h *AppHandler) Create(c *gin.Context) {
	var req model.CreateAppRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apierror.BadRequest(err.Error()))
		return
	}

	app, err := h.appSvc.Create(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, apierror.Internal("Failed to create app"))
		return
	}

	response.Created(c, app)
}

// ListApps godoc
// @Summary      List apps
// @Description  Get paginated list of apps with optional search and tag filters
// @Tags         apps
// @Produce      json
// @Security     BearerAuth
// @Param        page       query  int     false  "Page number"    minimum(1)  default(1)
// @Param        page_size  query  int     false  "Items per page" minimum(1)  maximum(100)  default(20)
// @Param        search     query  string  false  "Search by app name"
// @Param        tags       query  string  false  "Filter by tags"
// @Success      200  {object}  response.PaginatedResponse{data=[]model.App}
// @Failure      400  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /apps [get]
func (h *AppHandler) List(c *gin.Context) {
	var req model.ListAppsRequest
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

	apps, total, err := h.appSvc.List(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, apierror.Internal("Failed to fetch apps"))
		return
	}

	meta := pagination.NewMeta(req.Page, req.PageSize, total)
	response.Paginated(c, apps, &response.PaginationMeta{
		Page:       meta.Page,
		PageSize:   meta.PageSize,
		TotalItems: meta.TotalItems,
		TotalPages: meta.TotalPages,
	})
}

// GetApp godoc
// @Summary      Get an app by ID
// @Description  Get a single app by UUID (without nested resources)
// @Tags         apps
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "App ID (UUID)"
// @Success      200  {object}  response.SuccessResponse{data=model.App}
// @Failure      400  {object}  response.ErrorResponse
// @Failure      404  {object}  response.ErrorResponse
// @Router       /apps/{id} [get]
func (h *AppHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apierror.BadRequest("Invalid app ID"))
		return
	}

	app, err := h.appSvc.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, apierror.NotFound("App not found"))
		return
	}

	response.Success(c, app)
}

// GetAppDetail godoc
// @Summary      Get app detail
// @Description  Get a single app with environments and services
// @Tags         apps
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "App ID (UUID)"
// @Success      200  {object}  response.SuccessResponse{data=model.App}
// @Failure      400  {object}  response.ErrorResponse
// @Failure      404  {object}  response.ErrorResponse
// @Router       /apps/{id}/detail [get]
func (h *AppHandler) GetDetail(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apierror.BadRequest("Invalid app ID"))
		return
	}

	app, err := h.appSvc.GetByIDFull(c.Request.Context(), id)
	if err != nil {
		response.Error(c, apierror.NotFound("App not found"))
		return
	}

	response.Success(c, app)
}

// UpdateApp godoc
// @Summary      Update an app
// @Description  Update app fields (name, description, tags)
// @Tags         apps
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      string                true  "App ID (UUID)"
// @Param        request  body      model.UpdateAppRequest  true  "Update fields"
// @Success      200  {object}  response.SuccessResponse{data=model.App}
// @Failure      400  {object}  response.ErrorResponse
// @Failure      404  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /apps/{id} [put]
func (h *AppHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apierror.BadRequest("Invalid app ID"))
		return
	}

	var req model.UpdateAppRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apierror.BadRequest(err.Error()))
		return
	}

	app, err := h.appSvc.Update(c.Request.Context(), id, &req)
	if err != nil {
		switch err {
		case service.ErrAppNotFound:
			response.Error(c, apierror.NotFound("App not found"))
		default:
			response.Error(c, apierror.Internal("Failed to update app"))
		}
		return
	}

	response.Success(c, app)
}

// DeleteApp godoc
// @Summary      Delete an app
// @Description  Delete an app by ID
// @Tags         apps
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "App ID (UUID)"
// @Success      200  {object}  response.SuccessResponse
// @Failure      400  {object}  response.ErrorResponse
// @Failure      404  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /apps/{id} [delete]
func (h *AppHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apierror.BadRequest("Invalid app ID"))
		return
	}

	if err := h.appSvc.Delete(c.Request.Context(), id); err != nil {
		switch err {
		case service.ErrAppNotFound:
			response.Error(c, apierror.NotFound("App not found"))
		default:
			response.Error(c, apierror.Internal("Failed to delete app"))
		}
		return
	}

	response.NoContent(c)
}

// CreateFullApp godoc
// @Summary      Create app with nested resources
// @Description  Create a full app with environments and services in a single request
// @Tags         apps
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      model.CreateFullAppRequest  true  "Full app data"
// @Success      201  {object}  response.SuccessResponse{data=model.App}
// @Failure      400  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /apps/full [post]
func (h *AppHandler) CreateFull(c *gin.Context) {
	var req model.CreateFullAppRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apierror.BadRequest(err.Error()))
		return
	}

	app, err := h.appSvc.CreateFull(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, apierror.Internal("Failed to create app with nested resources"))
		return
	}

	response.Created(c, app)
}

// UpdateFullApp godoc
// @Summary      Update app with nested resources
// @Description  Update a full app with environments and services in a single request
// @Tags         apps
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      string                      true  "App ID (UUID)"
// @Param        request  body      model.UpdateFullAppRequest  true  "Full app update data"
// @Success      200  {object}  response.SuccessResponse{data=model.App}
// @Failure      400  {object}  response.ErrorResponse
// @Failure      404  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /apps/{id}/full [put]
func (h *AppHandler) UpdateFull(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apierror.BadRequest("Invalid app ID"))
		return
	}

	var req model.UpdateFullAppRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apierror.BadRequest(err.Error()))
		return
	}

	app, err := h.appSvc.UpdateFull(c.Request.Context(), id, &req)
	if err != nil {
		switch err {
		case service.ErrAppNotFound:
			response.Error(c, apierror.NotFound("App not found"))
		default:
			response.Error(c, apierror.Internal("Failed to update app with nested resources"))
		}
		return
	}

	response.Success(c, app)
}
