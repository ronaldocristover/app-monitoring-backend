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

type ServerHandler struct {
	serverSvc service.ServerService
}

func NewServerHandler(serverSvc service.ServerService) *ServerHandler {
	return &ServerHandler{serverSvc: serverSvc}
}

// CreateServer godoc
// @Summary      Create a server
// @Description  Create a new server
// @Tags         servers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      model.CreateServerRequest  true  "Server data"
// @Success      201  {object}  response.SuccessResponse{data=model.Server}
// @Failure      400  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /servers [post]
func (h *ServerHandler) Create(c *gin.Context) {
	var req model.CreateServerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apierror.BadRequest(err.Error()))
		return
	}

	server, err := h.serverSvc.Create(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, apierror.Internal("Failed to create server"))
		return
	}

	response.Created(c, server)
}

// GetServer godoc
// @Summary      Get a server by ID
// @Description  Get a single server by UUID
// @Tags         servers
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Server ID (UUID)"
// @Success      200  {object}  response.SuccessResponse{data=model.Server}
// @Failure      400  {object}  response.ErrorResponse
// @Failure      404  {object}  response.ErrorResponse
// @Router       /servers/{id} [get]
func (h *ServerHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apierror.BadRequest("Invalid server ID"))
		return
	}

	server, err := h.serverSvc.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, apierror.NotFound("Server not found"))
		return
	}

	response.Success(c, server)
}

// UpdateServer godoc
// @Summary      Update a server
// @Description  Update server fields (name, ip, provider)
// @Tags         servers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      string                      true  "Server ID (UUID)"
// @Param        request  body      model.UpdateServerRequest  true  "Update fields"
// @Success      200  {object}  response.SuccessResponse{data=model.Server}
// @Failure      400  {object}  response.ErrorResponse
// @Failure      404  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /servers/{id} [put]
func (h *ServerHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apierror.BadRequest("Invalid server ID"))
		return
	}

	var req model.UpdateServerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apierror.BadRequest(err.Error()))
		return
	}

	server, err := h.serverSvc.Update(c.Request.Context(), id, &req)
	if err != nil {
		switch err {
		case service.ErrServerNotFound:
			response.Error(c, apierror.NotFound("Server not found"))
		default:
			response.Error(c, apierror.Internal("Failed to update server"))
		}
		return
	}

	response.Success(c, server)
}

// DeleteServer godoc
// @Summary      Delete a server
// @Description  Delete a server by ID
// @Tags         servers
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Server ID (UUID)"
// @Success      200  {object}  response.SuccessResponse
// @Failure      400  {object}  response.ErrorResponse
// @Failure      404  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /servers/{id} [delete]
func (h *ServerHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apierror.BadRequest("Invalid server ID"))
		return
	}

	if err := h.serverSvc.Delete(c.Request.Context(), id); err != nil {
		switch err {
		case service.ErrServerNotFound:
			response.Error(c, apierror.NotFound("Server not found"))
		default:
			response.Error(c, apierror.Internal("Failed to delete server"))
		}
		return
	}

	response.NoContent(c)
}

// ListServers godoc
// @Summary      List servers
// @Description  Get paginated list of servers with optional search filter
// @Tags         servers
// @Produce      json
// @Security     BearerAuth
// @Param        page       query  int     false  "Page number"    minimum(1)  default(1)
// @Param        page_size  query  int     false  "Items per page" minimum(1)  maximum(100)  default(20)
// @Param        search     query  string  false  "Search by name or IP"
// @Success      200  {object}  response.PaginatedResponse{data=[]model.Server}
// @Failure      400  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /servers [get]
func (h *ServerHandler) List(c *gin.Context) {
	var req model.ListServersRequest
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

	servers, total, err := h.serverSvc.List(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, apierror.Internal("Failed to fetch servers"))
		return
	}

	meta := pagination.NewMeta(req.Page, req.PageSize, total)
	response.Paginated(c, servers, &response.PaginationMeta{
		Page:       meta.Page,
		PageSize:   meta.PageSize,
		TotalItems: meta.TotalItems,
		TotalPages: meta.TotalPages,
	})
}
