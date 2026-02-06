package handlers

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/falasefemi2/companyflowlow/dto"
	"github.com/falasefemi2/companyflowlow/services"
	"github.com/falasefemi2/companyflowlow/utils"
)

type RoleHandler struct {
	roleService services.IRoleService
}

func NewRoleHandler(roleService services.IRoleService) *RoleHandler {
	return &RoleHandler{roleService: roleService}
}

func (h *RoleHandler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/companies/{company_id}/roles", h.CreateRole).Methods(http.MethodPost)
	r.HandleFunc("/companies/{company_id}/roles", h.GetRoleList).Methods(http.MethodGet)
	r.HandleFunc("/roles/{id}", h.GetRoleByID).Methods(http.MethodGet)
	r.HandleFunc("/roles/{id}", h.UpdateRole).Methods(http.MethodPut, http.MethodPatch)
	r.HandleFunc("/roles/{id}", h.DeleteRole).Methods(http.MethodDelete)
}

// CreateRole godoc
// @Summary Create role
// @Description Create a new role.
// @Tags roles
// @Accept json
// @Produce json
// @Param company_id path string true "Company ID"
// @Param request body dto.CreateRoleRequest true "Create role payload"
// @Security BearerAuth
// @Success 201 {object} utils.APIResponse{data=dto.RoleResponse}
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /companies/{company_id}/roles [post]
func (h *RoleHandler) CreateRole(w http.ResponseWriter, r *http.Request) {
	claims, err := authorizeEmployeeToken(r, roleSuperAdmin, roleHRManager)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	companyID, err := parseEmployeeCompanyID(r, claims)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid company_id")
		return
	}

	var req dto.CreateRoleRequest
	if err := utils.DecodeJSONBody(r, &req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.roleService.CreateRole(r.Context(), companyID, &req)
	if err != nil {
		var vErr *utils.ValidationError
		if errors.As(err, &vErr) {
			utils.RespondWithError(w, http.StatusBadRequest, vErr.Message)
			return
		}
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusCreated, utils.APIResponse{
		Success: true,
		Data:    result,
	})
}

// GetRoleByID godoc
// @Summary Get role by ID
// @Description Retrieve a role by ID.
// @Tags roles
// @Produce json
// @Param id path string true "Role ID"
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse{data=dto.RoleResponse}
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /roles/{id} [get]
func (h *RoleHandler) GetRoleByID(w http.ResponseWriter, r *http.Request) {
	if _, err := authorizeEmployeeToken(r, roleSuperAdmin, roleHRManager); err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	idStr := mux.Vars(r)["id"]
	roleID, err := uuid.Parse(idStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid role id")
		return
	}

	result, err := h.roleService.GetRoleByID(r.Context(), roleID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Data:    result,
	})
}

// GetRoleList godoc
// @Summary List roles
// @Description Get paginated list of roles.
// @Tags roles
// @Produce json
// @Param company_id path string true "Company ID"
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Param search query string false "Search by name"
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse{data=PaginatedRoleResponse}
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /companies/{company_id}/roles [get]
func (h *RoleHandler) GetRoleList(w http.ResponseWriter, r *http.Request) {
	claims, err := authorizeEmployeeToken(r, roleSuperAdmin, roleHRManager)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	companyID, err := parseEmployeeCompanyID(r, claims)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid company_id")
		return
	}

	page, err := parsePositiveInt(r.URL.Query().Get("page"), 1)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid page")
		return
	}

	pageSize, err := parsePositiveInt(r.URL.Query().Get("page_size"), 10)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid page_size")
		return
	}

	req := &dto.RoleListRequest{
		Page:     page,
		PageSize: pageSize,
		Search:   r.URL.Query().Get("search"),
	}

	result, err := h.roleService.GetRoleList(r.Context(), companyID, req)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Data:    result,
	})
}

// UpdateRole godoc
// @Summary Update role
// @Description Update a role by ID.
// @Tags roles
// @Accept json
// @Produce json
// @Param id path string true "Role ID"
// @Param request body dto.UpdateRoleRequest true "Update role payload"
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse{data=dto.RoleResponse}
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /roles/{id} [put]
// @Router /roles/{id} [patch]
func (h *RoleHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	if _, err := authorizeEmployeeToken(r, roleSuperAdmin, roleHRManager); err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	idStr := mux.Vars(r)["id"]
	roleID, err := uuid.Parse(idStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid role id")
		return
	}

	var req dto.UpdateRoleRequest
	if err := utils.DecodeJSONBody(r, &req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.roleService.UpdateRole(r.Context(), roleID, &req)
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

// DeleteRole godoc
// @Summary Delete role
// @Description Delete a role by ID.
// @Tags roles
// @Produce json
// @Param id path string true "Role ID"
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse{message=string}
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /roles/{id} [delete]
func (h *RoleHandler) DeleteRole(w http.ResponseWriter, r *http.Request) {
	if _, err := authorizeEmployeeToken(r, roleSuperAdmin, roleHRManager); err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	idStr := mux.Vars(r)["id"]
	roleID, err := uuid.Parse(idStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid role id")
		return
	}

	if err := h.roleService.DeleteRole(r.Context(), roleID); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "role deleted",
	})
}
