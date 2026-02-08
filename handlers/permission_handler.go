package handlers

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/falasefemi2/companyflowlow/dto"
	"github.com/falasefemi2/companyflowlow/models"
	"github.com/falasefemi2/companyflowlow/services"
	"github.com/falasefemi2/companyflowlow/utils"
)

type PermissionHandler struct {
	permissionService services.IPermissionService
}

func NewPermissionHandler(permissionService services.IPermissionService) *PermissionHandler {
	return &PermissionHandler{permissionService: permissionService}
}

func (h *PermissionHandler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/roles/{role_id}/permissions", h.SetRolePermissions).Methods(http.MethodPost)
	r.HandleFunc("/roles/{role_id}/permissions", h.GetRolePermissions).Methods(http.MethodGet)
	r.HandleFunc("/roles/{role_id}/permissions/{permission_id}", h.AddPermissionToRole).Methods(http.MethodPut)
	r.HandleFunc("/roles/{role_id}/permissions/{permission_id}", h.RemovePermissionFromRole).Methods(http.MethodDelete)
}

// SetRolePermissions godoc
// @Summary Set role permissions
// @Description Replace all permissions for a role.
// @Tags permissions
// @Accept json
// @Produce json
// @Param role_id path string true "Role ID"
// @Param request body []dto.CreatePermissionRequest true "Permissions payload"
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse{data=dto.RoleResponse}
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /roles/{role_id}/permissions [post]
func (h *PermissionHandler) SetRolePermissions(w http.ResponseWriter, r *http.Request) {
	if _, err := authorizeEmployeeToken(r, roleSuperAdmin, roleHRManager); err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	roleIDStr := mux.Vars(r)["role_id"]
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid role_id")
		return
	}

	var reqs []dto.CreatePermissionRequest
	if err := utils.DecodeJSONBody(r, &reqs); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	permissions := make([]*models.Permission, len(reqs))
	for i, req := range reqs {
		permissions[i] = &models.Permission{
			Action:     req.Action,
			Resource:   req.Resource,
			Conditions: req.Conditions,
		}
	}

	result, err := h.permissionService.SetRolePermissions(r.Context(), roleID, permissions)
	if err != nil {
		var vErr *utils.ValidationError
		if errors.As(err, &vErr) {
			utils.RespondWithError(w, http.StatusBadRequest, vErr.Message)
			return
		}
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Data:    result,
	})
}

// GetRolePermissions godoc
// @Summary Get role permissions
// @Description Retrieve all permissions for a role.
// @Tags permissions
// @Produce json
// @Param role_id path string true "Role ID"
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse{data=[]dto.PermissionResponse}
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /roles/{role_id}/permissions [get]
func (h *PermissionHandler) GetRolePermissions(w http.ResponseWriter, r *http.Request) {
	if _, err := authorizeEmployeeToken(r, roleSuperAdmin, roleHRManager); err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	roleIDStr := mux.Vars(r)["role_id"]
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid role_id")
		return
	}

	result, err := h.permissionService.GetRolePermissions(r.Context(), roleID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Data:    result,
	})
}

// AddPermissionToRole godoc
// @Summary Add permission to role
// @Description Add a single permission to a role.
// @Tags permissions
// @Accept json
// @Produce json
// @Param role_id path string true "Role ID"
// @Param request body dto.CreatePermissionRequest true "Permission payload"
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse{data=dto.RoleResponse}
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /roles/{role_id}/permissions/{permission_id} [put]
func (h *PermissionHandler) AddPermissionToRole(w http.ResponseWriter, r *http.Request) {
	if _, err := authorizeEmployeeToken(r, roleSuperAdmin, roleHRManager); err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	roleIDStr := mux.Vars(r)["role_id"]
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid role_id")
		return
	}

	var req dto.CreatePermissionRequest
	if err := utils.DecodeJSONBody(r, &req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	permission := &models.Permission{
		Action:     req.Action,
		Resource:   req.Resource,
		Conditions: req.Conditions,
	}

	result, err := h.permissionService.AddPermissionToRole(r.Context(), roleID, permission)
	if err != nil {
		var vErr *utils.ValidationError
		if errors.As(err, &vErr) {
			utils.RespondWithError(w, http.StatusBadRequest, vErr.Message)
			return
		}
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Data:    result,
	})
}

// RemovePermissionFromRole godoc
// @Summary Remove permission from role
// @Description Remove a permission from a role.
// @Tags permissions
// @Produce json
// @Param role_id path string true "Role ID"
// @Param permission_id path string true "Permission ID"
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse{data=dto.RoleResponse}
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /roles/{role_id}/permissions/{permission_id} [delete]
func (h *PermissionHandler) RemovePermissionFromRole(w http.ResponseWriter, r *http.Request) {
	if _, err := authorizeEmployeeToken(r, roleSuperAdmin, roleHRManager); err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	roleIDStr := mux.Vars(r)["role_id"]
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid role_id")
		return
	}

	permissionIDStr := mux.Vars(r)["permission_id"]
	permissionID, err := uuid.Parse(permissionIDStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid permission_id")
		return
	}

	result, err := h.permissionService.RemovePermissionFromRole(r.Context(), roleID, permissionID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Data:    result,
	})
}
