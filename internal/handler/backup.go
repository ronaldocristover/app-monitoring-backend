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

type BackupHandler struct {
	svc service.BackupService
}

func NewBackupHandler(svc service.BackupService) *BackupHandler {
	return &BackupHandler{svc: svc}
}

// Create godoc
// @Summary      Create backup
// @Description  Create a new backup configuration for a service
// @Tags         backups
// @Accept       json
// @Produce      json
// @Param        id       path      string                    true  "Service ID"
// @Param        request  body      model.CreateBackupRequest  true  "Backup data"
// @Success      201      {object}  response.SuccessResponse
// @Failure      400      {object}  response.ErrorResponse
// @Failure      404      {object}  response.ErrorResponse
// @Failure      500      {object}  response.ErrorResponse
// @Router       /services/{id}/backups [post]
func (h *BackupHandler) Create(c *gin.Context) {
	serviceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apierror.BadRequest("Invalid service ID"))
		return
	}

	var req model.CreateBackupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apierror.BadRequest(err.Error()))
		return
	}
	req.ServiceID = serviceID

	backup, err := h.svc.Create(c.Request.Context(), &req)
	if err != nil {
		if err == service.ErrServiceNotFound {
			response.Error(c, apierror.NotFound("Service not found"))
			return
		}
		response.Error(c, apierror.Internal("Failed to create backup"))
		return
	}

	response.Created(c, backup)
}

// Get godoc
// @Summary      Get backup
// @Description  Get backup by ID
// @Tags         backups
// @Produce      json
// @Param        id   path      string  true  "Backup ID"
// @Success      200  {object}  response.SuccessResponse
// @Failure      400  {object}  response.ErrorResponse
// @Failure      404  {object}  response.ErrorResponse
// @Router       /backups/{id} [get]
func (h *BackupHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apierror.BadRequest("Invalid backup ID"))
		return
	}

	backup, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, apierror.NotFound("Backup not found"))
		return
	}

	response.Success(c, backup)
}

// Update godoc
// @Summary      Update backup
// @Description  Update an existing backup configuration
// @Tags         backups
// @Accept       json
// @Produce      json
// @Param        id       path      string                     true  "Backup ID"
// @Param        request  body      model.UpdateBackupRequest  true  "Backup data"
// @Success      200      {object}  response.SuccessResponse
// @Failure      400      {object}  response.ErrorResponse
// @Failure      404      {object}  response.ErrorResponse
// @Router       /backups/{id} [put]
func (h *BackupHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apierror.BadRequest("Invalid backup ID"))
		return
	}

	var req model.UpdateBackupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apierror.BadRequest(err.Error()))
		return
	}

	backup, err := h.svc.Update(c.Request.Context(), id, &req)
	if err != nil {
		if err == service.ErrBackupNotFound {
			response.Error(c, apierror.NotFound("Backup not found"))
			return
		}
		response.Error(c, apierror.Internal("Failed to update backup"))
		return
	}

	response.Success(c, backup)
}

// Delete godoc
// @Summary      Delete backup
// @Description  Delete a backup by ID
// @Tags         backups
// @Param        id  path  string  true  "Backup ID"
// @Success      204
// @Failure      400  {object}  response.ErrorResponse
// @Failure      404  {object}  response.ErrorResponse
// @Router       /backups/{id} [delete]
func (h *BackupHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apierror.BadRequest("Invalid backup ID"))
		return
	}

	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		if err == service.ErrBackupNotFound {
			response.Error(c, apierror.NotFound("Backup not found"))
			return
		}
		response.Error(c, apierror.Internal("Failed to delete backup"))
		return
	}

	response.NoContent(c)
}

// ListByService godoc
// @Summary      List backups
// @Description  List backups for a service with pagination
// @Tags         backups
// @Produce      json
// @Param        id         path   string  true   "Service ID"
// @Param        page       query  int     false  "Page number"
// @Param        page_size  query  int     false  "Items per page"
// @Param        status     query  string  false  "Filter by status"
// @Success      200  {object}  response.PaginatedResponse
// @Failure      400  {object}  response.ErrorResponse
// @Failure      404  {object}  response.ErrorResponse
// @Router       /services/{id}/backups [get]
func (h *BackupHandler) ListByService(c *gin.Context) {
	serviceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apierror.BadRequest("Invalid service ID"))
		return
	}

	var req model.ListBackupsRequest
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

	backups, total, err := h.svc.ListByService(c.Request.Context(), serviceID, &req)
	if err != nil {
		if err == service.ErrServiceNotFound {
			response.Error(c, apierror.NotFound("Service not found"))
			return
		}
		response.Error(c, apierror.Internal("Failed to fetch backups"))
		return
	}

	meta := pagination.NewMeta(req.Page, req.PageSize, total)
	response.Paginated(c, backups, &response.PaginationMeta{
		Page:       meta.Page,
		PageSize:   meta.PageSize,
		TotalItems: meta.TotalItems,
		TotalPages: meta.TotalPages,
	})
}
